package data

import (
	"context"
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repo interface {
	AddCase(ctx context.Context, date time.Time, amount int, sluggedPrefecture string) error
	AddFullInfo(ctx context.Context, fi *FullInfo) error
	AddCounty(ctx context.Context, geoInfo GeoInfo) error
	GetGeoInfo(ctx context.Context) ([]GeoInfo, error)
	GetCases(ctx context.Context, filter CasesFilter) ([]Case, error)
	GetFromTimeline(ctx context.Context, filter DatesFilter) ([]FullInfo, error)
	AddYearlyDeath(ctx context.Context, munId, deaths, year int) error
	AddMunicipality(ctx context.Context, name string) (int, error)
}

type DatesFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

type CasesFilter struct {
	DatesFilter
	GeoId int
}

type PgRepo struct {
	conn *pgxpool.Pool
}

func NewPgRepo(conn *pgxpool.Pool) *PgRepo {
	return &PgRepo{conn: conn}
}

func (r *PgRepo) AddCase(ctx context.Context, date time.Time, amount int, sluggedPrefecture string) error {
	sql := `INSERT INTO cases_per_prefecture (geo_id, date, cases) 
            VALUES ((SELECT id FROM counties WHERE slug=$1), $2, $3) ON CONFLICT DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, sluggedPrefecture, date, amount)
	if err != nil {
		return fmt.Errorf("could not insert row: %v", err)
	}

	return nil
}

func (r *PgRepo) AddFullInfo(ctx context.Context, fi *FullInfo) error {
	sql := `INSERT INTO greece_timeline (date,cases,total_reinfections,deaths,deaths_cum,recovered,beds_occupancy,
			 icu_occupancy,intubated,intubated_vac,intubated_unvac,hospital_admissions,hospital_discharges,
			 estimated_new_rtpcr_tests,estimated_new_rapid_tests,estimated_new_total_tests) 
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, fi.Date, fi.Cases, fi.TotalReinfections, fi.Deaths, fi.DeathsCum, fi.Recovered,
		fi.BedsOccupancy, fi.IcuOccupancy, fi.Intubated, fi.IntubatedVac, fi.IntubatedUnvac, fi.HospitalAdmissions,
		fi.HospitalDischarges, fi.EstimatedNewRtpcrTests, fi.EstimatedNewRapidTests, fi.EstimatedNewTotalTests)
	if err != nil {
		return fmt.Errorf("error inserting into greece_timeline table: %s", err)
	}

	return nil
}

func (r *PgRepo) AddCounty(ctx context.Context, geoInfo GeoInfo) error {
	sql := `INSERT INTO counties (slug, department, prefecture, county_normalized, county, pop_11) 
            VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, geoInfo.Slug, geoInfo.Department, geoInfo.Prefecture, geoInfo.CountyNormalized,
		geoInfo.County, geoInfo.Pop11)
	if err != nil {
		return fmt.Errorf("could not insert counties row: %v", err)
	}

	return nil
}

func (r *PgRepo) GetGeoInfo(ctx context.Context) ([]GeoInfo, error) {
	sql := `SELECT id,slug,department,prefecture,county_normalized,county,pop_11 from counties`
	rows, err := r.conn.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("could not get Geo Info from db: %s", err)
	}

	var res []GeoInfo
	for rows.Next() {
		var g GeoInfo
		if err := rows.Scan(&g.Id, &g.Slug, &g.Department, &g.Prefecture,
			&g.CountyNormalized, &g.County, &g.Pop11); err != nil {
			return nil, fmt.Errorf("could not scan Geo Info row: %s", err)
		}
		res = append(res, g)
	}

	return res, nil
}

type Case struct {
	GeoId int       `json:"geo_id"`
	Date  time.Time `json:"date"`
	Cases int       `json:"cases"`
}

func (r *PgRepo) GetCases(ctx context.Context, filter CasesFilter) ([]Case, error) {
	sql := `SELECT geo_id,date,cases FROM cases_per_prefecture WHERE 1=1 `
	counter := 1
	var args []interface{}

	if filter.GeoId > 0 {
		sql += fmt.Sprintf(" AND geo_id=$%d ", counter)
		counter++
		args = append(args, filter.GeoId)
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
		if err := rows.Scan(&c.GeoId, &c.Date, &c.Cases); err != nil {
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
	sql := `INSERT INTO municipalities (name, slug) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING id`
	row := r.conn.QueryRow(ctx, sql, name, slug.Make(name))
	var id int
	if err := row.Scan(&id); err != nil {
		return id, fmt.Errorf("cannot add municipality: %s", err)
	}

	return id, nil
}
