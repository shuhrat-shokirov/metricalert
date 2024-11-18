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

type config struct {
	logger        zap.SugaredLogger
	fileStorePath string
	port          int64
	storeInterval int
	restore       bool
}

func run(conf config) error {
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("can't create store: %w", err)
	}

	newApplication := application.NewApplication(newStore)

	if conf.restore {
		err = newApplication.LoadMetricsFromFile(conf.fileStorePath)
		if err != nil {
			return fmt.Errorf("can't load metrics from file: %w", err)
		}
	}

	api := rest.NewServerAPI(newApplication, conf.port, conf.logger)

	if conf.storeInterval > 0 {
		ticker := time.NewTicker(time.Duration(conf.storeInterval) * time.Second)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				err := newApplication.SaveMetricsToFile(conf.fileStorePath)
				if err != nil {
					conf.logger.Errorf("can't save metrics to file: %v", err)
				}
			}
		}()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		<-stop
		err := newApplication.SaveMetricsToFile(conf.fileStorePath)
		if err != nil {
			conf.logger.Errorf("can't save metrics to file: %v", err)
		}
		os.Exit(0)
	}()

	if err := api.Run(); err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}

	return nil
}
