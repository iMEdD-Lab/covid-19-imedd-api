package data

import (
	"context"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"covid19-greece-api/pkg/file"
)

// Repository for storing all COVID data.

type Repo interface {
	AddCase(ctx context.Context, date time.Time, amount int, sluggedRegionalUnit string) error
	AddFullInfo(ctx context.Context, fi *FullInfo) error
	AddRegionalUnit(ctx context.Context, rgu RegionalUnit) error
	GetRegionalUnits(ctx context.Context) ([]RegionalUnit, error)
	GetCases(ctx context.Context, filter CasesFilter) ([]Case, error)
	GetFromTimeline(ctx context.Context, filter DatesFilter) ([]FullInfo, error)
	AddYearlyDeath(ctx context.Context, munId, deaths, year int) error
	AddMunicipality(ctx context.Context, name string) (int, error)
	GetMunicipalities(ctx context.Context) ([]Municipality, error)
	GetDeathsPerMunicipality(ctx context.Context, filter DeathsFilter) ([]YearlyDeaths, error)
	GetDemographicInfo(ctx context.Context, filter DemographicFilter) ([]DemographicInfo, error)
	AddDemographicInfo(ctx context.Context, info DemographicInfo) error
}

type YpesMunicipality struct {
	Name         string
	Slug         string
	Code         string
	Population11 int
	Population21 int
}

type DatesFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

type CasesFilter struct {
	DatesFilter
	RegionalUnitId int
}

type PgRepo struct {
	conn    *pgxpool.Pool
	csvInfo map[string]YpesMunicipality
}

func NewPgRepo(conn *pgxpool.Pool, municipalitiesYpesCsvFile string) (*PgRepo, error) {
	ypesInfo := make(map[string]YpesMunicipality)
	if len(municipalitiesYpesCsvFile) == 0 {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return nil, fmt.Errorf("runtime.Caller error")
		}
		municipalitiesYpesCsvFile = filepath.Join(path.Dir(filename), "municipalities_ypes.csv")
	}

	data, err := file.ReadCsv(municipalitiesYpesCsvFile, true)
	if err != nil {
		return nil, fmt.Errorf("cannot read municipalities_ypes.csv: %s", err)
	}
	for i := 1; i < len(data); i++ {
		name := data[i][1]
		slugged := data[i][2]
		code := data[i][3]
		population11, err := strconv.Atoi(data[i][4])
		if err != nil {
			return nil, fmt.Errorf("municipalities_ypes.csv error at line %d: cannot convert %s to int: %s",
				i, data[i][4], err)
		}
		population21, err := strconv.Atoi(data[i][5])
		if err != nil {
			return nil, fmt.Errorf("municipalities_ypes.csv error at line %d: cannot convert %s to int: %s",
				i, data[i][5], err)
		}
		ypesInfo[slugged] = YpesMunicipality{
			Name:         name,
			Slug:         slugged,
			Code:         code,
			Population11: population11,
			Population21: population21,
		}
	}
	log.Printf("municipality info from YPES loaded successfully")
	return &PgRepo{
		conn:    conn,
		csvInfo: ypesInfo,
	}, nil
}

func (r *PgRepo) AddCase(ctx context.Context, date time.Time, amount int, slugged string) error {
	sql := `INSERT INTO cases_per_regional_unit (regional_unit_id, date, cases) 
            VALUES ((SELECT id FROM regional_units WHERE slug=$1), $2, $3) ON CONFLICT (regional_unit_id, date) DO UPDATE SET cases=$3`
	_, err := r.conn.Exec(ctx, sql, slugged, date, amount)
	if err != nil {
		return fmt.Errorf("could not insert row: %v", err)
	}

	return nil
}

