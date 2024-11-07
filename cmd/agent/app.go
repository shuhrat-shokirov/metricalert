package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

var (
	addr           = "http://localhost:8080"
	reportInterval time.Duration
	pollInterval   time.Duration
)

type DefaultParams struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func init() {

	var defaultParams DefaultParams

	err := env.Parse(&defaultParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	serverAddr := flag.String("a", "localhost:8080", "server address")
	report := flag.Int64("r", 10, "report interval")
	poll := flag.Int64("p", 2, "poll interval")
	flag.Parse()

	if defaultParams.Addr != "" {
		serverAddr = &defaultParams.Addr
	}

	if defaultParams.ReportInterval != 0 {
		report = &defaultParams.ReportInterval
	}

	if defaultParams.PollInterval != 0 {
		poll = &defaultParams.PollInterval
	}

	reportInterval = time.Duration(*report) * time.Second
	pollInterval = time.Duration(*poll) * time.Second
	addr = fmt.Sprintf("http://%s", *serverAddr)

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})
}

func run() error {
	client := client.NewClient(addr)
	collector := services.NewCollector()

	agent := application.NewApplication(client, collector)

	agent.Start(pollInterval, reportInterval)

	return nil
}
