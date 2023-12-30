package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Agenda struct {
	Events  []*AgendaEvent
	Service *calendar.Service
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
		agenda.ImportEvents(id, 20)
	}
	return agenda
}

func (a *Agenda) ImportEvents(calendarId string, maxResults int64) {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Format(time.RFC3339)

	events, err := a.Service.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(maxResults).OrderBy("startTime").Do()
	if err != nil {
		slog.Error("Unable to retrieve next ten of the user's events", "error", err)
	}

	if len(events.Items) < 1 {
		slog.Info("No upcoming events found for calendar", "id", calendarId)
		return
	}

	for _, item := range events.Items {
		event := NewAgendaEvent(item)
		a.Events = append(a.Events, event)
	}
}

func (a *Agenda) Output() bytes.Buffer {
	a.SortEvents()
	outputLines := 0
	var output bytes.Buffer
	var currentDate string
	for _, event := range a.Events {
		if outputLines >= MaxAgendaLines-2 {
			slog.Info("no more room in the buffer. ignoring the rest", "lines", outputLines)
			break
		}

		eDate := event.Date()
		eventLines := make([]string, 0)
		if currentDate != eDate {
			currentDate = eDate
			eventLines = append(eventLines, "", currentDate)
		}

		s := fmt.Sprintf("%8s: %s", event.Time(), event.Summary)
		eventLines = append(eventLines, s, "")
		outputLines += len(eventLines)
		output.WriteString(strings.Join(eventLines, "\n"))
	}

	slog.Info("full agenda", "size", output.Len())
	return output
}

func (a *Agenda) SortEvents() {
	eventDateTime := func(d1, d2 *AgendaEvent) bool {
		return d1.DateTime.Before(d2.DateTime)
	}
	eventSummary := func(s1, s2 *AgendaEvent) bool {
		return s1.Summary < s2.Summary
	}
	OrderedBy(eventDateTime, eventSummary).Sort(a.Events)
}
