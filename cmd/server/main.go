package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type configParams struct {
	Addr          string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStorePath string `env:"FILE_STORE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func main() {
	var defaultParams configParams

	err := env.Parse(&defaultParams)
	if err != nil {
		log.Fatalf("can't parse env: %v", err)
	}

	serverAddr := flag.String("a", "localhost:8080", "server address")
	storeInterval := flag.Int("i", 300, "store interval")
	fileStorePath := flag.String("f", "store.json", "file store path")
	restore := flag.Bool("r", true, "restore")

	flag.Parse()

	if defaultParams.Addr != "" {
		serverAddr = &defaultParams.Addr
	}

	if defaultParams.StoreInterval != -1 {
		storeInterval = &defaultParams.StoreInterval
	}

	if defaultParams.FileStorePath != "" {
		fileStorePath = &defaultParams.FileStorePath
	}

	if defaultParams.Restore {
		restore = &defaultParams.Restore
	}

	portService := func() int64 {
		split := strings.Split(*serverAddr, ":")
		if len(split) != 2 {
			log.Fatalf("can't parse address: %s", *serverAddr)
		}
		port, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			log.Fatalf("can't parse port: %v", err)
		}

		return port
	}()

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			log.Fatalf("can't parse flag: %s", f.Name)
		}
	})

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("can't sync logger: %v", err)
		}
	}()

	if err := run(config{
		port:          portService,
		storeInterval: *storeInterval,
		fileStorePath: *fileStorePath,
		restore:       *restore,
		logger:        *logger.Sugar(),
	}); err != nil {
		logger.Fatal("can't run server", zap.Error(err))
	}
}
