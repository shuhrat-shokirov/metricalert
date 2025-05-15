package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
	grpcclient "metricalert/internal/agent/infra/grpc"
)

type config struct {
	addr           string
	hashKey        string
	cryptoKey      string
	ipAddress      string
	grpcURL        string
	reportInterval time.Duration
	pollInterval   time.Duration
	rateLimit      int64
}

func run(ctx context.Context, conf *config) {
	var (
		newClient client.Client
		err       error
	)

	if conf.grpcURL != "" {
		newClient, err = grpcclient.NewMetricsClient(conf.grpcURL)
		if err != nil {
			fmt.Printf("failed to create gRPC client: %v\n", err)
			os.Exit(1)
			return
		}
	} else {
		newClient = client.NewClient(conf.addr, conf.hashKey, conf.cryptoKey)
	}

	collector := services.NewCollector()

	agent := application.NewApplication(newClient, collector, conf.ipAddress)

	agent.Start(ctx, application.Config{
		PoolInterval:   conf.pollInterval,
		ReportInterval: conf.reportInterval,
		RateLimit:      conf.rateLimit,
	})
}
