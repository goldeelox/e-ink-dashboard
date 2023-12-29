package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Agenda struct {
	Events          []*calendar.Event
	Service         *calendar.Service
	ProcessedEvents []ProcessedEvent
}

type ProcessedEvent struct {
	Summary string
	// date string formatted 2006-01-02
	Date string
	Time string
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

func parseDateAndTimeFromEvent(e *calendar.Event) (eventDate string, eventTime string) {
	if e.Start.Date != "" {
		return e.Start.Date, "All day"
	}

	t, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error("unable to parse datetime", "error", err)
	}
	return t.Format(time.DateOnly), t.Format(time.Kitchen)
}

func prettyDateOnly(d string) string {
	t, err := time.Parse(time.DateOnly, d)
	if err != nil {
		slog.Error("unable to parse date string. expecting time.DateOnly", "date", d, "error", err)
	}

	return t.Format("Monday Jan 02")
}

func eventsToMap(events []*calendar.Event) map[string][]string {
	eventMap := make(map[string][]string)
	for _, item := range events {
		eDate, eTime := parseDateAndTimeFromEvent(item)
		eSummary := fmt.Sprintf("%8s: %s\n", eTime, item.Summary)
		eventMap[eDate] = append(eventMap[eDate], eSummary)
	}
	return eventMap
}

func (a *Agenda) ProcessEvents() bytes.Buffer {
	// TODO: return []ProcessedEvents
	if len(a.Events) < 1 {
		slog.Info("No upcoming events found.")
		return bytes.Buffer{}
	}

	agenda := eventsToMap(a.Events)
	// sort the map keys
	agendaKeys := make([]string, 0, len(agenda))
	for k := range agenda {
		agendaKeys = append(agendaKeys, k)
	}
	sort.Strings(agendaKeys)

	var output bytes.Buffer
	lines := 0

finish:
	for _, agendaDate := range agendaKeys {
		if lines >= MaxAgendaLines-2 {
			slog.Info("no more room in the buffer. ignoring the rest", "lines", lines)
			break finish
		}
		// TODO: make this controllable in the request payload
		heading := "\n" + prettyDateOnly(agendaDate) + "\n"
		lines += 2

		output.WriteString(heading)
		for _, agendaSummary := range agenda[agendaDate] {
			if lines >= MaxAgendaLines {
				slog.Info("no more room in the buffer. ignoring the rest", "lines", lines)
				break finish
			}
			lines++
			output.WriteString(agendaSummary)
		}
	}

	slog.Info("full agenda", "size", output.Len())
	return output
}

func (a *Agenda) ProcessEventsNew() {
	if len(a.Events) < 1 {
		slog.Info("No upcoming events found.")
		return
	}

	for _, item := range a.Events {
		eDate, eTime := parseDateAndTimeFromEvent(item)
		eSummary := fmt.Sprintf("%8s: %s\n", eTime, item.Summary)
		e := ProcessedEvent{
			Summary: eSummary,
			Date:    eDate,
			Time:    eTime,
		}
		a.ProcessedEvents = append(a.ProcessedEvents, e)
	}
}

// TODO
func (a *Agenda) SortEvents() {}
