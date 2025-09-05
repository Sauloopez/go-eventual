package rabbit

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMqConsumer struct {
	channelHandler *ChannelHandler

	queueName     string
	consumerName  string
	queueDeclared bool
}

func NewConsumer(queueName string, consumerName string, connectionHandler *ConnectionHandler) *RabbitMqConsumer {
	return &RabbitMqConsumer{
		channelHandler: NewChannelHandler(connectionHandler),
		queueName:      queueName,
		consumerName:   consumerName,
	}
}

func (consumer *RabbitMqConsumer) getChannel() (*amqp.Channel, error) {
	var err error
	log.Print("[LOG] Getting channel in consumer...")
	channel, err := consumer.channelHandler.GetChannel()

	if !consumer.queueDeclared {
		log.Printf("[WARNING] Declaring queue in consumer...")
		_, err = channel.QueueDeclare(
			consumer.queueName, // name of the queue
			true,               // durable
			false,              // delete when unused
			false,              // exclusive
			false,              // no-wait
			nil,                // arguments
		)
		if err != nil {
			return nil, err
		}
		log.Printf("[LOG] Queue %s already declared, ok...", consumer.queueName)
		consumer.queueDeclared = true
	}

	return channel, err
}

func (consumer *RabbitMqConsumer) ConsumeQueue() (<-chan amqp.Delivery, error) {
	channel, err := consumer.getChannel()
	if err != nil {
		return nil, err
	}
	log.Printf("[LOG] Consuming %s queue as %s", consumer.queueName, consumer.consumerName)
	delivery, err := channel.Consume(
		consumer.queueName,    // queue name
		consumer.consumerName, // Consumer name
		false,                 // autoAck
		false,                 // exclusive
		false,                 // noLocal
		false,                 // noWait
		nil,                   // args
	)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Unexpected error while configuring consumer: %v", err)
	}

	return delivery, nil
}

func (p *RabbitMqConsumer) Close() error {
	log.Print("[LOG] Closing consumer...")
	return p.channelHandler.Close()
}
