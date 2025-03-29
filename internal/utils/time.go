package utils

import "time"

func GetClockMs(expectedClock time.Time) int {
	hours := expectedClock.Hour()
	minutes := expectedClock.Minute()
	seconds := expectedClock.Second()
	return (hours*3600 + minutes*60 + seconds) * 1000
}
