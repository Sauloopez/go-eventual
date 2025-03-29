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
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	QueueName    string
	consumerName string
}

// NewRabbitMQ creates a new connection and channel to RabbitMQ
func NewRabbitMQ(config *config.RabbitMQConfig) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", config.User, config.Password, config.Host, config.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare the queue
	_, err = ch.QueueDeclare(
		config.Queue, // name of the queue
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Error declarando la cola: %v", err)
	}

	log.Printf("Connected to RabbitMQ with user %s", config.User)

	return &RabbitMQ{
		Connection:   conn,
		Channel:      ch,
		QueueName:    config.Queue,
		consumerName: "eventual",
	}, nil
}

func (r *RabbitMQ) decodeJSONMessage(msg amqp.Delivery, event *db.EventDto) error {
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		return fmt.Errorf("error decoding JSON: %v", err)
	}
	return nil
}

func GetDelayForEvent(event *db.Event, currentTime time.Time) (error, string) {
	var delay string
	currentMs := currentTime.UnixMilli()
	clockMs := (currentTime.Hour()*3600 + currentTime.Minute()*60 + currentTime.Second()) * 1000
	if len(event.DaySchedules) > 0 {
		isToday := false
		for _, schedule := range event.DaySchedules {
			isToday = schedule.DayNumber == time.Weekday(currentTime.Day())
			if isToday {
				break
			}
		}
		if !isToday {
			return fmt.Errorf("event %d is not scheduled for today", event.ID), delay
		}
	}
	if event.ExpectedAt > currentMs {
		delay = fmt.Sprintf("%d", event.ExpectedAt-currentMs)
	}
	if event.ExpectedClock > clockMs {
		delay = fmt.Sprintf("%d", event.ExpectedClock-clockMs)
	}
	return nil, delay
}

// PublishDbEvent publishes event.Message to event.Exchange
func (r *RabbitMQ) PublishDbEvent(event db.Event, currentTime time.Time) error {
	error, delay := GetDelayForEvent(&event, currentTime)
	if error != nil {
		return error
	}
	headers := amqp.Table{}
	if len(delay) > 0 && delay[0] != '-' {
		headers["x-delay"] = delay
		log.Printf("[WARNING] Emitting event %d message with delay: %s ms", event.ID, delay)
	}

	err := r.Channel.Publish(event.Exchange, "", false, false, amqp.Publishing{
		Body:    []byte(event.Message),
		Headers: headers,
	})
	return err
}

func (r *RabbitMQ) ConfigureConsumer() (<-chan amqp.Delivery, error) {
	channel := r.Channel
	delivery, err := channel.Consume(
		r.QueueName,    // queue name
		r.consumerName, // Consumer name
		false,          // autoAck
		false,          // exclusive
		false,          // noLocal
		false,          // noWait
		nil,            // args
	)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Unexpected error while configuring consumer: %v", err)
	}

	return delivery, nil
}

func (r *RabbitMQ) ProcessEventMessage(msg amqp.Delivery, dbConnection *gorm.DB) error {
	event := db.EventDto{}
	if err := r.decodeJSONMessage(msg, &event); err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While decoding JSON: %v", err)
	}

	modelInstance, err := db.Transform(&event)
	if err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While transforming event to model: %v", err)
	}

	if err := dbConnection.Model(&db.Event{}).Create(modelInstance).Error; err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While saving event to database: %v", err)
	}
	return nil
}

// Close closes the connection and channel
func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Connection.Close()
}
