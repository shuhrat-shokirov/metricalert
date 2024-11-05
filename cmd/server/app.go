package main

import (
	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	store2 "metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

func run() error {

	newStore, err := store2.NewStore(store2.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return err
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication)

	return api.Run()
}
