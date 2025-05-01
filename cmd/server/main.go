package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

type configParams struct {
	Addr          string `json:"address"`
	FileStorePath string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	HashKey       string `json:"-"`
	CryptoKey     string `json:"crypto_key"`
	StoreInterval string `json:"store_interval"`
	TrustedSubnet string `json:"trusted_subnet"`
	Restore       bool   `json:"restore"`
	port          int64
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func loadServerConfig() (*configParams, error) {
	// Флаги
	const (
		defaultStoreInterval = "300s"
		defaultAddr          = "localhost:8080"
		defaultFileStorePath = "store.json"
	)
	configPath := flag.String("c", "", "Path to configuration file")
	address := flag.String("a", defaultAddr, "The address to listen on for HTTP requests.")
	restore := flag.Bool("r", true, "Restore from file")
	storeInterval := flag.String("i", defaultStoreInterval, "Store interval")
	fileStorePath := flag.String("f", defaultFileStorePath, "File store path")
	databaseDsn := flag.String("d", "", "Database dsn")
	hashKey := flag.String("k", "", "Hash key")
	cryptoKey := flag.String("s", "", "Crypto key")
	trustedSubnet := flag.String("t", "", "Trusted subnet")
	flag.Parse()

	// Переменные окружения
	envConfigPath := os.Getenv("CONFIG")
	envAddress := os.Getenv("ADDRESS")
	envRestore := os.Getenv("RESTORE")
	envStoreInterval := os.Getenv("STORE_INTERVAL")
	envFileStorePath := os.Getenv("FILE_STORE_PATH")
	envDatabaseDsn := os.Getenv("DATABASE_DSN")
	envHashKey := os.Getenv("KEY")
	envCryptoKey := os.Getenv("CRYPTO_KEY")
	envTrustedSubnet := os.Getenv("TRUSTED_SUBNET")

	// Проверка наличия конфигурационного файла
	var config = &configParams{}
	if *configPath != "" || envConfigPath != "" {
		path := *configPath
		if path == "" {
			path = envConfigPath
		}

		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("failed to close config file: %v", err)
			}
		}()

		if err := json.NewDecoder(file).Decode(&config); err != nil {
			return nil, fmt.Errorf("failed to decode config file: %w", err)
		}
	}

	if *address != "" {
		config.Addr = *address
	}

	if envAddress != "" {
		config.Addr = envAddress
	}

	if *restore {
		config.Restore = *restore
	}

	if envRestore == "true" {
		config.Restore = true
	}

	if *storeInterval != "" {
		config.StoreInterval = *storeInterval
	}

	if envStoreInterval != "" {
		config.StoreInterval = envStoreInterval
	}

	if *fileStorePath != "" {
		config.FileStorePath = *fileStorePath
	}

	if envFileStorePath != "" {
		config.FileStorePath = envFileStorePath
	}

	if *databaseDsn != "" {
		config.DatabaseDsn = *databaseDsn
	}

	if envDatabaseDsn != "" {
		config.DatabaseDsn = envDatabaseDsn
	}

	if *hashKey != "" {
		config.HashKey = *hashKey
	}

	if envHashKey != "" {
		config.HashKey = envHashKey
	}

	if *cryptoKey != "" {
		config.CryptoKey = *cryptoKey
	}

	if envCryptoKey != "" {
		config.CryptoKey = envCryptoKey
	}

	if _, err := strconv.Atoi(config.StoreInterval); err == nil {
		config.StoreInterval += "s"
	}

	portService := func() int64 {
		split := strings.Split(config.Addr, ":")
		const splitLen = 2

		if len(split) != splitLen {
			log.Fatalf("can't parse address: %s", config.Addr)
		}

		port, newErr := strconv.ParseInt(split[1], 10, 64)
		if newErr != nil {
			log.Fatalf("can't parse port: %v", newErr)
		}

		return port
	}()

	config.port = portService

	if *trustedSubnet != "" {
		config.TrustedSubnet = *trustedSubnet
	}

	if envTrustedSubnet != "" {
		config.TrustedSubnet = envTrustedSubnet
	}

	return config, nil
}

func buildInfo() {
	log.Printf("Version: %s\n", buildVersion)
	log.Printf("Date: %s\n", buildDate)
	log.Printf("Commit: %s\n", buildCommit)
}

func main() {
	serverConfig, err := loadServerConfig()
	if err != nil {
		log.Fatalf("can't load config: %v", err)
	}

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

	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancelFunc()

	go func() {
		sig := <-ctx.Done()
		log.Printf("Received signal: %s", sig)
		cancelFunc()
	}()

	stop := make(chan struct{})

	go run(ctx, &config{
		port:          serverConfig.port,
		storeInterval: serverConfig.StoreInterval,
		fileStorePath: serverConfig.FileStorePath,
		restore:       serverConfig.Restore,
		logger:        *logger.Sugar(),
		databaseDsn:   serverConfig.DatabaseDsn,
		hashKey:       serverConfig.HashKey,
		cryptoKey:     serverConfig.CryptoKey,
		trustedSubnet: serverConfig.TrustedSubnet,
	}, stop)

	<-stop
}
