package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Event represents an event in the database
type Event struct {
	gorm.Model
	ExpectedAt     int64         `gorm:"index"` // Unix milisecond timestamp of the expected date
	Message        string        `gorm:"not null"`
	TimesRemaining int           `gorm:"default:1"`
	DaySchedules   []DaySchedule `gorm:"many2many:event_day_schedules"`
	Exchange       string        `gorm:"index"`
	ExpectedClock  int           `gorm:"index"`
}

func (event *Event) GetDelay(currentTime time.Time) string {
	var delay string
	currentMs := currentTime.UnixMilli()
	clockMs := (currentTime.Hour()*3600 + currentTime.Minute()*60 + currentTime.Second()) * 1000
	if event.ExpectedAt > currentMs {
		delay = fmt.Sprintf("%d", event.ExpectedAt-currentMs)
	}
	if event.ExpectedClock > clockMs {
		delay = fmt.Sprintf("%d", event.ExpectedClock-clockMs)
	}
	return delay
}

type DaySchedule struct {
	DayNumber time.Weekday `gorm:"primaryKey"`
	Name      string       `gorm:"not null;unique"`
	Events    []Event      `gorm:"many2many:event_day_schedules"`
}
