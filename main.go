package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOKEN_FILE string = "token.json"
)

var (
	mqttBroker         string
	calendarIds        calendarIdsFlag
	credentialsFile    string
	maxAgendaLines     int
	mqttTopicNamespace string
)

func forever() {
	for {
		time.Sleep(5 * time.Second)
	}
}

func init() {
	flag.StringVar(&credentialsFile, "oauth-credentials", "credentials.json", "path to google oauth2 credentials file")
	flag.StringVar(&mqttBroker, "broker", "", "MQTT broker URI (e.g, mqtt://mqtt.example.com:1883)")
	flag.StringVar(&mqttTopicNamespace, "topic-namespace", "kitchen/dashboard/agenda",
		"MQTT topic namespace. Message will be published to and subscribed from <topic-namespace>/response and <topic-namespace>/request respectively")
	flag.Var(&calendarIds, "calendar-id", "ID of Google calendar to fetch events from. Can be specified multiple times.")
	flag.IntVar(&maxAgendaLines, "max-agenda-lines", 36, "Maximum lines to return when creating the agenda")
	flag.Parse()

	slog.Info("config parsed",
		"calendars", calendarIds.String(),
		"broker", mqttBroker,
	)
	// TODO: check for empty flag
}

func main() {
	oauthClient := NewOauth2Client()

	// mqtt
	mqOpts := mqttClientOptions(mqttBroker)

	var RequestMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		slog.Info("got new message",
			"topic", msg.Topic(),
			"msg", string(msg.Payload()),
		)

		agenda := NewAgenda(oauthClient, calendarIds)
		out := agenda.ProcessEvents()
		agenda.ProcessEventsNew()
		slog.Info("processed events", "events",
			fmt.Sprintf("%v", agenda.ProcessedEvents),
		)

		// send agenda to PUB_TOPIC
		client.Publish(mqttTopicNamespace+"/response", byte(0), false, out)
	}

	mqOpts.SetOnConnectHandler(func(client mqtt.Client) {
		slog.Info("connected to mqtt")
		client.Subscribe(mqttTopicNamespace+"/request", byte(0), RequestMessageHandler)
	})

	mqttConnect(mqOpts)

	go forever()
	select {}
}
