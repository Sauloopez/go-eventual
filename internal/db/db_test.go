package db_test

import (
	"eventual/internal/db"
	"fmt"
	"path"
	"testing"
	"time"
)

func TestMigration(t *testing.T) {
	conn, err := db.NewDBConnection(path.Join("..", "..", "data"))
	if err != nil {
		t.Fatal(err)
	}
	db.Migrate(conn, true)
}

func TestCreateEvent(t *testing.T) {
	conn, err := db.NewDBConnection(path.Join("..", "..", "data"))
	if err != nil {
		t.Fatal(err)
	}
	dto := db.EventDto{
		Message:       "test message",
		ExpectedClock: "10:25:00",
		ExpectedAt:    "2025-03-27",
		Exchange:      "test-exchange",
	}

	expectedAt := time.Date(2025, 3, 27, 0, 0, 0, 0, time.UTC)

	modelInstance, err := dto.Transform()
	if err != nil {
		t.Fatal(err)
	}

	if expectedAt.UnixMilli() != modelInstance.ExpectedAt {
		t.Errorf("Expected %v, got %v", expectedAt, modelInstance.ExpectedAt)
	}

	data := conn.Create(&modelInstance)
	fmt.Printf("data: %v\n", data)
}

func TestQueryEventsAt(t *testing.T) {
	conn, err := db.NewDBConnection(path.Join("..", "..", "data"))
	if err != nil {
		t.Fatal(err)
	}
	dtoEvents := []db.EventDto{
		{
			Message:       "test message",
			ExpectedClock: "10:25:00",
			ExpectedAt:    "2025-03-27",
			Exchange:      "test-exchange",
		},
		{
			Message:       "test message",
			ExpectedClock: "10:25:00",
			ExpectedAt:    "2025-03-27",
			Exchange:      "test-exchange",
		},
	}

	expectedDateTime := time.Date(2025, 3, 27, 10, 25, 0, 0, time.UTC)
	maxDateTime := time.Date(2025, 3, 27, 10, 26, 0, 0, time.UTC)

	var events []db.Event
	for _, dto := range dtoEvents {
		instance, error := dto.Transform()
		if error != nil {
			t.Fatal(error)
		}
		events = append(events, *instance)
	}
	conn.Create(&events)

	queryResult := db.QueryEventsAt(conn, &expectedDateTime, &maxDateTime)
	fmt.Printf("queryResult: %v\n", queryResult[0])

	conn.Delete(&events)
}
