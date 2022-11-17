package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekToDateRange(t *testing.T) {
	//24-30
	expStart := time.Date(2022, 1, 24, 0, 0, 0, 0, time.Local)
	expEnd := time.Date(2022, 1, 30, 0, 0, 0, 0, time.Local)
	start, end := WeekToDateRange(2022, 4)
	assert.Equal(t, expStart, start)
	assert.Equal(t, expEnd, end)
}
