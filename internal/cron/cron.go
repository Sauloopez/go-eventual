package cron

import (
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func StartCronJob(db *gorm.DB, rabbitMQ *rabbit.RabbitMQ) (*cron.Cron, error) {
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		sendEventsToRabbitMQ(db, rabbitMQ)
	})
	if err != nil {
		return nil, err
	}
	c.Start()
	return c, nil
}

func sendEventsToRabbitMQ(database *gorm.DB, rabbitMQ *rabbit.RabbitMQ) error {
	var events []db.Event
	now := time.Now()
	database.Where("timestamp <= ?", now).Find(&events)

	for _, event := range events {
		rabbitMQ.PublishMessage(event.Message)
		database.Delete(&event)
		log.Printf("Processed event and deleted: %v", event)
	}
	return nil
}
