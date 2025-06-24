package core

import (
	"context"
	"eventual/internal/config"
	"eventual/internal/cron"
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Eventual struct {
	RabbitMQ *rabbit.RabbitMQ
	Db       *gorm.DB
	Ctx      context.Context
}

func NewEventual(ctx context.Context) (*Eventual, error) {
	logger := log.Default()
	err := godotenv.Load()
	if err != nil {
		logger.Printf("[WARNING] %s", err)
	}

	config, error := config.BuildConfig()
	if error != nil {
		return nil, error
	}

	rabbitMq, err := rabbit.NewRabbitMQ(config.RabbitMQConfig)
	if err != nil {
		return nil, err
	}

	db, err := db.NewDBConnection(config.DBPath)
	if err != nil {
		return nil, err
	}

	return &Eventual{
		RabbitMQ: rabbitMq,
		Db:       db,
		Ctx:      ctx,
	}, nil

}

func (eventual *Eventual) Start() error {
	consumer, err := eventual.RabbitMQ.ConfigureConsumer()
	if err != nil {
		return err
	}

	errChan := make(chan error)

	// RabbitMQ messaging processing goroutine
	go func() {
		log.Print("[LOG] Starting RabbitMQ Consumer/Producer...")
		for {
			select {
			case <-eventual.Ctx.Done():
				log.Print("[LOG] Stopping RabbitMQ")
				return
			case delivery, ok := <-consumer:
				if !ok {
					return
				}
				if err := eventual.RabbitMQ.ProcessEventMessage(delivery, eventual.Db); err != nil {
					errChan <- fmt.Errorf("[ERROR] Processing message: %w", err)
				}
			}
		}
	}()

	// Task cron goroutine
	go func() {
		log.Print("[LOG] Starting job...")
		if err := cron.StartJob(eventual.Ctx, eventual.Db, eventual.RabbitMQ, errChan); err != nil {
			errChan <- fmt.Errorf("[ERROR] Running cron job: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-eventual.Ctx.Done():
				log.Println("[LOG] Shutting down error monitor...")
				return
			case err := <-errChan:
				log.Printf("[ERROR] %v", err)
			}
		}
	}()

	return nil
}
