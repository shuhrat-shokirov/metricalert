package main

import (
	"fmt"
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

type config struct {
	addr           string
	hashKey        string
	cryptoKey      string
	reportInterval string
	pollInterval   string
	rateLimit      int64
}

func run(conf config) error {
	newClient := client.NewClient(conf.addr,
		conf.hashKey,
		conf.cryptoKey)
	collector := services.NewCollector()

	agent := application.NewApplication(newClient, collector)

	reportInterval, err := time.ParseDuration(conf.reportInterval)
	if err != nil {
		return fmt.Errorf("can't parse report interval: %w", err)
	}

	pollInterval, err := time.ParseDuration(conf.pollInterval)
	if err != nil {
		return fmt.Errorf("can't parse poll interval: %w", err)
	}

	if conf.rateLimit == 0 {
		conf.rateLimit = 1
	}

	agent.Start(application.Config{
		PoolInterval:   reportInterval,
		ReportInterval: pollInterval,
		RateLimit:      conf.rateLimit,
	})

	return nil
}
