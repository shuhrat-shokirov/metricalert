package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type configParams struct {
	Addr           string `json:"address"`
	HashKey        string `json:"-"`
	CryptoKey      string `json:"crypto_key"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	RateLimit      int64  `json:"-"`
}

func loadAgentConfig() (*configParams, error) {
	// Флаги
	const (
		defaultReportInterval = "10s"
		defaultPollInterval   = "2s"
		defaultAddr           = "localhost:8080"
	)
	serverAddr := flag.String("a", defaultAddr, "server address")
	report := flag.String("r", defaultReportInterval, "report interval")
	poll := flag.String("p", defaultPollInterval, "poll interval")
	hashKey := flag.String("k", "", "hash key")
	rateLimit := flag.Int64("l", 0, "rate limit")
	cryptoKey := flag.String("s", "", "crypto key")
	configPath := flag.String("c", "", "Path to configuration file")
	flag.Parse()

	// Переменные окружения
	envConfigPath := os.Getenv("CONFIG")
	envAddress := os.Getenv("ADDRESS")
	envHashKey := os.Getenv("HASH_KEY")
	envCryptoKey := os.Getenv("CRYPTO_KEY")
	envReportInterval := os.Getenv("REPORT_INTERVAL")
	envPollInterval := os.Getenv("POLL_INTERVAL")
	envRateLimit := os.Getenv("RATE_LIMIT")

	// Проверка наличия конфигурационного файла
	var config = &configParams{}
	if *configPath != "" || envConfigPath != "" {
		path := *configPath
		if path == "" {
			path = envConfigPath
		}

		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open configuration file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("failed to close configuration file: %v", err)
			}
		}()

		if err := json.NewDecoder(file).Decode(&config); err != nil {
			return nil, fmt.Errorf("failed to decode configuration file: %w", err)
		}
	}

	if *serverAddr != "" {
		config.Addr = *serverAddr
	} else if envAddress != "" {
		config.Addr = envAddress
	}

	if *report != "" {
		config.ReportInterval = *report
	} else if envReportInterval != "" {
		config.ReportInterval = envReportInterval
	}

	if *poll != "" {
		config.PollInterval = *poll
	} else if envPollInterval != "" {
		config.PollInterval = envPollInterval
	}

	if *hashKey != "" {
		config.HashKey = *hashKey
	} else if envHashKey != "" {
		config.HashKey = envHashKey
	}

	if *rateLimit != 0 {
		config.RateLimit = *rateLimit
	} else if envRateLimit != "" {
		var err error
		config.RateLimit, err = strconv.ParseInt(envRateLimit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rate limit: %w", err)
		}
	}

	if *cryptoKey != "" {
		config.CryptoKey = *cryptoKey
	} else if envCryptoKey != "" {
		config.CryptoKey = envCryptoKey
	}

	return config, nil
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
	agentConfig, err := loadAgentConfig()
	if err != nil {
		log.Fatalf("failed to load agent config: %v", err)
	}

	buildInfo()

	if err = run(config{
		addr:           agentConfig.Addr,
		reportInterval: agentConfig.ReportInterval,
		pollInterval:   agentConfig.PollInterval,
		hashKey:        agentConfig.HashKey,
		rateLimit:      agentConfig.RateLimit,
		cryptoKey:      agentConfig.CryptoKey,
	}); err != nil {
		log.Fatalf("failed to run agent: %v", err)
	}
}
