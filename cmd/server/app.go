package main

import (
	"metricalert/internal/core/application"
	"metricalert/internal/core/services"
	"metricalert/internal/infra/api/rest"
	"metricalert/internal/infra/store"
	"metricalert/internal/infra/store/memory"
)

func run() error {

	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return err
	}

	repo := services.NewRepo(newStore)

	newApplication := application.NewApplication(repo)

	api := rest.NewServerAPI(newApplication)

	return api.Run()
}
