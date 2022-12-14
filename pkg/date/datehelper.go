package date

import (
	"time"
)

func WeekToDateRange(year int, week int) []time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, time.Local)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year {
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week {
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}

	var res []time.Time
	iter := date
	endDate := date.AddDate(0, 0, 6)
	for !iter.After(endDate) {
		res = append(res, iter)
		iter = iter.AddDate(0, 0, 1)
	}

	return res
}
