package main

import (
	"context"
	"flag"
	"log"
	"time"

	"covid-data-transformation/internal/data"
	"covid-data-transformation/pkg/db"
	"covid-data-transformation/pkg/env"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	casesCsvDefaultUrl    = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greece_cases_v2.csv`
	timelineDefaultCsvUrl = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greeceTimeline.csv`
	dateLayout            = "01/02/06"
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

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)

	dataManager, err := data.NewManager(dbConn, casesCsvUrl, timelineCsvUrl)
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
