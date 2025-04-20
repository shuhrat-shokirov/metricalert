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
	FileStorePath string `env:"FILE_STORE_PATH"`
	DatabaseDsn   string `env:"DATABASE_DSN"`
	HashKey       string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	StoreInterval int    `env:"STORE_INTERVAL" envDefault:"-1"`
	Restore       bool   `env:"RESTORE"`
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func buildInfo() {
	log.Printf("Version: %s\n", buildVersion)
	log.Printf("Date: %s\n", buildDate)
	log.Printf("Commit: %s\n", buildCommit)
}

func main() {
	var defaultParams configParams

	err := env.Parse(&defaultParams)
	if err != nil {
		log.Fatalf("can't parse env: %v", err)
	}

	const (
		defaultAddr          = "localhost:8080"
		defaultStoreInterval = 300
		defaultFileStorePath = "store.json"
		defaultRestore       = true
	)

	serverAddr := flag.String("a", defaultAddr, "server address")
	storeInterval := flag.Int("i", defaultStoreInterval, "store interval")
	fileStorePath := flag.String("f", defaultFileStorePath, "file store path")
	restore := flag.Bool("r", defaultRestore, "restore")
	dataBaseDsn := flag.String("d", "", "database dsn")
	hashKey := flag.String("k", "", "hash key")
	cryptoKey := flag.String("c", "", "crypto key")
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

	if defaultParams.DatabaseDsn != "" {
		dataBaseDsn = &defaultParams.DatabaseDsn
	}

	if defaultParams.HashKey != "" {
		hashKey = &defaultParams.HashKey
	}

	portService := func() int64 {
		split := strings.Split(*serverAddr, ":")
		const splitLen = 2

		if len(split) != splitLen {
			log.Fatalf("can't parse address: %s", *serverAddr)
		}

		port, newErr := strconv.ParseInt(split[1], 10, 64)
		if newErr != nil {
			log.Fatalf("can't parse port: %v", newErr)
		}

		return port
	}()

	if defaultParams.CryptoKey != "" {
		cryptoKey = &defaultParams.CryptoKey
	}

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

	buildInfo()

	if err := run(&config{
		port:          portService,
		storeInterval: *storeInterval,
		fileStorePath: *fileStorePath,
		restore:       *restore,
		logger:        *logger.Sugar(),
		databaseDsn:   *dataBaseDsn,
		hashKey:       *hashKey,
		cryptoKey:     *cryptoKey,
	}); err != nil {
		logger.Fatal("can't run server", zap.Error(err))
	}
}
