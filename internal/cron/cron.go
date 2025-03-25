package cron

import (
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func StartCronJob(db *gorm.DB, rabbitMQ *rabbit.RabbitMQ, errorChannel chan<- error) (*cron.Cron, error) {
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		sendEventsToRabbitMQ(db, rabbitMQ, errorChannel)
	})
	if err != nil {
		return nil, err
	}
	c.Start()
	return c, nil
}

func sendEventsToRabbitMQ(database *gorm.DB, rabbitMQ *rabbit.RabbitMQ, errorChannel chan<- error) error {
	var events []db.Event
	minTime := time.Now()
	maxTime := time.Now().Add(time.Minute)

	database.Model(&db.Event{}).
		Preload("ScheduleDays", "day_number = ?", minTime.Weekday()).
		Where("expectedAt BETWEEN ? AND ? AND (times_remaining > 0 OR times_remaining = -1) AND expected_clock BETWEEN ? AND ?",
			minTime, maxTime, minTime.Hour(), maxTime.Hour()).Find(&events)

	for _, event := range events {
		error := rabbitMQ.PublishDbEvent(event, minTime)
		if error != nil {
			errorChannel <- error
		}
		timesRemaining := event.TimesRemaining - 1
		if timesRemaining > 0 {
			database.Model(&db.Event{}).Where("id = ?", event.ID).Update("times_remaining = ?", timesRemaining)
		}
		log.Printf("Processed event and deleted: %v", event)
	}
	return nil
}
