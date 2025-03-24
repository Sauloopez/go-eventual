package core

import (
	"eventual/internal/config"
	"eventual/internal/cron"
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Eventual struct {
	RabbitMQ *rabbit.RabbitMQ
	Db       *gorm.DB
}

func NewEventual() (*Eventual, error) {
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
	}, nil

}

func Start(eventual *Eventual) error {
	consumer, err := eventual.RabbitMQ.ConfigureConsumer()
	if err != nil {
		return err
	}

	errChan := make(chan error)
	var wg sync.WaitGroup

	// RabbitMQ messaging processing goroutine
	wg.Add(1)
	go func() {
		defer close(errChan)
		defer wg.Done()
		log.Print("Starting RabbitMQ")
		for delivery := range consumer {
			if err := eventual.RabbitMQ.ProcessEventMessage(delivery, eventual.Db); err != nil {
				errChan <- fmt.Errorf("[ERROR] Processing message: %w", err)
			}
		}
	}()

	// Task cron goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Print("Starting cron job...")
		if _, err := cron.StartCronJob(eventual.Db, eventual.RabbitMQ); err != nil {
			fmt.Printf("err: %+v\n", err)
			errChan <- fmt.Errorf("[ERROR] Running cron job: %v", err)
		}
	}()

	// Error monitoring goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errChan {
			log.Printf("[ERROR] %v", err)
		}
	}()

	go func() {
		wg.Wait()
		log.Println("Exiting...")
	}()

	return nil
}
