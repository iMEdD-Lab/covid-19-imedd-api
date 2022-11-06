package main

import (
	"context"
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
	repo := data.NewPgRepo(dbConn)

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)
	deathsCsvUrl := env.EnvOrDefault("DEATHS_PER_MUNICIPALITY_CSV_URL", deathsPerMunicipalityCsvUrl)

	// initialize data manager for database population
	dataManager, err := data.NewService(repo, casesCsvUrl, timelineCsvUrl, deathsCsvUrl, false)
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

	app := api.NewApi(repo)

	server := &http.Server{Addr: "0.0.0.0:8080", Handler: app.Router}

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
	log.Println("Starting COVID19 API")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error serving: %s", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	log.Printf("server was gracefully stopped. Bye!")
}
