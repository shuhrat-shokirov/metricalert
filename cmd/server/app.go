package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

var portService int64 = 8080

type DefaultParams struct {
	Addr string `env:"ADDRESS"`
}

func init() {

	var defaultParams DefaultParams

	err := env.Parse(&defaultParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	serverAddr := flag.String("a", "localhost:8080", "server address")

	flag.Parse()

	if defaultParams.Addr != "" {
		serverAddr = &defaultParams.Addr
	}

	portService = func() int64 {
		split := strings.Split(*serverAddr, ":")
		if len(split) != 2 {
			fmt.Fprintf(os.Stderr, "Invalid server address: %s\n", *serverAddr)
			os.Exit(1)
		}
		port, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid port: %s\n", split[1])
			os.Exit(1)
		}

		return port
	}()

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})
}

func run() error {
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return err
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication, portService)

	return api.Run()
}
