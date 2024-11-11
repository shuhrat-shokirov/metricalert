package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type configParams struct {
	Addr string `env:"ADDRESS"`
}

func main() {
	var defaultParams configParams

	err := env.Parse(&defaultParams)
	if err != nil {
		log.Printf("can't parse env: %v\n", err)
		os.Exit(1)
	}

	serverAddr := flag.String("a", "localhost:8080", "server address")

	flag.Parse()

	if defaultParams.Addr != "" {
		serverAddr = &defaultParams.Addr
	}

	portService := func() int64 {
		split := strings.Split(*serverAddr, ":")
		if len(split) != 2 {
			log.Printf("Invalid address: %s\n", *serverAddr)
			os.Exit(1)
		}
		port, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			log.Printf("Invalid port: %s\n", split[1])
			os.Exit(1)
		}

		return port
	}()

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			log.Printf("Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Printf("can't create logger: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("can't sync logger: %v", err)
		}
	}()

	if err := run(portService, *logger.Sugar()); err != nil {
		logger.Error("can't run server", zap.Error(err))
		os.Exit(1)
	}
}
