package main

import (
	"log/slog"
	"time"

	"google.golang.org/api/calendar/v3"
)

// TODO: this should include a time.Time field and methods to return strings formatted how we want
type ProcessedEvent struct {
	Summary string
	// date string formatted 2006-01-02
	Date   string
	Time   int64
	AllDay bool
}

func parseDateAndTimeFromEvent(e *calendar.Event) *ProcessedEvent {
	if e.Start.Date != "" {
		t, _ := time.ParseInLocation(time.DateOnly, e.Start.Date, time.Local)
		return &ProcessedEvent{
			Date:   e.Start.Date,
			Time:   t.Unix(),
			AllDay: true,
		}
	}

	t, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error("unable to parse datetime", "error", err)
	}
	return &ProcessedEvent{
		Date:   t.Format(time.DateOnly),
		Time:   t.Unix(),
		AllDay: false,
	}
}

func convertEventTime(e ProcessedEvent) string {
	if e.AllDay {
		return "All day"
	}
	t, _ := time.Parse(time.DateOnly, e.Date)
	_, tzOffset := t.Local().Zone()
	return time.Unix(e.Time+int64(tzOffset), 0).Format(time.Kitchen)
}
