package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Agenda struct {
	Events          []*calendar.Event
	ProcessedEvents []ProcessedEvent
	Service         *calendar.Service
}

type ProcessedEvent struct {
	Summary string
	// date string formatted 2006-01-02
	Date   string
	Time   int64
	AllDay bool
}

func NewAgenda(client *http.Client, calendarIds []string) *Agenda {
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		slog.Error("Unable to retrieve Calendar client", err)
	}

	agenda := &Agenda{
		Service: srv,
	}

	for _, id := range calendarIds {
		agenda.GetEvents(id, 20)
	}
	return agenda
}

func (a *Agenda) GetEvents(calendarId string, maxResults int64) {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Format(time.RFC3339)

	events, err := a.Service.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(maxResults).OrderBy("startTime").Do()
	if err != nil {
		slog.Error("Unable to retrieve next ten of the user's events", "error", err)
	}

	a.Events = append(a.Events, events.Items...)
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

func prettyDateOnly(d string) string {
	t, err := time.Parse(time.DateOnly, d)
	if err != nil {
		slog.Error("unable to parse date string. expecting time.DateOnly", "date", d, "error", err)
	}

	return t.Format("Monday Jan 02")
}

func (a *Agenda) Generate() bytes.Buffer {
	a.ProcessEvents()
	a.SortEvents()
	lines := 0
	var output bytes.Buffer
	var currentDate string
	for _, event := range a.ProcessedEvents {
		if lines >= MaxAgendaLines-2 {
			slog.Info("no more room in the buffer. ignoring the rest", "lines", lines)
			break
		}

		eDate := prettyDateOnly(event.Date)
		if currentDate != eDate {
			currentDate = prettyDateOnly(event.Date)
			heading := "\n" + currentDate + "\n"
			lines += 2
			output.WriteString(heading)
		}
		output.WriteString(event.Summary)
		lines++
	}

	slog.Info("full agenda", "size", output.Len())
	return output
}

func (a *Agenda) ProcessEvents() {
	if len(a.Events) < 1 {
		slog.Info("No upcoming events found.")
		return
	}

	for _, item := range a.Events {
		pEvent := parseDateAndTimeFromEvent(item)
		eTime := convertEventTime(*pEvent)
		pEvent.Summary = fmt.Sprintf("%8s: %s\n", eTime, item.Summary)
		a.ProcessedEvents = append(a.ProcessedEvents, *pEvent)
	}
}

func convertEventTime(e ProcessedEvent) string {
	if e.AllDay {
		return "All day"
	}
	return time.Unix(e.Time, 0).Format(time.Kitchen)
}

func (a *Agenda) SortEvents() {
	eventDate := func(d1, d2 *ProcessedEvent) bool {
		return d1.Date < d2.Date
	}
	eventTime := func(t1, t2 *ProcessedEvent) bool {
		return t1.Time < t2.Time
	}
	eventSummary := func(s1, s2 *ProcessedEvent) bool {
		return s1.Summary < s2.Summary
	}
	OrderedBy(eventDate, eventTime, eventSummary).Sort(a.ProcessedEvents)
}
