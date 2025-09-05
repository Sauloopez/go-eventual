package rabbit

import (
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQProducer struct {
	channelHandler *ChannelHandler
}

func NewProducer(connectionHandler *ConnectionHandler) *RabbitMQProducer {
	return &RabbitMQProducer{
		channelHandler: NewChannelHandler(connectionHandler),
	}
}

func (p *RabbitMQProducer) Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	channel, err := p.channelHandler.GetChannel()
	if err != nil {
		return err
	}

	return channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (p *RabbitMQProducer) Close() error {
	log.Print("[LOG] Closing producer...")
	return p.channelHandler.Close()
}
