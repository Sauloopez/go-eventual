package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, automigrate bool) error {
	if automigrate {
		db.AutoMigrate(&Event{}, &DaySchedule{})
	}

	for dayNumber := range 7 {
		weekDay := time.Weekday(dayNumber)
		longName := weekDay.String()
		daySchedule := DaySchedule{Name: longName, DayNumber: weekDay}
		if err := db.FirstOrCreate(&daySchedule).Error; err != nil {
			return fmt.Errorf("[ERROR] creating DaySchedule %s: %w", longName, err)
		}
	}
	return nil
}
