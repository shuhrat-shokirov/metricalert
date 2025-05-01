package main

import (
	"context"
	"time"

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
	cryptoKey     string
	storeInterval string
	trustedSubnet string
	port          int64
	restore       bool
}

func run(ctx context.Context, conf *config, stop chan<- struct{}) {
	var (
		newStore store.Store
		err      error
		dbConfig *db.Config
	)

	storeInterval, err := time.ParseDuration(conf.storeInterval)
	if err != nil {
		conf.logger.Fatalf("failed to parse store interval: %v", err)
	}

	fileConfig := &file.Config{
		StoreInterval: storeInterval,
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
		conf.logger.Fatalf("failed to create store: %v", err)
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(&rest.Config{
		Server:        newApplication,
		Port:          conf.port,
		Logger:        conf.logger,
		HashKey:       conf.hashKey,
		CryptoKey:     conf.cryptoKey,
		TrustedSubnet: conf.trustedSubnet,
	})

	go func() {
		newStore.Sync(ctx)
	}()

	go func() {
		<-ctx.Done()
		if err := api.Shutdown(context.Background()); err != nil {
			conf.logger.Errorw("can't shutdown server", "error", err)
		}

		conf.logger.Info("server shutdown")

		if err := newStore.Close(); err != nil {
			conf.logger.Errorw("can't close store", "error", err)
		}
		conf.logger.Info("store closed")

		stop <- struct{}{}
	}()

	if err := api.Run(); err != nil {
		conf.logger.Fatalw("failed to run server", "error", err)
	}
}
