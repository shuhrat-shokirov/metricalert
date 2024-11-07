package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

func init() {
}

func run(port int64) error {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			panic(err)
		}
	}()
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return fmt.Errorf("can't create store: %w", err)
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication, port, *logger.Sugar())

	return api.Run()
}
