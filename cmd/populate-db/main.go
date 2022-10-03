package main

import (
	"context"
	"flag"
	"log"
	"time"

	"covid19-greece-api/internal/data"
	"covid19-greece-api/pkg/db"
	"covid19-greece-api/pkg/env"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	casesCsvDefaultUrl    = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greece_cases_v2.csv`
	timelineDefaultCsvUrl = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greeceTimeline.csv`
)

func main() {
	var skipGeo, skipCases, skipTimeline bool
	flag.BoolVar(&skipGeo, "skipGeo", false, "skips populating greece_geo_info table")
	flag.BoolVar(&skipCases, "skipCases", false, "skips populating cases_per_prefecture table")
	flag.BoolVar(&skipTimeline, "skipTimeline", false, "skips populating greece_timeline table")
	flag.Parse()

	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConn, err := db.InitPostgresDb(ctx)
	if err != nil {
		log.Fatalf("cannot start pg connection: %s", err)
	}
	repo := data.NewPgRepo(dbConn)

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)

	dataManager, err := data.NewService(repo, casesCsvUrl, timelineCsvUrl, false)
	if err != nil {
		log.Fatalf("cannot init data manager: %s", err)
	}

	if !skipGeo {
		if err := dataManager.PopulateGeo(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if !skipCases {
		if err := dataManager.PopulateCases(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if !skipTimeline {
		if err := dataManager.PopulateTimeline(ctx); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Finished after %v", time.Since(start))
}
