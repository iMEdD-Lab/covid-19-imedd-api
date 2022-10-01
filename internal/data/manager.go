package data

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4/pgxpool"

	"covid-data-transformation/pkg/file"
	"covid-data-transformation/pkg/vartypes"
)

const (
	dateLayout       = "01/02/06"
	simpleDateLayout = "2006-01-02"
)

type Manager struct {
	conn           *pgxpool.Pool
	casesCsvUrl    string
	timelineCsvUrl string
}

type FullInfo struct {
	Date                   time.Time
	Cases                  int
	TotalReinfections      int
	Deaths                 int
	DeathsCum              int
	Recovered              int
	HospitalAdmissions     int
	HospitalDischarges     int
	Intubated              int
	IntubatedVac           int
	IntubatedUnvac         int
	IcuOccupancy           float64
	BedsOccupancy          float64
	EstimatedNewRtpcrTests int
	EstimatedNewRapidTests int
	EstimatedNewTotalTests int
}

type GeoInfo struct {
	Slug             string `json:"slug"`
	Department       string `json:"department"`
	Prefecture       string `json:"prefecture"`
	CountyNormalized string `json:"county_normalized"`
	County           string `json:"county"`
	Pop11            int    `json:"pop_11"`
}

func NewManager(
	conn *pgxpool.Pool,
	casesCsvUrl string,
	timelineCsvUrl string,
) (*Manager, error) {
	return &Manager{
		conn:           conn,
		casesCsvUrl:    casesCsvUrl,
		timelineCsvUrl: timelineCsvUrl,
	}, nil
}

func (m *Manager) PopulateGeo(ctx context.Context) error {
	data, err := file.ReadCSVFromUrl(m.casesCsvUrl)
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
		err := m.addGeoRow(ctx, GeoInfo{
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

		log.Printf("added region %s", row[2])
	}

	return nil
}

func (m *Manager) PopulateCases(ctx context.Context) error {
	data, err := file.ReadCSVFromUrl(m.casesCsvUrl)
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
			if err := m.addCase(ctx, date, amount, sl); err != nil {
				log.Fatalf("Error adding death day: %v", err)
			}

			log.Printf("added case for date %s and region %s", date.Format(simpleDateLayout), sl)
		}
	}

	return nil
}

func (m *Manager) PopulateTimeline(ctx context.Context) error {
	data, err := file.ReadCSVFromUrl(m.timelineCsvUrl)
	if err != nil {
		log.Fatalf("Error reading csv file: %v", err)
	}

	// take dates from csv first row
	dateHeaders := []time.Time{}
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

	for _, fl := range tl {
		if err := m.addFullInfo(ctx, fl); err != nil {
			log.Fatal(err)
		}

		log.Printf("added full info for date %s", fl.Date.Format(simpleDateLayout))
	}

	return nil
}

func (m *Manager) addCase(ctx context.Context, date time.Time, amount int, sluggedPrefecture string) error {
	sql := "INSERT INTO cases_per_prefecture (geo_id, date, cases) " +
		"VALUES ((SELECT id FROM greece_geo_info WHERE slug=$1), $2, $3) ON CONFLICT DO NOTHING"
	_, err := m.conn.Exec(ctx, sql, sluggedPrefecture, date, amount)
	if err != nil {
		return fmt.Errorf("could not insert row: %v", err)
	}

	return nil
}

func (m *Manager) addFullInfo(ctx context.Context, fi *FullInfo) error {
	sql := `INSERT INTO greece_timeline (date,cases,total_reinfections,deaths,deaths_cum,recovered,beds_occupancy,
			 icu_occupancy,intubated,intubated_vac,intubated_unvac,hospital_admissions,hospital_discharges,
			 estimated_new_rtpcr_tests,estimated_new_rapid_tests,estimated_new_total_tests) 
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT DO NOTHING`
	_, err := m.conn.Exec(ctx, sql, fi.Date, fi.Cases, fi.TotalReinfections, fi.Deaths, fi.DeathsCum, fi.Recovered,
		fi.BedsOccupancy, fi.IcuOccupancy, fi.Intubated, fi.IntubatedVac, fi.IntubatedUnvac, fi.HospitalAdmissions,
		fi.HospitalDischarges, fi.EstimatedNewRtpcrTests, fi.EstimatedNewRapidTests, fi.EstimatedNewTotalTests)
	if err != nil {
		return fmt.Errorf("error inserting into greece_timeline table: %s", err)
	}

	return nil
}

func (m *Manager) addGeoRow(ctx context.Context, geoInfo GeoInfo) error {
	sql := `INSERT INTO greece_geo_info (slug, department, prefecture, county_normalized, county, pop_11) ` +
		`VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	_, err := m.conn.Exec(ctx, sql, geoInfo.Slug, geoInfo.Department, geoInfo.Prefecture, geoInfo.CountyNormalized,
		geoInfo.County, geoInfo.Pop11)
	if err != nil {
		return fmt.Errorf("could not insert greece_geo_info row: %v", err)
	}

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
