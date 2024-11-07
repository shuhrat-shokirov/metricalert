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
	addr           = "http://localhost:8080"
	reportInterval time.Duration
	pollInterval   time.Duration
)

func init() {
	serverAddr := flag.String("a", "localhost:8080", "server address")
	report := flag.Int64("r", 10, "report interval")
	poll := flag.Int64("p", 2, "poll interval")
	flag.Parse()

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