func (r *PgRepo) AddFullInfo(ctx context.Context, fi *FullInfo) error {
	sql := `INSERT INTO greece_timeline (date,cases,total_reinfections,deaths,deaths_cum,recovered,beds_occupancy,
			 icu_occupancy,intubated,intubated_vac,intubated_unvac,hospital_admissions,hospital_discharges,
			 estimated_new_rtpcr_tests,estimated_new_rapid_tests,estimated_new_total_tests,cases_cum,waste_highest_place,
             waste_highest_percentage,waste_highest_place_en) 
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT (date) DO UPDATE SET 
           cases=$2,total_reinfections=$3,deaths=$4,deaths_cum=$5,recovered=$6,beds_occupancy=$7,icu_occupancy=$8,
           intubated=$9,intubated_vac=$10,intubated_unvac=$11,hospital_admissions=$12,hospital_discharges=$13,
           estimated_new_rtpcr_tests=$14,estimated_new_rapid_tests=$15,estimated_new_total_tests=$16,cases_cum=$17,
           waste_highest_place=$18,waste_highest_percentage=$19,waste_highest_place_en=$20`
	_, err := r.conn.Exec(ctx, sql, fi.Date, fi.Cases, fi.TotalReinfections, fi.Deaths, fi.DeathsCum, fi.Recovered,
		fi.BedsOccupancy, fi.IcuOccupancy, fi.Intubated, fi.IntubatedVac, fi.IntubatedUnvac, fi.HospitalAdmissions,
		fi.HospitalDischarges, fi.EstimatedNewRtpcrTests, fi.EstimatedNewRapidTests, fi.EstimatedNewTotalTests,
		fi.CasesCum, fi.WasteHighestPlace, fi.WasteHighestPercent, fi.WasteHighestPlaceEn)
	if err != nil {
		return fmt.Errorf("error inserting into greece_timeline table: %s", err)
	}

	return nil
}

