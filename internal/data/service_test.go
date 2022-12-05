package data

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DataServiceSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	repoMock *RepoMock
	srv      *Service
}

func (s *DataServiceSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.repoMock = NewRepoMock(s.ctrl)
	path, err := os.Getwd()
	if err != nil {
		assert.Nil(s.T(), err)
	}
	srv, err := NewService(
		s.repoMock,
		filepath.Join(path, "test_csv/testing_cases.csv"),
		filepath.Join(path, "test_csv/testing_timeline.csv"),
		filepath.Join(path, "test_csv/testing_deaths.csv"),
		filepath.Join(path, "test_csv/testing_demographics.csv"),
		filepath.Join(path, "test_csv/testing_waste.csv"),
		true,
	)
	assert.Nil(s.T(), err)
	s.srv = srv
}

func (s *DataServiceSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestDataServiceSuite(t *testing.T) {
	suite.Run(t, new(DataServiceSuite))
}

func (s *DataServiceSuite) TestPopulateRegionalUnits() {
	ctx := context.Background()

	s.repoMock.EXPECT().AddRegionalUnit(gomock.Any(), RegionalUnit{
		Slug:                   "county_1",
		Department:             "Department_1",
		Prefecture:             "Prefecture_1",
		RegionalUnitNormalized: "County_1",
		RegionalUnit:           "county_one",
		Pop11:                  10000,
	})

	s.repoMock.EXPECT().AddRegionalUnit(gomock.Any(), RegionalUnit{
		Slug:                   "county_2",
		Department:             "Department_2",
		Prefecture:             "Prefecture_2",
		RegionalUnitNormalized: "County_2",
		RegionalUnit:           "county_two",
		Pop11:                  20000,
	})

	s.repoMock.EXPECT().AddRegionalUnit(gomock.Any(), RegionalUnit{
		Slug:                   "county_3",
		Department:             "Department_3",
		Prefecture:             "Prefecture_3",
		RegionalUnitNormalized: "County_3",
		RegionalUnit:           "county_three",
		Pop11:                  30000,
	})

	assert.Nil(s.T(), s.srv.PopulateRegionalUnits(ctx))
}

func (s *DataServiceSuite) TestPopulateCases() {
	ctx := context.Background()

	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 26, 0, 0, 0, 0, time.UTC), 1, "county_1")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 27, 0, 0, 0, 0, time.UTC), 2, "county_1")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 28, 0, 0, 0, 0, time.UTC), 3, "county_1")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC), 4, "county_1")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC), 5, "county_1")

	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 26, 0, 0, 0, 0, time.UTC), 6, "county_2")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 27, 0, 0, 0, 0, time.UTC), 7, "county_2")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 28, 0, 0, 0, 0, time.UTC), 8, "county_2")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC), 9, "county_2")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC), 10, "county_2")

	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 26, 0, 0, 0, 0, time.UTC), 11, "county_3")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 27, 0, 0, 0, 0, time.UTC), 12, "county_3")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 28, 0, 0, 0, 0, time.UTC), 13, "county_3")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC), 14, "county_3")
	s.repoMock.EXPECT().AddCase(gomock.Any(), time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC), 15, "county_3")

	assert.Nil(s.T(), s.srv.PopulateCases(ctx))
}

