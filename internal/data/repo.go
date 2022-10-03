package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Repo interface {
	AddCase(ctx context.Context, date time.Time, amount int, sluggedPrefecture string) error
	AddFullInfo(ctx context.Context, fi *FullInfo) error
	AddGeoRow(ctx context.Context, geoInfo GeoInfo) error
}

type PgRepo struct {
	conn *pgxpool.Pool
}

func NewPgRepo(conn *pgxpool.Pool) *PgRepo {
	return &PgRepo{conn: conn}
}

func (r *PgRepo) AddCase(ctx context.Context, date time.Time, amount int, sluggedPrefecture string) error {
	sql := "INSERT INTO cases_per_prefecture (geo_id, date, cases) " +
		"VALUES ((SELECT id FROM greece_geo_info WHERE slug=$1), $2, $3) ON CONFLICT DO NOTHING"
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

func (r *PgRepo) AddGeoRow(ctx context.Context, geoInfo GeoInfo) error {
	sql := `INSERT INTO greece_geo_info (slug, department, prefecture, county_normalized, county, pop_11) ` +
		`VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	_, err := r.conn.Exec(ctx, sql, geoInfo.Slug, geoInfo.Department, geoInfo.Prefecture, geoInfo.CountyNormalized,
		geoInfo.County, geoInfo.Pop11)
	if err != nil {
		return fmt.Errorf("could not insert greece_geo_info row: %v", err)
	}

	return nil
}
