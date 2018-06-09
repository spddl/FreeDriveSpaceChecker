package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func formatSize(data float64) string {
	var num float64
	var str string
	switch {
	case data > PB:
		num = data / PB
		str = " PB"
	case data > TB:
		num = data / TB
		str = " TB"
	case data > GB:
		num = data / GB
		str = " GB"
	case data > MB:
		num = data / MB
		str = " MB"
	case data > KB:
		num = data / KB
		str = " KB"
	default:
		num = data
		str = " B"
	}
	return strconv.FormatFloat(num, 'f', 2, 64) + str
}

func fmtDuration(d time.Duration) (string, string) {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d", h), fmt.Sprintf("%02d", m)
}

func durchschnitt(nums []uint64) float64 {
	var sum float64
	for _, num := range nums {
		sum += float64(num)
	}
	durchschnitt := sum / float64(len(nums))
	return durchschnitt
}

// Round ...
func Round(val float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(-places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= .5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
