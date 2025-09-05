package rabbit

import (
	"log"

	"github.com/streadway/amqp"
)

type ChannelHandler struct {
	connectionHandler *ConnectionHandler
	channel           *amqp.Channel

	reopenChannel      bool // handles the reopening channel semaphore
	onCloseChannelChan chan *amqp.Error
}

func NewChannelHandler(connectionHandler *ConnectionHandler) *ChannelHandler {
	i := &ChannelHandler{
		connectionHandler:  connectionHandler,
		channel:            nil,
		reopenChannel:      true, // starts in true, because is opening
		onCloseChannelChan: make(chan *amqp.Error, 1),
	}

	go func() {
		i.onChannelClosed()
	}()

	return i
}

func openChannelInConnection(connection *amqp.Connection) (*amqp.Channel, error) {
	log.Print("[WARNING] Generating new channel...")
	channel, err := connection.Channel()
	if err != nil {
		log.Printf("[ERROR] While openning channel %s", err.Error())
		return nil, err
	}
	return channel, err
}

func (p *ChannelHandler) onChannelClosed() {
	if closeError := <-p.onCloseChannelChan; closeError != nil {
		log.Printf("Closed channel. Reason %s", closeError.Reason)
		p.channel = nil
		p.GetChannel()
	}
}

func (p *ChannelHandler) GetChannel() (*amqp.Channel, error) {
	var err error
	log.Print("[LOG] Getting channel...")
	connection, err := p.connectionHandler.getConnection()
	if err != nil {
		return nil, err
	}
	if p.channel == nil && p.reopenChannel {
		p.channel, err = openChannelInConnection(connection)
		p.channel.NotifyClose(p.onCloseChannelChan)
	}

	return p.channel, err
}

func (p *ChannelHandler) Close() error {
	p.reopenChannel = false
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
