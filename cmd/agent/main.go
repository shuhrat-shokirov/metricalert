package main

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type configParams struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

func main() {
	var defaultParams configParams

	err := env.Parse(&defaultParams)
	if err != nil {
		log.Fatalf("can't parse env: %v", err)
	}

	const (
		defaultReportInterval = 10
		defaultPollInterval   = 2
		defaultAddr           = "localhost:8080"
	)

	serverAddr := flag.String("a", defaultAddr, "server address")
	report := flag.Int64("r", defaultReportInterval, "report interval")
	poll := flag.Int64("p", defaultPollInterval, "poll interval")
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

	reportInterval := time.Duration(*report) * time.Second
	pollInterval := time.Duration(*poll) * time.Second
	addr := *serverAddr

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			log.Fatalf("unknown flag: %s", f.Name)
		}
	})

	run(addr, reportInterval, pollInterval)
}
