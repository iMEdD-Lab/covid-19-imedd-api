package data

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gosimple/slug"

	"covid19-greece-api/pkg/file"
	"covid19-greece-api/pkg/vartypes"
)

// todo reduce some logs

const (
	dateLayout       = "01/02/06"
	simpleDateLayout = "2006-01-02"
)

type Service struct {
	repo      Repo
	fromFiles bool

	// CSV files, most probably coming from here (https://github.com/iMEdD-Lab/open-data/).
	// We strictly follow their format.
	casesCsvSrc              string
	timelineCsvSrc           string
	deathsPerMunicipalitySrc string
}

type FullInfo struct {
	Date                   time.Time `json:"date"`
	Cases                  int       `json:"cases"`
	TotalReinfections      int       `json:"total_reinfections"`
	Deaths                 int       `json:"deaths"`
	DeathsCum              int       `json:"deaths_cum"`
	Recovered              int       `json:"recovered"`
	HospitalAdmissions     int       `json:"hospital_admissions"`
	HospitalDischarges     int       `json:"hospital_discharges"`
	Intubated              int       `json:"intubated"`
	IntubatedVac           int       `json:"intubated_vac"`
	IntubatedUnvac         int       `json:"intubated_unvac"`
	IcuOccupancy           float64   `json:"icu_occupancy"`
	BedsOccupancy          float64   `json:"beds_occupancy"`
	EstimatedNewRtpcrTests int       `json:"estimated_new_rtpcr_tests"`
	EstimatedNewRapidTests int       `json:"estimated_new_rapid_tests"`
	EstimatedNewTotalTests int       `json:"estimated_new_total_tests"`
}

type County struct {
	Id               int    `json:"id"`
	Slug             string `json:"slug"`
	Department       string `json:"department"`
	Prefecture       string `json:"prefecture"`
	CountyNormalized string `json:"county_normalized"`
	County           string `json:"county"`
	Pop11            int    `json:"pop_11"`
}

type Municipality struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type YearlyDeaths struct {
	MunId  int `json:"municipality_id"`
	Deaths int `json:"deaths"`
	Year   int `json:"year"`
}

type DemographicInfo struct {
	Date              time.Time `json:"date"`
	Category          string    `json:"category"`
	Cases             int       `json:"cases"`
	Deaths            int       `json:"deaths"`
	Intensive         int       `json:"intensive"`
	Discharged        int       `json:"discharged"`
	Hospitalized      int       `json:"hospitalized"`
	HospitalizedInIcu int       `json:"hospitalized_in_icu"`
	PassedAway        int       `json:"passed_away"`
	Recovered         int       `json:"recovered"`
	TreatedAtHome     int       `json:"treated_at_home"`
}

func NewService(
	repo Repo,
	casesSrc string,
	timelineSrc string,
	deathsPerMunicipalitySrc string,
	fromFiles bool,
) (*Service, error) {
	return &Service{
		repo:                     repo,
		casesCsvSrc:              casesSrc,
		timelineCsvSrc:           timelineSrc,
		deathsPerMunicipalitySrc: deathsPerMunicipalitySrc,
		fromFiles:                fromFiles,
	}, nil
}

