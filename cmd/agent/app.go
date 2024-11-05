package main

import (
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

func run() error {
	client := client.NewClient("http://localhost:8080")
	collector := services.NewCollector()

	agent := application.NewApplication(client, collector)

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	agent.Start(pollInterval, reportInterval)

	return nil
}