func (s *DataServiceSuite) TestPopulateTimeline() {
	ctx := context.Background()

	s.repoMock.EXPECT().AddFullInfo(gomock.Any(), &FullInfo{
		Date:                   time.Date(2020, 2, 26, 0, 0, 0, 0, time.UTC),
		Cases:                  1,
		TotalReinfections:      3,
		Deaths:                 4,
		DeathsCum:              5,
		Recovered:              6,
		HospitalAdmissions:     8,
		HospitalDischarges:     9,
		Intubated:              12,
		IntubatedVac:           14,
		IntubatedUnvac:         13,
		IcuOccupancy:           15,
		BedsOccupancy:          16,
		EstimatedNewRtpcrTests: 18,
		EstimatedNewRapidTests: 20,
		EstimatedNewTotalTests: 21,
		CasesCum:               22,
		WasteHighestPlace:      "Κουκουβάουνες",
		WasteHighestPlaceEn:    "koukouvaounes",
		WasteHighestPercent:    0.69,
	})
	s.repoMock.EXPECT().AddFullInfo(gomock.Any(), &FullInfo{
		Date:                   time.Date(2020, 2, 27, 0, 0, 0, 0, time.UTC),
		Cases:                  1 + 22,
		TotalReinfections:      3 + 22,
		Deaths:                 4 + 22,
		DeathsCum:              5 + 22,
		Recovered:              6 + 22,
		HospitalAdmissions:     8 + 22,
		HospitalDischarges:     9 + 22,
		Intubated:              12 + 22,
		IntubatedVac:           14 + 22,
		IntubatedUnvac:         13 + 22,
		IcuOccupancy:           15 + 22,
		BedsOccupancy:          16 + 22,
		EstimatedNewRtpcrTests: 18 + 22,
		EstimatedNewRapidTests: 20 + 22,
		EstimatedNewTotalTests: 21 + 22,
		CasesCum:               22 + 22,
		WasteHighestPlace:      "Κουκουβάουνες",
		WasteHighestPlaceEn:    "koukouvaounes",
		WasteHighestPercent:    0.69,
	})
	s.repoMock.EXPECT().AddFullInfo(gomock.Any(), &FullInfo{
		Date:                   time.Date(2020, 2, 28, 0, 0, 0, 0, time.UTC),
		Cases:                  1 + 44,
		TotalReinfections:      3 + 44,
		Deaths:                 4 + 44,
		DeathsCum:              5 + 44,
		Recovered:              6 + 44,
		HospitalAdmissions:     8 + 44,
		HospitalDischarges:     9 + 44,
		Intubated:              12 + 44,
		IntubatedVac:           14 + 44,
		IntubatedUnvac:         13 + 44,
		IcuOccupancy:           15 + 44,
		BedsOccupancy:          16 + 44,
		EstimatedNewRtpcrTests: 18 + 44,
		EstimatedNewRapidTests: 20 + 44,
		EstimatedNewTotalTests: 21 + 44,
		CasesCum:               22 + 44,
		WasteHighestPlace:      "Κουκουβάουνες",
		WasteHighestPlaceEn:    "koukouvaounes",
		WasteHighestPercent:    0.69,
	})

	assert.Nil(s.T(), s.srv.PopulateTimeline(ctx))
}

func (s *DataServiceSuite) TestPopulateMunicipalities() {
	ctx := context.Background()
	s.repoMock.EXPECT().AddMunicipality(gomock.Any(), "Λιλιπούπολης").Return(50, nil)
	s.repoMock.EXPECT().AddMunicipality(gomock.Any(), "Κουκουβάουνες").Return(60, nil)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 50, 1, 2020)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 50, 2, 2021)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 50, 3, 2034)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 60, 10, 2020)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 60, 20, 2021)
	s.repoMock.EXPECT().AddYearlyDeath(gomock.Any(), 60, 30, 2034)

	assert.Nil(s.T(), s.srv.PopulateDeathsPerMunicipality(ctx))
}

func (s *DataServiceSuite) TestPopulateDemographics() {
	ctx := context.Background()
	info1 := DemographicInfo{
		Date:              time.Date(2020, 1, 25, 0, 0, 0, 0, time.UTC),
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
	}
	info2 := DemographicInfo{
		Date:              time.Date(2020, 1, 26, 0, 0, 0, 0, time.UTC),
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
	}
	s.repoMock.EXPECT().AddDemographicInfo(gomock.Any(), info1)
	s.repoMock.EXPECT().AddDemographicInfo(gomock.Any(), info2)

	assert.Nil(s.T(), s.srv.PopulateDemographic(ctx))
}
