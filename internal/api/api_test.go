package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"covid19-greece-api/internal/data"
)

type ApiSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	api  *Api
	repo *data.RepoMock
}

func (s *ApiSuite) SetupSuite() {
	mockCtrl := gomock.NewController(s.T())
	s.ctrl = mockCtrl
	repo := data.NewRepoMock(s.ctrl)
	s.repo = repo
	srv, _ := data.NewService(
		repo,
		"../data/test_csv/testing_cases.csv",
		"../data/test_csv/testing_timeline.csv",
		"../data/test_csv/testing_deaths.csv",
		"../data/test_csv/testing_demographics.csv",
		"../data/test_csv/testing_waste.csv",
		true,
	)
	s.api = NewApi(
		repo,
		srv,
		"abcd",
	)
}

func (s *ApiSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func (s *ApiSuite) TestHealth() {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)
}

func (s *ApiSuite) TestGetRegionalUnits() {
	expected := []data.RegionalUnit{{
		Id:                     1,
		Slug:                   "aitoloakarnanias",
		Department:             "Στερεά Ελλάδα",
		Prefecture:             "Περιφέρεια Δυτικής Ελλάδας",
		RegionalUnitNormalized: "ΑΙΤΩΛΟΑΚΑΡΝΑΝΙΑΣ",
		RegionalUnit:           "Π.Ε. Αιτωλοακαρνανίας",
		Pop11:                  210802,
	}, {
		Id:                     2,
		Slug:                   "argolidas",
		Department:             "Πελοπόννησος",
		Prefecture:             "Περιφέρεια Πελοποννήσου",
		RegionalUnitNormalized: "ΑΡΓΟΛΙΔΑΣ",
		RegionalUnit:           "Π.Ε. Αργολίδας",
		Pop11:                  97044,
	}}
	s.repo.EXPECT().GetRegionalUnits(gomock.Any()).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/regional_units", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var regionalUnits []data.RegionalUnit
	err = json.Unmarshal(bodyBytes, &regionalUnits)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), expected, regionalUnits)
}

func (s *ApiSuite) TestGetMunicipalities() {
	expected := []data.Municipality{{
		Id:   1,
		Name: "Municipality 1",
		Slug: "municipality-1",
	}, {
		Id:   2,
		Name: "Municipality 2",
		Slug: "municipality-2",
	}}
	s.repo.EXPECT().GetMunicipalities(gomock.Any()).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/municipalities", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var muns []data.Municipality
	err = json.Unmarshal(bodyBytes, &muns)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), expected, muns)
}

func (s *ApiSuite) TestGetDeathsPerMunicipality() {
	expected := []data.YearlyDeaths{{
		MunId:  1,
		Deaths: 123,
		Year:   2021,
	}}
	s.repo.EXPECT().GetDeathsPerMunicipality(gomock.Any(), data.DeathsFilter{
		MunId: 1,
		Year:  2021,
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/deaths_per_municipality?year=2021&municipality_id=1", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var deaths []data.YearlyDeaths
	err = json.Unmarshal(bodyBytes, &deaths)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), expected, deaths)
}

func (s *ApiSuite) TestGetCases() {
	expected := []data.Case{{
		RegionalUnitId: 1,
		Date:           time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Cases:          234,
	}, {
		RegionalUnitId: 3,
		Date:           time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Cases:          45454,
	}}
	s.repo.EXPECT().GetCases(gomock.Any(), data.CasesFilter{
		RegionalUnitId: 1,
		DatesFilter: data.DatesFilter{
			StartDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
		},
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/cases?regional_unit_id=1&start_date=2021-01-01&end_date=2021-01-10", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var cases []data.Case
	err = json.Unmarshal(bodyBytes, &cases)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), expected, cases)
}

func (s *ApiSuite) TestGetTimelineFields() {
	req, _ := http.NewRequest(http.MethodGet, "/timeline_fields", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)
	var fields []string
	err = json.Unmarshal(bodyBytes, &fields)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), tlFields, fields)
}

func (s *ApiSuite) TestGetTimeline() {
	expected := []data.FullInfo{{
		Date:                   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Cases:                  1,
		TotalReinfections:      2,
		Deaths:                 3,
		DeathsCum:              4,
		Recovered:              5,
		HospitalAdmissions:     6,
		HospitalDischarges:     7,
		Intubated:              8,
		IntubatedVac:           9,
		IntubatedUnvac:         10,
		IcuOccupancy:           11,
		BedsOccupancy:          12,
		EstimatedNewRtpcrTests: 13,
		EstimatedNewRapidTests: 14,
		EstimatedNewTotalTests: 15,
	}, {
		Date:                   time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Cases:                  100,
		TotalReinfections:      200,
		Deaths:                 300,
		DeathsCum:              400,
		Recovered:              500,
		HospitalAdmissions:     600,
		HospitalDischarges:     700,
		Intubated:              800,
		IntubatedVac:           900,
		IntubatedUnvac:         1000,
		IcuOccupancy:           1100,
		BedsOccupancy:          1200,
		EstimatedNewRtpcrTests: 1300,
		EstimatedNewRapidTests: 1400,
		EstimatedNewTotalTests: 1500,
	}}
	s.repo.EXPECT().GetFromTimeline(gomock.Any(), data.DatesFilter{
		StartDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/timeline?start_date=2021-01-01&end_date=2021-01-10", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var info []data.FullInfo
	err = json.Unmarshal(bodyBytes, &info)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), expected, info)
}

