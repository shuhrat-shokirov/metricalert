package main

import (
	"context"
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

type config struct {
	addr           string
	hashKey        string
	cryptoKey      string
	reportInterval time.Duration
	pollInterval   time.Duration
	rateLimit      int64
}

func run(ctx context.Context, conf *config) {
	newClient := client.NewClient(conf.addr,
		conf.hashKey,
		conf.cryptoKey)
	collector := services.NewCollector()

	agent := application.NewApplication(newClient, collector)

	agent.Start(ctx, application.Config{
		PoolInterval:   conf.pollInterval,
		ReportInterval: conf.reportInterval,
		RateLimit:      conf.rateLimit,
	})
}
