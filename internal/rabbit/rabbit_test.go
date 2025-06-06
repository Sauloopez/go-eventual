package rabbit_test

import (
	"eventual/internal/config"
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestGetDelay(t *testing.T) {
	conn, err := db.NewDBConnection(path.Join("..", "..", "data"))
	if err != nil {
		t.Fatal(err)
	}
	currentTime := time.Now()
	maxDateTime := currentTime.Add(time.Hour)
	expectedDelay := maxDateTime.Sub(currentTime).Milliseconds()
	fmt.Printf("expectedDelay: %v\n", expectedDelay)
	dto := db.EventDto{
		Message: "test message",
		//ExpectedClock: "10:25:01",
		ExpectedClock: maxDateTime.Format(time.TimeOnly),
		Exchange:      "test-exchange",
		Days:          []int8{int8(currentTime.Weekday() + 1)},
	}

	error, event := db.CreateEventFromDto(conn, &dto)
	if error != nil {
		t.Fatal(error)
	}

	queryResult := db.QueryEventsAt(conn, &currentTime, &maxDateTime)
	dbEvent := queryResult[0]
	if dbEvent.ID != event.ID {
		t.Errorf("expected event ID %d, got %d", event.ID, dbEvent.ID)
	}
	fmt.Printf("dbEvent.DaySchedules: %v\n", dbEvent.DaySchedules)
	delay := dbEvent.GetDelay(currentTime)
	fmt.Printf("result delay: %v\n", delay)
	if delay != expectedDelay {
		t.Errorf("expected delay %v, got %v", expectedDelay, delay)
	}

	conn.Delete(event)
}

func TestPublishEvent(t *testing.T) {
	conn, err := db.NewDBConnection(path.Join("..", "..", "data"))
	if err != nil {
		t.Fatal(err)
	}
	godotenv.Load(path.Join("..", "..", ".env"))
	config, err := config.BuildConfig()
	if err != nil {
		t.Fatal(err)
	}

	dto := db.EventDto{
		Message:       "test message",
		ExpectedClock: "10:25:00",
		ExpectedAt:    "2025-03-27",
		Exchange:      "test-exchange",
	}

	expectedAt := time.Date(2025, 3, 27, 10, 0, 0, 0, time.UTC)

	error, instance := db.CreateEventFromDto(conn, &dto)

	rabbit, error := rabbit.NewRabbitMQ(config.RabbitMQConfig)
	if error != nil {
		t.Fatal(error)
	}

	rabbit.PublishDbEvent(conn, instance, expectedAt)

	// retry again to validate max delay
	rabbit.PublishDbEvent(conn, instance, expectedAt)

	conn.Delete(&instance)
}
