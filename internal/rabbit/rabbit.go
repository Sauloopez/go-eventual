package rabbit

import (
	"encoding/json"
	"eventual/internal/config"
	"eventual/internal/db"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type RabbitMQ struct {
	connectionHandler *ConnectionHandler
	maxDelay          int64

	consumer *RabbitMqConsumer
	producer *RabbitMQProducer
}

// NewRabbitMQ creates a new connection and channel to RabbitMQ
func NewRabbitMQ(config *config.RabbitMQConfig) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", config.User, config.Password, config.Host, config.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	log.Printf("[LOG] Connected to RabbitMQ in host %s:%s with user %s", config.Host, config.Port, config.User)

	var maxDelay int64
	if config.MaxDelay != "" {
		duration, err := time.ParseDuration(config.MaxDelay)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] parsing max delay: %v", err)
		}
		maxDelay = duration.Milliseconds()
		if maxDelay < 0 {
			return nil, fmt.Errorf("[ERROR] max delay must be positive")
		}
	} else {
		maxDelay = time.Second.Milliseconds() * 10 // ten secons
	}

	connectionHandler := NewConnectionHandler(config, conn)

	consumer := NewConsumer(config.Queue, "eventual", connectionHandler)

	producer := NewProducer(connectionHandler)

	return &RabbitMQ{
		connectionHandler: connectionHandler,
		maxDelay:          maxDelay,
		consumer:          consumer,
		producer:          producer,
	}, nil
}

func (r *RabbitMQ) decodeJSONMessage(msg amqp.Delivery, event *db.EventDto) error {
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		return fmt.Errorf("[ERROR] decoding JSON: %v", err)
	}
	return nil
}

// PublishDbEvent publishes event.Message to event.Exchange
func (r *RabbitMQ) PublishDbEvent(dbConnection *gorm.DB, event *db.Event, currentTime time.Time) error {
	var error error
	if event.TimesRemaining > 0 || event.TimesRemaining == -1 {
		var delay int64
		// validate day schedule
		if len(event.DaySchedules) > 0 {
			isToday := false
			for _, schedule := range event.DaySchedules {
				isToday = schedule.DayNumber == time.Weekday(currentTime.Day())
				if isToday {
					break
				}
			}
			if !isToday {
				fmt.Printf("[LOG] event %d is not scheduled for today", event.ID)
				return nil
			}
		}
		// get event delay
		delay = event.GetDelay(currentTime)
		// validate delay not major to maxDelay
		if delay > r.maxDelay {
			fmt.Printf("[LOG] delay %d is greater than maxDelay %d", delay, r.maxDelay)
			return nil
		}
		// check if LastDispatchedAt is within maxDelay time range
		if event.LastDispatchedAt != 0 && currentTime.UnixMilli()-event.LastDispatchedAt <= r.maxDelay {
			fmt.Printf("[LOG] event %d has been dispatched within maxDelay", event.ID)
			return nil
		}
		headers := amqp.Table{}
		// validate delay major or equal to zero
		if delay >= 0 {
			headers["x-delay"] = delay
			log.Printf("[WARNING] Emitting event %d message with delay: %d ms", event.ID, delay)
		} else {
			log.Printf("[LOG] Event %d has been passed with delay: %d ms", event.ID, delay)
			return nil
		}

		error = r.producer.Publish(event.Exchange, event.RoutingKey, false, false, amqp.Publishing{
			Body:    []byte(event.Message),
			Headers: headers,
		})

		if error != nil {
			return error
		}
		db.AckEventPublished(dbConnection, event)
	}

	db.AckEventTimes(dbConnection, event)

	return error
}

func (r *RabbitMQ) ConfigureConsumer() (<-chan amqp.Delivery, error) {
	return r.consumer.ConsumeQueue()
}

func (r *RabbitMQ) ProcessEventMessage(msg amqp.Delivery, dbConnection *gorm.DB) error {
	event := db.EventDto{}
	if err := r.decodeJSONMessage(msg, &event); err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While decoding JSON: %v", err)
	}

	modelInstance, err := event.Transform()
	if err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While transforming event to model: %v", err)
	}

	if err := dbConnection.Model(&db.Event{}).Create(modelInstance).Error; err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While saving event to database: %v", err)
	}
	msg.Ack(true)
	return nil
}

// Close closes the connection and channel
func (r *RabbitMQ) Close() error {
	var err error
	err = r.producer.Close()
	if err != nil {
		return err
	}
	err = r.consumer.Close()
	err = r.connectionHandler.Close()
	return err
}
