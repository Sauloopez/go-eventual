package rabbit

import (
	"encoding/json"
	"eventual/internal/config"
	"eventual/internal/db"
	"fmt"
	"log"

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

// PublishMessage publishes a message to the queue
func (r *RabbitMQ) PublishMessage(message string) {
	err := r.Channel.Publish(
		"",          // exchange
		r.QueueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Printf("Error publicando mensaje: %v", err)
	}
	log.Printf("Mensaje publicado en RabbitMQ: %s", message)
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
	var event db.EventDto
	if err := r.decodeJSONMessage(msg, &event); err != nil {
		msg.Nack(false, false)
		return fmt.Errorf("[ERROR] While decoding JSON: %v", err)
	}

	if err := dbConnection.Create(&event).Error; err != nil {
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
