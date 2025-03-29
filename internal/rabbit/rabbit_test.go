package rabbit_test

import (
	"eventual/internal/db"
	"eventual/internal/rabbit"
	"fmt"
	"path"
	"testing"
	"time"
)

func TestGetDelayForEvent(t *testing.T) {
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
	error, delay := rabbit.GetDelayForEvent(&dbEvent, currentTime)
	fmt.Printf("result delay: %v\n", delay)
	if delay != fmt.Sprintf("%d", expectedDelay) {
		t.Errorf("expected delay %v, got %v", expectedDelay, delay)
	}

	conn.Delete(event)
}
