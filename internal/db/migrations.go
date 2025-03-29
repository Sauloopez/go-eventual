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

	for dayNumber := 0; dayNumber < 7; dayNumber++ {
		weekDay := time.Weekday(dayNumber)
		longName := weekDay.String()
		var daySchedule = DaySchedule{Name: longName, DayNumber: weekDay}
		if err := db.Where("day_number = ?", daySchedule.DayNumber).FirstOrCreate(&daySchedule).Error; err != nil {
			return fmt.Errorf("error creating DaySchedule %s: %w", longName, err)
		}
	}
	return nil
}
