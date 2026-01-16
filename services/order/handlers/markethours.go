package handlers

import (
	"os"
	"time"
)

var nyLocation = mustLoadNY()

func mustLoadNY() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.FixedZone("EST", -5*60*60)
	}
	return loc
}

func tradingAlwaysOpen() bool {
	return os.Getenv("GLYPH_TRADING_247") == "true"
}

func IsMarketOpen(now time.Time) bool {
	if tradingAlwaysOpen() {
		return true
	}
	ny := now.In(nyLocation)
	switch ny.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	}
	minutes := ny.Hour()*60 + ny.Minute()
	return minutes >= 9*60+30 && minutes < 16*60
}

func inCloseSweepWindow(now time.Time) bool {
	ny := now.In(nyLocation)
	switch ny.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	}
	minutes := ny.Hour()*60 + ny.Minute()
	return minutes >= 16*60 && minutes < 16*60+10
}
