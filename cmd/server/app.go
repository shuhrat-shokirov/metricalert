package main

import (
	"fmt"

	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

func run(port int64, logger zap.SugaredLogger) error {
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("can't create store: %w", err)
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication, port, logger)

	return api.Run()
}
