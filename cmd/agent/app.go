package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"metricalert/internal/agent/core/application"
	"metricalert/internal/agent/core/client"
	"metricalert/internal/agent/core/services"
)

type Network struct {
	Host string
	Port int
}

func (n *Network) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func (n *Network) Set(value string) error {
	split := strings.Split(value, ":")
	if len(split) != 2 {
		return fmt.Errorf("invalid format")
	}

	n.Host = split[0]

	atoi, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	n.Port = atoi

	return nil
}

var (
	addr           = "http://localhost:8080"
	reportInterval time.Duration
	pollInterval   time.Duration
)

func init() {
	network := new(Network)

	flag.Var(network, "a", "server address")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "poll interval")
	flag.Parse()

	if network.Port != 0 {
		addr = fmt.Sprintf("http://%s:%d", network.Host, network.Port)
	}

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