func (s *Service) PopulateEverything(ctx context.Context) error {
	start := time.Now()
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.PopulateCounties(ctx); err != nil {
			return fmt.Errorf("error populating geo: %s", err)
		}
		if err := s.PopulateCases(ctx); err != nil {
			return fmt.Errorf("error populating cases per county: %s", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.PopulateTimeline(ctx); err != nil {
			return fmt.Errorf("error populating timeline: %s", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.PopulateDeathsPerMunicipality(ctx); err != nil {
			return fmt.Errorf("error populating municipalities: %s", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error populating db: %s", err)
	}

	log.Printf("database populated successfully after %s", time.Since(start).String())

	return nil
}

func (s *Service) PopulateDeathsPerMunicipality(ctx context.Context) error {
	data, err := file.ReadCsv(s.deathsPerMunicipalitySrc, s.fromFiles)
	if err != nil {
		return fmt.Errorf("error reading csv file: %v", err)
	}
	headers := data[0] // first element is always "municipality", then rest of columns are like deaths_covid_{year}

	var years []int
	for _, h := range headers[1:] {
		// extract year number from header
		parts := strings.Split(h, "_")
		years = append(years, vartypes.StringToInt(parts[len(parts)-1]))
	}

	for _, d := range data[1:] {
		name := d[0]
		id, err := s.repo.AddMunicipality(ctx, name)
		if err != nil {
			return err
		}
		for i, yearlyDeath := range d[1:] {
			// match specific amount of deaths to specific year and municipality_id
			if err := s.repo.AddYearlyDeath(ctx, id, vartypes.StringToInt(yearlyDeath), years[i]); err != nil {
				return err
			}

		}
	}

	log.Printf("added deaths for %d municipalities and years %v", len(data)-1, years)

	return nil
}

func (s *Service) PopulateCounties(ctx context.Context) error {
	data, err := file.ReadCsv(s.casesCsvSrc, s.fromFiles)
	if err != nil {
		log.Fatalf("Error reading csv file: %v", err)
	}

	// take dates from csv first row
	var dateHeaders []time.Time
	headers := data[0]
	for _, header := range headers[5:] {
		t, err := csvHeaderToDate(header)
		if err != nil {
			log.Fatal(err)
		}
		dateHeaders = append(dateHeaders, t)
	}

	for _, row := range data[1:] {
		err := s.repo.AddCounty(ctx, County{
			Slug:             slug.Make(row[2]),
			Department:       row[0],
			Prefecture:       row[1],
			CountyNormalized: row[2],
			County:           row[3],
			Pop11:            vartypes.StringToInt(row[4]),
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("added %d counties", len(data)-1)

	return nil
}

func (s *Service) PopulateCases(ctx context.Context) error {
	data, err := file.ReadCsv(s.casesCsvSrc, s.fromFiles)
	if err != nil {
		log.Fatalf("Error reading csv file: %v", err)
	}

	// take dates from csv first row
	var dateHeaders []time.Time
	headers := data[0]
	for _, header := range headers[5:] {
		t, err := csvHeaderToDate(header)
		if err != nil {
			log.Fatal(err)
		}
		dateHeaders = append(dateHeaders, t)
	}

	// from 12/7 and later, EODY started giving weekly info instead of daily.
	// From this date on, we only take the 1st of 7
	weeklyDates := make(map[string]struct{})
	startWithoutEody := time.Date(2022, 7, 12, 0, 0, 0, 0, time.Local)
	it := startWithoutEody
	for {
		weeklyDates[it.Format(simpleDateLayout)] = struct{}{}
		it = it.Add(7 * 24 * time.Hour)
		if it.After(time.Now()) {
			break
		}
	}

	for _, row := range data[1:] {
		for i, date := range dateHeaders {
			if date.IsZero() {
				log.Fatalf("invalid date for column %d: %v", i, row[i+5])
			}
			amount := vartypes.StringToInt(row[i+5])
			if date.After(startWithoutEody) {
				_, exists := weeklyDates[date.Format(simpleDateLayout)]
				if !exists {
					continue
				}
			}
			sl := slug.Make(row[2])
			if err := s.repo.AddCase(ctx, date, amount, sl); err != nil {
				log.Fatalf("Error adding death day: %v", err)
			}
		}

		log.Printf("added all cases for county %s", row[2])
	}

	return nil
}

func (s *Service) PopulateTimeline(ctx context.Context) error {
	data, err := file.ReadCsv(s.timelineCsvSrc, s.fromFiles)
	if err != nil {
		log.Fatalf("Error reading csv file: %v", err)
	}

	// take dates from csv first row
	var dateHeaders []time.Time
	headers := data[0]
	for _, header := range headers[3:] {
		t, err := csvHeaderToDate(header)
		if err != nil {
			log.Fatal(err)
		}
		dateHeaders = append(dateHeaders, t)
	}

	tl := make(map[string]*FullInfo)

	for index, row := range data {
		for i, date := range dateHeaders {
			key := date.Format(simpleDateLayout)
			if _, ok := tl[key]; !ok {
				tl[key] = &FullInfo{
					Date: date,
				}
			}
			if date.IsZero() {
				log.Fatalf("invalid date for column %d: %v", i, row[i+3])
			}
			amount := vartypes.StringToInt(row[i+3])
			switch index {
			case 1: //cases
				tl[key].Cases = amount
			case 3: //total_reinfections
				tl[key].TotalReinfections = amount
			case 4: //deaths
				tl[key].Deaths = amount
			case 5: //deaths_cum
				tl[key].DeathsCum = amount
			case 6: //recovered
				tl[key].Recovered = amount
			case 8: //hospital_admissions
				tl[key].HospitalAdmissions = amount
			case 9: //hospital_discharges
				tl[key].HospitalDischarges = amount
			case 12: //intubated
				tl[key].Intubated = amount
			case 13: //intubated_unvac
				tl[key].IntubatedUnvac = amount
			case 14: //intubated_vac
				tl[key].IntubatedVac = amount
			case 15: //icu_occupancy
				tl[key].IcuOccupancy = vartypes.StringToFloat(row[i+3])
			case 16: //beds_occupancy
				tl[key].BedsOccupancy = vartypes.StringToFloat(row[i+3])
			case 18: //estimated_new_rtpcr_tests
				tl[key].EstimatedNewRtpcrTests = amount
			case 20: //esitmated_new_rapid_tests
				tl[key].EstimatedNewRapidTests = amount
			case 21: //estimated_new_total_tests
				tl[key].EstimatedNewTotalTests = amount
			default:
				// do nothing
			}
		}
	}

	end := time.Time{}
	start := time.Now().Add(1000 * time.Hour * 24)
	for _, fl := range tl {
		if err := s.repo.AddFullInfo(ctx, fl); err != nil {
			log.Fatal(err)
		}
		if fl.Date.Before(start) {
			start = fl.Date
		} else if fl.Date.After(end) {
			end = fl.Date
		}
	}

	log.Printf(
		"added full info for %d dates, from %s until %s", len(tl),
		start.Format(simpleDateLayout),
		end.Format(simpleDateLayout),
	)

	return nil
}

func csvHeaderToDate(s string) (time.Time, error) {
	parts := strings.Split(s, "/")
	newParts := make([]string, len(parts))
	for i := range parts {
		if len(parts[i]) == 1 {
			newParts[i] = "0" + parts[i]
			continue
		}
		newParts[i] = parts[i]
	}
	s = strings.Join(newParts, "/")
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse date %s: %s", s, err)
	}

	return t, nil
}
