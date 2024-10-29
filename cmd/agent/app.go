package main

import (
	"time"

	"metricalert/internal/core/application"
	"metricalert/internal/core/services"
	"metricalert/internal/infra/api/client"
)

func run() error {
	client := client.NewClient("http://localhost:8081")
	collector := services.NewCollector()

	agent := application.NewAgent(client, collector)

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	agent.Start(pollInterval, reportInterval)

	return nil
}
