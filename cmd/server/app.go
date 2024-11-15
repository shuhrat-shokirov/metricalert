package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

func run(port int64, logger zap.SugaredLogger, storeInterval int, fileStorePath string, restore bool) error {
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("can't create store: %w", err)
	}

	newApplication := application.NewApplication(newStore)

	if restore {
		err = newApplication.LoadMetricsFromFile(fileStorePath)
		if err != nil {
			return fmt.Errorf("can't load metrics from file: %w", err)
		}
	}

	api := rest.NewServerAPI(newApplication, port, logger)

	if storeInterval > 0 {
		ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				err := newApplication.SaveMetricsToFile(fileStorePath)
				if err != nil {
					logger.Errorf("can't save metrics to file: %v", err)
				}
			}
		}()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop
		err := newApplication.SaveMetricsToFile(fileStorePath)
		if err != nil {
			logger.Errorf("can't save metrics to file: %v", err)
		}
		os.Exit(0)
	}()

	return api.Run()
}
