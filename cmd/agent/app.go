package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

var (
	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
)

func init() {
	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "poll interval")
	flag.Parse()

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})
}

func run() error {
	client := client.NewClient(fmt.Sprintf("http://%s", addr))
	collector := services.NewCollector()

	agent := application.NewApplication(client, collector)

	agent.Start(pollInterval, reportInterval)

	return nil
}
