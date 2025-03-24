package db

import (
	"errors"
	"fmt"
	"time"
)

// EventDto represents a data transfer object for an event.
type EventDto struct {
	Message       string // event message content
	ExpectedAt    string // event expected time
	Exchange      string // event RabbitMQ exchange name send to
	DaySchedule   int8   // event day schedule
	ExpectedClock string // event expected clock time
	Times         int    // expected times to send event
}

func Transform(dto *EventDto) (*Event, error) {
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

	if dto.ExpectedClock == "" && dto.ExpectedAt == "" {
		return nil, errors.New("expected clock or expected at is required")
	}

	if dto.ExpectedAt != "" {
		expectedDate, err := time.Parse("2006-01-02", dto.ExpectedAt)
		if err != nil {
			return nil, fmt.Errorf("expected at must be in format YYYY-MM-DD: %v", err)
		}
		modelInstance.ExpectedAt = expectedDate
	}

	if dto.ExpectedClock != "" {
		expectedClock, err := time.Parse("15:04", dto.ExpectedClock)
		if err != nil {
			return nil, fmt.Errorf("expected clock must be in format HH:MM: %v", err)
		}
		modelInstance.ExpectedClock = expectedClock
	}

	if dto.DaySchedule != 0 && dto.DaySchedule < 1 || dto.DaySchedule > 7 {
		return nil, errors.New("day schedule must be between 1 and 7")
	}

	return &modelInstance, nil
}
