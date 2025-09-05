package cron

import (
	"context"
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"log"
	"time"

	"gorm.io/gorm"
)

func StartJob(ctx context.Context, db *gorm.DB, rabbitMQ *rabbit.RabbitMQ, errorChannel chan<- error) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	// use custom scheduler with while loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := sendEventsToRabbitMQ(db, rabbitMQ, errorChannel)
			if err != nil {
				errorChannel <- err
			}
		}
	}
}

func sendEventsToRabbitMQ(database *gorm.DB, rabbitMQ *rabbit.RabbitMQ, errorChannel chan<- error) error {
	minTime := time.Now()
	// max ten seconds
	maxTime := time.Now().Add(time.Second * 10)
	events := db.QueryEventsAt(database, &minTime, &maxTime)
	lenEvents := len(events)

	if lenEvents <= 0 {
		return nil
	}
	log.Printf("Found %d events...", lenEvents)
	for _, event := range events {
		error := rabbitMQ.PublishDbEvent(database, &event, minTime)
		if error != nil {
			errorChannel <- error
		}
		log.Printf("[LOG] Processed event: %v", event.ID)
	}
	return nil
}
