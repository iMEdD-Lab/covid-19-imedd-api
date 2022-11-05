package data

import (
	"context"
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository for storing all COVID data.

type Repo interface {
	AddCase(ctx context.Context, date time.Time, amount int, sluggedCounty string) error
	AddFullInfo(ctx context.Context, fi *FullInfo) error
	AddCounty(ctx context.Context, county County) error
	GetCounties(ctx context.Context) ([]County, error)
	GetCases(ctx context.Context, filter CasesFilter) ([]Case, error)
	GetFromTimeline(ctx context.Context, filter DatesFilter) ([]FullInfo, error)
	AddYearlyDeath(ctx context.Context, munId, deaths, year int) error
	AddMunicipality(ctx context.Context, name string) (int, error)
	GetMunicipalities(ctx context.Context) ([]Municipality, error)
	GetDeathsPerMunicipality(ctx context.Context, filter DeathsFilter) ([]YearlyDeaths, error)
}

type DatesFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

type CasesFilter struct {
	DatesFilter
	CountyId int
}

type PgRepo struct {
	conn *pgxpool.Pool
}

func NewPgRepo(conn *pgxpool.Pool) *PgRepo {
	return &PgRepo{conn: conn}
}

func (r *PgRepo) AddCase(ctx context.Context, date time.Time, amount int, sluggedCounty string) error {
	sql := `INSERT INTO cases_per_county (county_id, date, cases) 
            VALUES ((SELECT id FROM counties WHERE slug=$1), $2, $3) ON CONFLICT (county_id, date) DO UPDATE SET cases=$3`
	_, err := r.conn.Exec(ctx, sql, sluggedCounty, date, amount)
	if err != nil {
		return fmt.Errorf("could not insert row: %v", err)
	}

	return nil
}

func (r *PgRepo) AddFullInfo(ctx context.Context, fi *FullInfo) error {
	sql := `INSERT INTO greece_timeline (date,cases,total_reinfections,deaths,deaths_cum,recovered,beds_occupancy,
			 icu_occupancy,intubated,intubated_vac,intubated_unvac,hospital_admissions,hospital_discharges,
			 estimated_new_rtpcr_tests,estimated_new_rapid_tests,estimated_new_total_tests) 
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT (date) DO UPDATE SET cases=$2,
           total_reinfections=$3,deaths=$4,deaths_cum=$5,recovered=$6,beds_occupancy=$7,icu_occupancy=$8,intubated=$9,
           intubated_vac=$10,intubated_unvac=$11,hospital_admissions=$12,hospital_discharges=$13,
           estimated_new_rtpcr_tests=$14,estimated_new_rapid_tests=$15,estimated_new_total_tests=$16`
	_, err := r.conn.Exec(ctx, sql, fi.Date, fi.Cases, fi.TotalReinfections, fi.Deaths, fi.DeathsCum, fi.Recovered,
		fi.BedsOccupancy, fi.IcuOccupancy, fi.Intubated, fi.IntubatedVac, fi.IntubatedUnvac, fi.HospitalAdmissions,
		fi.HospitalDischarges, fi.EstimatedNewRtpcrTests, fi.EstimatedNewRapidTests, fi.EstimatedNewTotalTests)
	if err != nil {
		return fmt.Errorf("error inserting into greece_timeline table: %s", err)
	}

	return nil
}

func (r *PgRepo) AddCounty(ctx context.Context, county County) error {
	sql := `INSERT INTO counties (slug, department, prefecture, county_normalized, county, pop_11) 
            VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (county_normalized) DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, county.Slug, county.Department, county.Prefecture, county.CountyNormalized,
		county.County, county.Pop11)
	if err != nil {
		return fmt.Errorf("could not insert counties row: %v", err)
	}

	return nil
}

func (r *PgRepo) GetCounties(ctx context.Context) ([]County, error) {
	sql := `SELECT id,slug,department,prefecture,county_normalized,county,pop_11 from counties`
	rows, err := r.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("could not get County from db: %s", err)
	}

	var res []County
	for rows.Next() {
		var g County
		if err := rows.Scan(&g.Id, &g.Slug, &g.Department, &g.Prefecture,
			&g.CountyNormalized, &g.County, &g.Pop11); err != nil {
			return nil, fmt.Errorf("could not scan counties row: %s", err)
		}
		res = append(res, g)
	}

	return res, nil
}

func (r *PgRepo) GetMunicipalities(ctx context.Context) ([]Municipality, error) {
	sql := `SELECT id,name,slug FROM municipalities`
	rows, err := r.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("cannot get from municipalities table: %s", err)
	}
	var res []Municipality
	for rows.Next() {
		var m Municipality
		if err := rows.Scan(&m.Id, &m.Name, &m.Slug); err != nil {
			return nil, fmt.Errorf("could not scan municipalities row: %s", err)
		}
		res = append(res, m)
	}

	return res, nil
}

type Case struct {
	CountyId int       `json:"county_id"`
	Date     time.Time `json:"date"`
	Cases    int       `json:"cases"`
}

func (r *PgRepo) GetCases(ctx context.Context, filter CasesFilter) ([]Case, error) {
	sql := `SELECT county_id,date,cases FROM cases_per_county WHERE 1=1 `
	counter := 1
	var args []interface{}

	if filter.CountyId > 0 {
		sql += fmt.Sprintf(" AND county_id=$%d ", counter)
		counter++
		args = append(args, filter.CountyId)
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
		if err := rows.Scan(&c.CountyId, &c.Date, &c.Cases); err != nil {
			return nil, fmt.Errorf("could not scan cases row: %s", err)
		}
		res = append(res, c)
	}

	return res, nil
}

func (r *PgRepo) GetFromTimeline(ctx context.Context, filter DatesFilter) ([]FullInfo, error) {
	sql := `SELECT * FROM greece_timeline WHERE 1=1 `
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
			&fi.EstimatedNewTotalTests); err != nil {
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
	sql := `SELECT id from municipalities WHERE name=$1`
	var id int
	row := r.conn.QueryRow(ctx, sql, name)
	err := row.Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return id, fmt.Errorf("cannot get municipality by name: %s", err)
	}
	sql = `INSERT INTO municipalities (name, slug) VALUES ($1,$2) ON CONFLICT (name) DO NOTHING RETURNING id`
	row = r.conn.QueryRow(ctx, sql, name, slug.Make(name))
	if err := row.Scan(&id); err != nil {
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
