package db

import (
	"time"

	"gorm.io/gorm"
)

// Event represents an event in the database
type Event struct {
	gorm.Model
	ExpectedAt     time.Time     `gorm:"index"`
	Message        string        `gorm:"not null"`
	TimesRemaining int           `gorm:"default:1"`
	DaySchedules   []DaySchedule `gorm:"many2many:event_day_schedules"`
	Exchange       string        `gorm:"index"`
	ExpectedClock  time.Time     `gorm:"index"`
}

type DaySchedule struct {
	DayNumber time.Weekday `gorm:"primaryKey"`
	Name      string       `gorm:"not null;unique"`
	Events    []Event      `gorm:"many2many:event_day_schedules"`
}
