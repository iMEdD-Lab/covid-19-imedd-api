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
	casesCsvDefaultUrl          = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greece_cases_v2.csv`
	timelineDefaultCsvUrl       = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/greeceTimeline.csv`
	deathsPerMunicipalityCsvUrl = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/deaths%20covid%20greece%20municipality%2020%2021.csv`
	demographicsUrl             = `https://raw.githubusercontent.com/Sandbird/covid19-Greece/master/demography_total_details.csv`
	wasteUrl                    = `https://raw.githubusercontent.com/iMEdD-Lab/open-data/master/COVID-19/viral_waste_water.csv`
)

func main() {
	var skipRegionalUnits, skipCases, skipTimeline, skipDeaths, skipDemographics bool
	flag.BoolVar(&skipRegionalUnits, "skipRegionalUnits", false, "skips populating regional_units table")
	flag.BoolVar(&skipCases, "skipCases", false, "skips populating cases_per_regional_unit table")
	flag.BoolVar(&skipTimeline, "skipTimeline", false, "skips populating greece_timeline table")
	flag.BoolVar(&skipDeaths, "skipDeaths", false, "skips populating deaths_per_municipality table")
	flag.BoolVar(&skipDemographics, "skipDemographics", false, "skips populating demography_per_age table")
	flag.Parse()

	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConn, err := db.InitPostgresDb(ctx)
	if err != nil {
		log.Fatalf("cannot start pg connection: %s", err)
	}
	repo, err := data.NewPgRepo(dbConn, "")
	if err != nil {
		log.Fatalf("cannot initialize data repository: %s", err)
	}

	casesCsvUrl := env.EnvOrDefault("CASES_CSV_URL", casesCsvDefaultUrl)
	timelineCsvUrl := env.EnvOrDefault("TIMELINE_CSV_URL", timelineDefaultCsvUrl)
	deathsCsvUrl := env.EnvOrDefault("DEATHS_PER_MUNICIPALITY_CSV_URL", deathsPerMunicipalityCsvUrl)
	demographicsCsvUrl := env.EnvOrDefault("DEMOGRAPHICS_CSV_URL", demographicsUrl)
	wasteCsvUrl := env.EnvOrDefault("WASTE_CSV_URL", wasteUrl)

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

	if !skipRegionalUnits {
		if err := dataManager.PopulateRegionalUnits(ctx); err != nil {
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

	if !skipDeaths {
		if err := dataManager.PopulateDeathsPerMunicipality(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if !skipDemographics {
		if err := dataManager.PopulateDemographic(ctx); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Finished after %v", time.Since(start))
}
