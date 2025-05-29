package db

import (
	"errors"
	"eventual/internal/utils"
	"fmt"
	"time"
)

// EventDto represents a data transfer object for an event.
type EventDto struct {
	Message       string // event message content
	ExpectedAt    string // event expected time
	Exchange      string // event RabbitMQ exchange name send to
	RoutingKey    string
	Days          []int8 // event days schedule
	ExpectedClock string // event expected clock time
	Times         int    // expected times to send event
}

func (dto *EventDto) Transform() (*Event, error) {
	modelInstance := Event{}
	// validate not empty message
	if dto.Message == "" {
		return nil, errors.New("message is required")
	}

	modelInstance.Message = dto.Message
	// validate not empty exchange
	if dto.Exchange == "" {
		return nil, errors.New("exchange is required")
	}
	modelInstance.Exchange = dto.Exchange

	if dto.RoutingKey == "" {
		return nil, errors.New("routing key is required")
	}
	modelInstance.RoutingKey = dto.RoutingKey

	if dto.ExpectedClock == "" && dto.ExpectedAt == "" {
		return nil, errors.New("expected clock or expected at is required")
	}

	if dto.ExpectedAt != "" {
		expectedDate, err := time.Parse(time.DateOnly, dto.ExpectedAt)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		modelInstance.ExpectedAt = expectedDate.UnixMilli()
	}

	if dto.ExpectedClock != "" {
		expectedClock, err := time.Parse(time.TimeOnly, dto.ExpectedClock)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] expected clock must be in format HH:MM: %v", err)
		}

		modelInstance.ExpectedClock = utils.GetClockMs(expectedClock)
	}

	if len(dto.Days) != 0 {
		for _, day := range dto.Days {
			if day < 1 || day > 7 {
				return nil, errors.New("[ERROR] day schedule must be between 1 and 7")
			}
		}
	}

	return &modelInstance, nil
}
