package main

import (
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

func run(addr string, reportInterval, pollInterval time.Duration) error {
	client := client.NewClient(addr)
	collector := services.NewCollector()

	agent := application.NewApplication(client, collector)

	agent.Start(pollInterval, reportInterval)

	return nil
}
