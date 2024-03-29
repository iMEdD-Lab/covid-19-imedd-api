package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"covid19-greece-api/internal/api"
	"covid19-greece-api/internal/data"
	"covid19-greece-api/pkg/db"
	"covid19-greece-api/pkg/env"
)

const (
	casesCsvDefaultUrl          = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greece_cases_v2.csv`
	timelineDefaultCsvUrl       = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greeceTimeline.csv`
	deathsPerMunicipalityCsvUrl = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/deaths%20covid%20greece%20municipality%2020%2021.csv`
	demographicsUrl             = `https://raw.githubusercontent.com/Sandbird/covid19-Greece/master/demography_total_details.csv`
	wasteUrl                    = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/viral_waste_water.csv`
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// sleep a bit to avoid race condition with postgres docker start
	time.Sleep(3 * time.Second)

	// initialize postgres database connection and repository
	dbConn, err := db.InitPostgresDb(ctx)
	if err != nil {
		log.Fatalf("cannot start pg connection: %s", err)
	}
	repo, err := data.NewPgRepo(dbConn, os.Getenv("YPES_MUNICIPALITIES_CSV_FILE"))
	if err != nil {
		log.Fatalf("cannot initialize data repository: %s", err)
	}

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)
	deathsCsvUrl := env.EnvOrDefault("DEATHS_PER_MUNICIPALITY_CSV_URL", deathsPerMunicipalityCsvUrl)
	demographicsCsvUrl := env.EnvOrDefault("DEMOGRAPHICS_CSV_URL", demographicsUrl)
	wasteCsvUrl := env.EnvOrDefault("WASTE_CSV_URL", wasteUrl)

	// initialize data manager for database population
	dataManager, err := data.NewService(
		repo,
		casesCsvUrl,
		timelineCsvUrl,
		deathsCsvUrl,
		demographicsCsvUrl,
		wasteCsvUrl,
		false,
	)
	if err != nil {
		log.Fatalf("cannot init data manager: %s", err)
	}

	// populate database with new data at startup and every 24 hours
	if env.BoolEnvOrDefault("POPULATE_DB", true) {
		ticker := time.NewTicker(24 * time.Hour)
		go func() {
			for ; true; <-ticker.C {
				if err := dataManager.PopulateEverything(ctx); err != nil {
					log.Printf("ERROR: database population failed: %s", err)
				}
			}
		}()
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	token := os.Getenv("SECRET_TOKEN")
	if len(token) < 10 {
		log.Fatalf("SECRET_TOKEN too short. Please give a safe secret token")
	}
	app := api.NewApi(repo, dataManager, token)

	port := env.IntEnvOrDefault("PORT", 8080)
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: app.Router,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		log.Printf("received termination signal")

		// if graceful shutdown lasts longer than 30sec, kill it
		shutdownCtx, sdCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer sdCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("error while shutting down: %s", err)
		}
		serverStopCtx()
	}()

	// Run the server
	log.Printf("Starting COVID-19 API (port %d)...\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error serving: %s", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	log.Printf("server was gracefully stopped. Bye!")
}
