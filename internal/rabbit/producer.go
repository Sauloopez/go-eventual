package rabbit

import (
	"eventual/internal/config"

	"github.com/streadway/amqp"
)

type RabbitMQProducer struct {
	connectionHandler *ConnectionHandler
}

func NewProducer(config *config.RabbitMQConfig, connection *amqp.Connection) *RabbitMQProducer {
	connectionHandler := NewConnectionHandler(config, connection)
	return &RabbitMQProducer{
		connectionHandler: connectionHandler,
	}
}

func (p *RabbitMQProducer) Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	channel, err := p.connectionHandler.GetChannel()
	if err != nil {
		return err
	}

	return channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (p *RabbitMQProducer) Close() error {
	return p.connectionHandler.Close()
}
