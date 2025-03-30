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
	minTime := time.Now()
	maxTime := time.Now().Add(time.Minute)
	events := db.QueryEventsAt(database, &minTime, &maxTime)

	for _, event := range events {
		error := rabbitMQ.PublishDbEvent(database, &event, minTime)
		if error != nil {
			errorChannel <- error
		}
		log.Printf("[LOG] Processed event: %v", event.ID)
	}
	return nil
}
