package main

import (
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

func run(addr string, reportInterval, pollInterval time.Duration) {
	newClient := client.NewClient(addr)
	collector := services.NewCollector()

	agent := application.NewApplication(newClient, collector)

	agent.Start(pollInterval, reportInterval)
}
