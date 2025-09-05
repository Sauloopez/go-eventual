package rabbit

import (
	"eventual/internal/config"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Handles heartbeat connection with rabbit mq
type ConnectionHandler struct {
	connection *amqp.Connection

	connectionUrl string

	reopenConnection      bool // handles connection reopening semaphore
	onCloseConnectionChan chan *amqp.Error
}

// Factory function to create a new ConnectionHandler
func NewConnectionHandler(config *config.RabbitMQConfig, connection *amqp.Connection) *ConnectionHandler {
	i := &ConnectionHandler{
		connection:            connection,
		reopenConnection:      true, // starts in true, to reopen ever
		onCloseConnectionChan: make(chan *amqp.Error, 1),
		connectionUrl:         fmt.Sprintf("amqp://%s:%s@%s:%s", config.User, config.Password, config.Host, config.Port),
	}
	go func() {
		i.onConnectionClosed()
	}()

	log.Print("build connection handler")
	return i
}

// Returns the current connection or reconnect if nedded
func (p *ConnectionHandler) getConnection() (*amqp.Connection, error) {
	log.Print("[LOG] Getting connection...")
	var err error

	if (p.connection == nil || p.connection.IsClosed()) && p.reopenConnection {
		p.connection, err = generateNewConnection(p.connectionUrl)
		p.connection.NotifyClose(p.onCloseConnectionChan)
	}
	return p.connection, err
}

func generateNewConnection(connectionUrl string) (*amqp.Connection, error) {
	log.Printf("[WARNING] Generating new connection...")
	url := connectionUrl
	connection, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[ERROR] While generating connection %s", err.Error())
		return nil, err
	}
	return connection, err
}

func (p *ConnectionHandler) onConnectionClosed() {
	if closeError := <-p.onCloseConnectionChan; closeError != nil {
		log.Printf("Closed connection. Reason %s", closeError.Reason)
		p.connection = nil
		p.getConnection()
	}
}

// Close the connection definitely stoping reopening it
func (c *ConnectionHandler) Close() error {
	var err error
	c.reopenConnection = false // close definitely connection and do not reopen
	if c.connection != nil {
		err = c.connection.Close()
		if err != nil {
			return err
		}
	}
	return err
}
