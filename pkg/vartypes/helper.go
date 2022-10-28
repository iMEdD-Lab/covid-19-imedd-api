package vartypes

import (
	"strconv"
)

func StringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func StringToInt(s string) int {
	return int(StringToFloat(s))
}