func (s *ApiSuite) TestGetTimelineWithSpecificFields() {
	expected := []data.FullInfo{{
		Date:                   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Cases:                  1,
		TotalReinfections:      2,
		Deaths:                 3,
		DeathsCum:              4,
		Recovered:              5,
		HospitalAdmissions:     6,
		HospitalDischarges:     7,
		Intubated:              8,
		IntubatedVac:           9,
		IntubatedUnvac:         10,
		IcuOccupancy:           11,
		BedsOccupancy:          12,
		EstimatedNewRtpcrTests: 13,
		EstimatedNewRapidTests: 14,
		EstimatedNewTotalTests: 15,
	}, {
		Date:                   time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Cases:                  100,
		TotalReinfections:      200,
		Deaths:                 300,
		DeathsCum:              400,
		Recovered:              500,
		HospitalAdmissions:     600,
		HospitalDischarges:     700,
		Intubated:              800,
		IntubatedVac:           900,
		IntubatedUnvac:         1000,
		IcuOccupancy:           1100,
		BedsOccupancy:          1200,
		EstimatedNewRtpcrTests: 1300,
		EstimatedNewRapidTests: 1400,
		EstimatedNewTotalTests: 1500,
	}}
	s.repo.EXPECT().GetFromTimeline(gomock.Any(), data.DatesFilter{
		StartDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/timeline?start_date=2021-01-01&end_date=2021-01-10&fields=beds_occupancy,total_reinfections", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var info []map[string]interface{}
	err = json.Unmarshal(bodyBytes, &info)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), info[0]["beds_occupancy"], float64(12))
	assert.Equal(s.T(), info[0]["total_reinfections"], float64(2))
	assert.NotContains(s.T(), info[0], "intubated")
	assert.NotContains(s.T(), info[0], "intubated_vac")
	assert.Equal(s.T(), info[1]["beds_occupancy"], float64(1200))
	assert.Equal(s.T(), info[1]["total_reinfections"], float64(200))
	assert.NotContains(s.T(), info[1], "deaths")
	assert.NotContains(s.T(), info[1], "deaths_cum")
}

func (s *ApiSuite) TestGetTimelineOneField() {
	expected := []data.FullInfo{{
		Date:                   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Cases:                  1,
		TotalReinfections:      2,
		Deaths:                 3,
		DeathsCum:              4,
		Recovered:              5,
		HospitalAdmissions:     6,
		HospitalDischarges:     7,
		Intubated:              8,
		IntubatedVac:           9,
		IntubatedUnvac:         10,
		IcuOccupancy:           11,
		BedsOccupancy:          12,
		EstimatedNewRtpcrTests: 13,
		EstimatedNewRapidTests: 14,
		EstimatedNewTotalTests: 15,
	}, {
		Date:                   time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Cases:                  100,
		TotalReinfections:      200,
		Deaths:                 300,
		DeathsCum:              400,
		Recovered:              500,
		HospitalAdmissions:     600,
		HospitalDischarges:     700,
		Intubated:              800,
		IntubatedVac:           900,
		IntubatedUnvac:         1000,
		IcuOccupancy:           1100,
		BedsOccupancy:          1200,
		EstimatedNewRtpcrTests: 1300,
		EstimatedNewRapidTests: 1400,
		EstimatedNewTotalTests: 1500,
	}}
	s.repo.EXPECT().GetFromTimeline(gomock.Any(), data.DatesFilter{
		StartDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/beds_occupancy?start_date=2021-01-01&end_date=2021-01-10", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var info []map[string]interface{}
	err = json.Unmarshal(bodyBytes, &info)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), info[0]["beds_occupancy"], float64(12))
	assert.NotContains(s.T(), info[0], "intubated")
	assert.NotContains(s.T(), info[0], "total_reinfections")
	assert.Equal(s.T(), info[1]["beds_occupancy"], float64(1200))
}

func (s *ApiSuite) TestGetDemographics() {
	expected := []data.DemographicInfo{{
		Date:              time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Category:          "0-17",
		Cases:             1,
		Deaths:            2,
		Intensive:         3,
		Discharged:        4,
		Hospitalized:      5,
		HospitalizedInIcu: 6,
		PassedAway:        7,
		Recovered:         8,
		TreatedAtHome:     9,
	}, {
		Date:              time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
		Category:          "18-39",
		Cases:             10,
		Deaths:            11,
		Intensive:         12,
		Discharged:        13,
		Hospitalized:      14,
		HospitalizedInIcu: 15,
		PassedAway:        16,
		Recovered:         17,
		TreatedAtHome:     18,
	}}
	s.repo.EXPECT().GetDemographicInfo(gomock.Any(), data.DemographicFilter{
		DatesFilter: data.DatesFilter{
			StartDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		Category: "18-39",
	}).Times(1).Return(expected, nil)
	req, _ := http.NewRequest(http.MethodGet, "/demographics?start_date=2021-01-01&end_date=2021-01-10&category=18-39", nil)
	w := httptest.NewRecorder()
	s.api.Router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(s.T(), 200, w.Code)
	bodyBytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(s.T(), err)

	var info []data.DemographicInfo
	err = json.Unmarshal(bodyBytes, &info)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), expected, info)
}
