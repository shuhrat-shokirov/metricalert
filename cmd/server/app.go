package main

import (
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
	"metricalert/internal/server/infra/store/memory"
)

type config struct {
	logger        zap.SugaredLogger
	fileStorePath string
	databaseDsn   string
	hashKey       string
	port          int64
	storeInterval int
	restore       bool
}

func run(conf *config) error {
	var (
		newStore store.Store
		err      error
		dbConfig *db.Config
	)

	fileConfig := &file.Config{
		StoreInterval: conf.storeInterval,
		Restore:       conf.restore,
		FilePath:      conf.fileStorePath,
		MemoryStore:   &memory.Config{},
	}

	if conf.databaseDsn != "" {
		dbConfig = &db.Config{
			DSN: conf.databaseDsn,
		}
	}

	newStore, err = store.NewStore(store.Config{
		File: fileConfig,
		DB:   dbConfig,
	})
	if err != nil {
		return fmt.Errorf("can't create store: %w", err)
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(rest.Config{
		Server:  newApplication,
		Port:    conf.port,
		Logger:  conf.logger,
		HashKey: conf.hashKey,
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop
		if err := newStore.Close(); err != nil {
			conf.logger.Errorf("can't close store: %v", err)
		}
	}()

	if err := api.Run(); err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}

	return nil
}
