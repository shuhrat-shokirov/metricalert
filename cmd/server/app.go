package main

import (
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

	api := rest.NewServerApi(newStore)

	return api.Run()
}
