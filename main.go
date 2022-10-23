package main

import (
	"context"
	"log"
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

	dbConn, err := db.InitPostgresDb(ctx)
	if err != nil {
		log.Fatalf("cannot start pg connection: %s", err)
	}

	repo := data.NewPgRepo(dbConn)

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)
	deathsCsvUrl := env.EnvOrDefault("DEATHS_PER_MUNICIPALITY_CSV_URL", deathsPerMunicipalityCsvUrl)

	dataManager, err := data.NewService(repo, casesCsvUrl, timelineCsvUrl, deathsCsvUrl, false)
	if err != nil {
		log.Fatalf("cannot init data manager: %s", err)
	}

	if env.BoolEnvOrDefault("POPULATE_DB", false) {
		ticker := time.NewTicker(24 * time.Hour) // every day
		go func() {
			for ; true; <-ticker.C {
				if err := dataManager.PopulateEverything(ctx); err != nil {
					log.Printf("ERROR: database population failed: %s", err)
				}
			}
		}()
	}

	app := api.NewApi(repo)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}

	// todo add graceful stop
}
