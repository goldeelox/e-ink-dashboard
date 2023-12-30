package main

import (
	"log/slog"
	"time"

	"google.golang.org/api/calendar/v3"
)

// TODO: this should include a time.Time field and methods to return strings formatted how we want
type AgendaEvent struct {
	Summary  string
	AllDay   bool
	DateTime time.Time
}

func NewAgendaEvent(e *calendar.Event) *AgendaEvent {
	agendaEvent := &AgendaEvent{
		Summary: e.Summary,
		AllDay:  false,
	}

	if e.Start.Date != "" {
		t, _ := time.ParseInLocation(time.DateOnly, e.Start.Date, time.Local)
		agendaEvent.AllDay = true
		agendaEvent.DateTime = t
		return agendaEvent
	}

	t, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error("unable to parse datetime", "error", err)
	}
	agendaEvent.DateTime = t
	return agendaEvent
}

func (e *AgendaEvent) DateString() string {
	return e.DateTime.Format("Monday Jan 02")
}

func (e *AgendaEvent) TimeString() string {
	if e.AllDay {
		return "All day"
	}

	return e.DateTime.Format(time.Kitchen)
}
