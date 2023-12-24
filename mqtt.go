package main

import (
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var ReconnectHandler mqtt.ReconnectHandler = func(client mqtt.Client, opts *mqtt.ClientOptions) {
	slog.Warn("attempting to reconnect to mqtt",
		"isConnected", client.IsConnected(),
		"isConnectionOpen", client.IsConnectionOpen(),
	)
}

var OnConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	slog.Info("connected to mqtt")
}

var ConnectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	slog.Error("lost connection to mqtt",
		"error", err,
		"isConnected", client.IsConnected(),
		"isConnectionOpen", client.IsConnectionOpen(),
	)
}

func mqttClientOptions(broker string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetClientID("dashboard-agenda-server")
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetConnectionLostHandler(ConnectionLostHandler)
	opts.SetMaxReconnectInterval(1 * time.Minute)
	opts.SetOnConnectHandler(OnConnectHandler)
	opts.SetReconnectingHandler(ReconnectHandler)
	return opts
}

func mqttConnect(opts *mqtt.ClientOptions) mqtt.Client {
	client := mqtt.NewClient(opts)

	for {
		if token := client.Connect(); token.WaitTimeout(15*time.Second) && token.Error() != nil {
			slog.Error("initial connection to mqtt failed",
				"error", token.Error(),
				"isConnected", client.IsConnected(),
				"isConnectionOpen", client.IsConnectionOpen(),
			)

			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	return client
}

func checkConnection(c mqtt.Client) bool {
	conn := c.IsConnectionOpen()
	if !conn {
		slog.Warn("connection to mqtt is not open")
	}

	return conn
}

func mqttPublish(c mqtt.Client, message string, topic string) {
	if !checkConnection(c) {
		return
	}

	slog.Debug("publishing message to mqtt",
		"topic", topic,
		"message", message,
	)

	if token := c.Publish(topic, byte(0), false, message); token.WaitTimeout(2*time.Second) && token.Error() != nil {
		slog.Error(token.Error().Error())
	}
}
