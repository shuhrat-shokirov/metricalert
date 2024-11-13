package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
		log.Printf("can't parse env: %v\n", err)
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

	reportInterval := time.Duration(*report) * time.Second
	pollInterval := time.Duration(*poll) * time.Second
	addr := fmt.Sprintf("http://%s", *serverAddr)

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			log.Printf("Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})

	if err := run(addr, reportInterval, pollInterval); err != nil {
		log.Fatalf("can't run server: %v", err)
	}
}
