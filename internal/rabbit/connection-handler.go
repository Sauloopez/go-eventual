package rabbit

import (
	"eventual/internal/config"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type ConnectionHandler struct {
	connection *amqp.Connection
	channel    *amqp.Channel

	config *config.RabbitMQConfig

	connectionIsClosing bool
	channelIsClosing    bool
}

func NewConnectionHandler(config *config.RabbitMQConfig, connection *amqp.Connection) *ConnectionHandler {
	instance := &ConnectionHandler{
		config:              config,
		connection:          connection,
		channel:             nil,
		connectionIsClosing: true,
		channelIsClosing:    true,
	}

	return instance
}

func (p *ConnectionHandler) getConnection() (*amqp.Connection, error) {
	log.Print("[LOG] Getting connection...")
	var err error
	if p.connection == nil || p.connectionIsClosing || p.connection.IsClosed() {
		log.Printf("[WARNING] Generating new connection for user %s", p.config.User)
		url := fmt.Sprintf("amqp://%s:%s@%s:%s", p.config.User, p.config.Password, p.config.Host, p.config.Port)
		p.connection, err = amqp.Dial(url)
		if err != nil {
			log.Printf("[ERROR] While generating connection %s", err.Error())
			return nil, err
		}
		p.connectionIsClosing = false

		onCloseConnectionChan := make(chan *amqp.Error, 1)

		p.connection.NotifyClose(onCloseConnectionChan)

		// goroutine for reconnect
		go func() {
			if closeError := <-onCloseConnectionChan; closeError != nil {
				log.Printf("Closed connection. Reason %s", closeError.Reason)
				p.channel = nil
				p.connection = nil
				p.connectionIsClosing = true
			}
		}()
	}
	return p.connection, err
}

func (p *ConnectionHandler) GetChannel() (*amqp.Channel, error) {
	var err error
	log.Print("[LOG] Getting channel...")
	connection, err := p.getConnection()
	if err != nil {
		return nil, err
	}
	if p.channel == nil || p.channelIsClosing {
		log.Print("[WARNING] Generating new channel...")
		p.channel, err = connection.Channel()
		if err != nil {
			log.Printf("[ERROR] While generating channel %s", err.Error())
			return nil, err
		}
		p.channelIsClosing = false

		onCloseChannelChan := make(chan *amqp.Error, 1)

		p.channel.NotifyClose(onCloseChannelChan)

		go func() {
			if closeError := <-onCloseChannelChan; closeError != nil {
				log.Printf("Closed channel. Reason %s", closeError.Reason)
				p.channelIsClosing = true
				p.channel = nil
			}
		}()
	}

	return p.channel, err
}

func (c *ConnectionHandler) Close() error {
	var err error
	if c.channel != nil {
		err = c.channel.Close()
		if err != nil {
			return err
		}
	}
	if c.connection != nil {
		err = c.connection.Close()
		if err != nil {
			return err
		}
	}
	return err
}