func (r *PgRepo) AddRegionalUnit(ctx context.Context, ru RegionalUnit) error {
	sql := `INSERT INTO regional_units (slug, department, prefecture, regional_unit_normalized, regional_unit, pop_11) 
            VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (regional_unit_normalized) DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, ru.Slug, ru.Department, ru.Prefecture,
		ru.RegionalUnitNormalized, ru.RegionalUnit, ru.Pop11)
	if err != nil {
		return fmt.Errorf("could not insert regional_units row: %v", err)
	}

	return nil
}

func (r *PgRepo) GetRegionalUnits(ctx context.Context) ([]RegionalUnit, error) {
	sql := `SELECT id,slug,department,prefecture,regional_unit_normalized,regional_unit,pop_11 from regional_units`
	rows, err := r.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("could not get regional unit from db: %s", err)
	}

	var res []RegionalUnit
	for rows.Next() {
		var g RegionalUnit
		if err := rows.Scan(&g.Id, &g.Slug, &g.Department, &g.Prefecture,
			&g.RegionalUnitNormalized, &g.RegionalUnit, &g.Pop11); err != nil {
			return nil, fmt.Errorf("could not scan regional_units row: %s", err)
		}
		res = append(res, g)
	}

	return res, nil
}

func (r *PgRepo) GetMunicipalities(ctx context.Context) ([]Municipality, error) {
	sql := `SELECT id,name,slug,code,pop_11,pop_21 FROM municipalities`
	rows, err := r.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("cannot get from municipalities table: %s", err)
	}
	var res []Municipality
	for rows.Next() {
		var m Municipality
		if err := rows.Scan(&m.Id, &m.Name, &m.Slug, &m.Code, &m.Population11, &m.Population21); err != nil {
			return nil, fmt.Errorf("could not scan municipalities row: %s", err)
		}
		res = append(res, m)
	}
	return res, nil
}

type Case struct {
	RegionalUnitId int       `json:"regional_unit_id"`
	Date           time.Time `json:"date"`
	Cases          int       `json:"cases"`
}

func (r *PgRepo) GetCases(ctx context.Context, filter CasesFilter) ([]Case, error) {
	sql := `SELECT regional_unit_id,date,cases FROM cases_per_regional_unit WHERE 1=1 `
	counter := 1
	var args []interface{}

	if filter.RegionalUnitId > 0 {
		sql += fmt.Sprintf(" AND regional_unit_id=$%d ", counter)
		counter++
		args = append(args, filter.RegionalUnitId)
	}

	if !filter.StartDate.IsZero() {
		sql += fmt.Sprintf(" AND date >= $%d ", counter)
		counter++
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		sql += fmt.Sprintf(" AND date <= $%d ", counter)
		counter++
		args = append(args, filter.EndDate)
	}

	sql += " ORDER BY date ASC "

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get cases from db: %s", err)
	}

	var res []Case
	for rows.Next() {
		var c Case
		if err := rows.Scan(&c.RegionalUnitId, &c.Date, &c.Cases); err != nil {
			return nil, fmt.Errorf("could not scan cases row: %s", err)
		}
		res = append(res, c)
	}

	return res, nil
}

func (r *PgRepo) GetFromTimeline(ctx context.Context, filter DatesFilter) ([]FullInfo, error) {
	sql := `SELECT date,cases,total_reinfections,deaths,deaths_cum,recovered,beds_occupancy,
			 icu_occupancy,intubated,intubated_vac,intubated_unvac,hospital_admissions,hospital_discharges,
			 estimated_new_rtpcr_tests,estimated_new_rapid_tests,estimated_new_total_tests,cases_cum,waste_highest_place,
             waste_highest_percentage,waste_highest_place_en FROM greece_timeline WHERE 1=1 `
	var args []interface{}
	counter := 1
	if !filter.StartDate.IsZero() {
		sql += fmt.Sprintf(" AND date >= $%d ", counter)
		counter++
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		sql += fmt.Sprintf(" AND date <= $%d ", counter)
		counter++
		args = append(args, filter.EndDate)
	}

	sql += " ORDER BY date ASC "

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("db error getting from greece_timeline: %s", err)
	}

	var fullInfos []FullInfo
	for rows.Next() {
		var fi FullInfo
		if err := rows.Scan(&fi.Date, &fi.Cases, &fi.TotalReinfections, &fi.Deaths, &fi.DeathsCum, &fi.Recovered,
			&fi.BedsOccupancy, &fi.IcuOccupancy, &fi.Intubated, &fi.IntubatedVac, &fi.IntubatedUnvac,
			&fi.HospitalAdmissions, &fi.HospitalDischarges, &fi.EstimatedNewRtpcrTests, &fi.EstimatedNewRapidTests,
			&fi.EstimatedNewTotalTests, &fi.CasesCum, &fi.WasteHighestPlace, &fi.WasteHighestPercent,
			&fi.WasteHighestPlaceEn); err != nil {
			return nil, fmt.Errorf("db error scanning greece_timeline: %s", err)
		}
		fullInfos = append(fullInfos, fi)
	}

	return fullInfos, nil
}

func (r *PgRepo) AddYearlyDeath(ctx context.Context, munId, deaths, year int) error {
	sql := `INSERT INTO deaths_per_municipality_cum (year, municipality_id, deaths_cum) VALUES($1,$2,$3)
			ON CONFLICT (year, municipality_id) DO UPDATE SET deaths_cum=$3`
	if _, err := r.conn.Exec(ctx, sql, year, munId, deaths); err != nil {
		return fmt.Errorf("could not add to deaths_per_municipality_cum: %s", err)
	}

	return nil
}

func (r *PgRepo) AddMunicipality(ctx context.Context, name string) (int, error) {
	slugged := slug.Make(name)
	sql := `SELECT id from municipalities WHERE slug=$1`
	var id int
	row := r.conn.QueryRow(ctx, sql, slugged)
	err := row.Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return id, fmt.Errorf("cannot get municipality by name: %s", err)
	}
	sql = `INSERT INTO municipalities (name, slug, code, pop_11, pop_21) VALUES ($1,$2,$3,$4,$5)  
		   ON CONFLICT DO NOTHING RETURNING id`
	fromCsv, ok := r.csvInfo[slugged]
	if !ok {
		return 0, fmt.Errorf("name %s, slugged %s, not found in municipalities ypes csv", name, slugged)
	}
	row = r.conn.QueryRow(ctx, sql, name, slugged, fromCsv.Code, fromCsv.Population11, fromCsv.Population21)
	if err := row.Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println()
		}
		return id, fmt.Errorf("cannot add municipality: %s", err)
	}

	return id, nil
}

type DeathsFilter struct {
	MunId int
	Year  int
}

func (r *PgRepo) GetDeathsPerMunicipality(ctx context.Context, filter DeathsFilter) ([]YearlyDeaths, error) {
	sql := `SELECT year,municipality_id,deaths_cum FROM deaths_per_municipality_cum WHERE 1=1 `
	counter := 1
	var args []interface{}

	if filter.MunId > 0 {
		sql += fmt.Sprintf(` AND municipality_id=$%d `, counter)
		counter++
		args = append(args, filter.MunId)
	}

	if filter.Year > 0 {
		sql += fmt.Sprintf(` AND year=$%d `, counter)
		counter++
		args = append(args, filter.Year)
	}

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot query deaths_per_municipality_cum: %s", err)
	}

	var res []YearlyDeaths
	for rows.Next() {
		var y YearlyDeaths
		if err := rows.Scan(&y.Year, &y.MunId, &y.Deaths); err != nil {
			return nil, fmt.Errorf("cannot scan deaths_per_municipality_cum row: %s", err)
		}
		res = append(res, y)
	}

	return res, nil
}

func (r *PgRepo) AddDemographicInfo(ctx context.Context, info DemographicInfo) error {
	sql := `INSERT INTO demography_per_age (date,category,cases,deaths,intensive,discharged,hospitalized,
            hospitalized_in_icu,passed_away,recovered,treated_at_home) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) 
            ON CONFLICT (date,category) DO UPDATE SET cases=$3,deaths=$4,intensive=$5,discharged=$6,hospitalized=$7, 
            hospitalized_in_icu=$8,passed_away=$9,recovered=$10,treated_at_home=$11`
	_, err := r.conn.Exec(ctx, sql, info.Date, info.Category, info.Cases, info.Deaths, info.Intensive, info.Discharged,
		info.Hospitalized, info.HospitalizedInIcu, info.PassedAway, info.Recovered, info.TreatedAtHome)
	if err != nil {
		return fmt.Errorf("cannot add demographic info: %s", err)
	}
	return nil
}

type DemographicFilter struct {
	DatesFilter
	Category string
}

func (r *PgRepo) GetDemographicInfo(ctx context.Context, filter DemographicFilter) ([]DemographicInfo, error) {
	sql := `SELECT date,category,cases,deaths,intensive,discharged,hospitalized,hospitalized_in_icu,passed_away,
       recovered,treated_at_home FROM demography_per_age WHERE 1=1 `
	var args []interface{}
	counter := 1
	if !filter.StartDate.IsZero() {
		sql += fmt.Sprintf(" AND date >= $%d ", counter)
		counter++
		args = append(args, filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		sql += fmt.Sprintf(" AND date <= $%d ", counter)
		counter++
		args = append(args, filter.EndDate)
	}

	if len(filter.Category) > 0 {
		sql += fmt.Sprintf(" AND category = $%d ", counter)
		counter++
		args = append(args, filter.Category)
	}

	sql += " ORDER BY date ASC "

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot get demographic info: %s", err)
	}
	var res []DemographicInfo
	for rows.Next() {
		var info DemographicInfo
		if err := rows.Scan(&info.Date, &info.Category, &info.Cases, &info.Deaths, &info.Intensive, &info.Discharged,
			&info.Hospitalized, &info.HospitalizedInIcu, &info.PassedAway, &info.Recovered,
			&info.TreatedAtHome); err != nil {
			return nil, fmt.Errorf("cannot scan demographic info: %s", err)
		}
		res = append(res, info)
	}
	return res, nil
}
