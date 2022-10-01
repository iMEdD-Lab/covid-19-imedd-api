package main

import (
	"context"
	"log"
	"time"

	"covid19-greece-api/internal/data"
	"covid19-greece-api/pkg/db"
	"covid19-greece-api/pkg/env"
)

const (
	casesCsvDefaultUrl    = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greece_cases_v2.csv`
	timelineDefaultCsvUrl = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greeceTimeline.csv`
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConn, err := db.InitPostgresDb(ctx)
	if err != nil {
		log.Fatalf("cannot start pg connection: %s", err)
	}

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)

	dataManager, err := data.NewManager(dbConn, casesCsvUrl, timelineCsvUrl)
	if err != nil {
		log.Fatalf("cannot init data manager: %s", err)
	}

	ticker := time.NewTicker(24 * time.Hour) // every day
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := dataManager.PopulateEverything(ctx); err != nil {
					log.Printf("ERROR: database population failed: %s", err)
				}
			}
		}
	}()

	select {}
}
