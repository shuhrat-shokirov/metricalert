//go:build exclude

package main

import (
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

type config struct {
	addr           string
	hashKey        string
	reportInterval time.Duration
	pollInterval   time.Duration
	rateLimit      int64
}

func run(conf config) {
	newClient := client.NewClient(conf.addr, conf.hashKey)
	collector := services.NewCollector()

	agent := application.NewApplication(newClient, collector)

	agent.Start(application.Config{
		PoolInterval:   conf.pollInterval,
		ReportInterval: conf.reportInterval,
		RateLimit:      conf.rateLimit,
	})
}
