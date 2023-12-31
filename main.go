package main

import (
	"flag"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	MqttBroker         string
	CalendarIds        calendarIdsFlag
	CredentialsFile    string
	MaxAgendaLines     int
	MqttTopicNamespace string
)

func forever() {
	for {
		time.Sleep(5 * time.Second)
	}
}

func init() {
	flag.StringVar(&CredentialsFile, "oauth-credentials", "credentials.json", "path to google oauth2 credentials file")
	flag.StringVar(&MqttBroker, "broker", "", "MQTT broker URI (e.g, mqtt://mqtt.example.com:1883)")
	flag.StringVar(&MqttTopicNamespace, "topic-namespace", "kitchen/dashboard/agenda",
		"MQTT topic namespace. Message will be published to and subscribed from <topic-namespace>/response and <topic-namespace>/request respectively")
	flag.Var(&CalendarIds, "calendar-id", "ID of Google calendar to fetch events from. Can be specified multiple times.")
	flag.IntVar(&MaxAgendaLines, "max-agenda-lines", 36, "Maximum lines to return when creating the agenda")
	flag.Parse()

	slog.Info("config parsed",
		"broker", MqttBroker,
		"calendars", CalendarIds.String(),
		"max-agenda-lines", MaxAgendaLines,
		"topic-namespace", MqttTopicNamespace,
	)
	// TODO: check for empty flag
}

func main() {
	oauthClient := NewOauth2Client()

	// mqtt
	mqOpts := mqttClientOptions(MqttBroker)

	var RequestMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		slog.Info("got new message",
			"topic", msg.Topic(),
			"msg", string(msg.Payload()),
		)

		agenda := NewAgenda(oauthClient, CalendarIds)
		// send agenda to PUB_TOPIC
		client.Publish(MqttTopicNamespace+"/response", byte(0), false, agenda.Output())
	}

	mqOpts.SetOnConnectHandler(func(client mqtt.Client) {
		slog.Info("connected to mqtt")
		client.Subscribe(MqttTopicNamespace+"/request", byte(0), RequestMessageHandler)
	})

	mqttConnect(mqOpts)

	go forever()
	select {}
}
