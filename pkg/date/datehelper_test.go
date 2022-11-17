package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekToDateRange(t *testing.T) {
	dateRange := WeekToDateRange(2022, 4)
	assert.Len(t, dateRange, 7)
	for i := 24; i < 31; i++ {
		assert.Contains(
			t,
			dateRange,
			time.Date(2022, 1, i, 0, 0, 0, 0, time.Local),
		)
	}
}

func TestWeekToDateRange2(t *testing.T) {
	dateRange := WeekToDateRange(2022, 9)
	assert.Len(t, dateRange, 7)
	assert.Contains(
		t,
		dateRange,
		time.Date(2022, 2, 28, 0, 0, 0, 0, time.Local),
	)
	for i := 1; i < 7; i++ {
		assert.Contains(
			t,
			dateRange,
			time.Date(2022, 3, i, 0, 0, 0, 0, time.Local),
		)
	}
}
