package db

import (
	"eventual/internal/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func QueryEventsAt(database *gorm.DB, minTime *time.Time, maxTime *time.Time) []Event {
	var events []Event
	maxClockMs := utils.GetClockMs(*maxTime)
	minClockMs := utils.GetClockMs(*minTime)

	database.Model(&Event{}).
		Preload("DaySchedules", "day_number = ?", minTime.Weekday()).
		Where("expected_at BETWEEN ? AND ? OR (times_remaining > 0 OR times_remaining = -1) OR expected_clock BETWEEN ? AND ?",
			minTime.UnixMilli(), maxTime.UnixMilli(), minClockMs, maxClockMs).
		Order("created_at DESC").
		Find(&events)

	return events
}

func queryDaySchedules(database *gorm.DB, daysNumber []int8) []DaySchedule {
	var daySchedules []DaySchedule
	database.Model(&DaySchedule{}).Where("day_number IN ?", daysNumber).Find(&daySchedules)
	return daySchedules
}

func relateEventScheduleDays(conn *gorm.DB, event *Event, daysNumber []int8) error {
	daySchedules := queryDaySchedules(conn, daysNumber)
	if len(daysNumber) != len(daySchedules) {
		return fmt.Errorf("[ERROR] number of days and day schedules do not match")
	}
	event.DaySchedules = daySchedules
	return nil
}

func CreateEventFromDto(conn *gorm.DB, eventDto *EventDto) (error, *Event) {
	var error error
	instance, error := eventDto.Transform()
	if error != nil {
		return error, nil
	}
	if len(eventDto.Days) > 0 {
		error = relateEventScheduleDays(conn, instance, eventDto.Days)
		if error != nil {
			return error, nil
		}
	}
	error = conn.Create(&instance).Association("DaySchedules").Append(&instance.DaySchedules)
	return error, instance
}

func AckEventTimes(conn *gorm.DB, event *Event) {
	// exit if times remaining is -1 (infinity)
	if event.TimesRemaining == -1 {
		return
	}
	timesRemaining := event.TimesRemaining - 1
	// reduce on times remaining > 0
	if timesRemaining > 0 {
		conn.Model(&Event{}).Where("id = ?", event.ID).Update("times_remaining = ?", timesRemaining)
	}
	// delete event if times remaining is 0
	if timesRemaining <= 0 {
		conn.Delete(&event)
	}
}

func AckEventPublished(conn *gorm.DB, event *Event) {
	dispatchedAt := time.Now().UnixMilli()
	event.LastDispatchedAt = dispatchedAt
	conn.Save(&event)
}
