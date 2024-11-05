package main

import (
	"flag"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	store2 "metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

var port int64

func init() {
	flag.Int64Var(&port, "port", 8080, "port to listen on")
	flag.Parse()
}

func run() error {
	newStore, err := store2.NewStore(store2.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return err
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication, port)

	return api.Run()
}
