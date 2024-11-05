package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/infra/api/rest"
	"metricalert/internal/server/infra/store"
	"metricalert/internal/server/infra/store/memory"
)

type Network struct {
	Host string
	Port int
}

func (n *Network) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func (n *Network) Set(value string) error {
	split := strings.Split(value, ":")
	if len(split) != 2 {
		return fmt.Errorf("invalid format")
	}

	n.Host = split[0]

	atoi, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	n.Port = atoi

	return nil
}

var port int64 = 8080

func init() {

	addr := new(Network)

	flag.Var(addr, "a", "server address")

	flag.Parse()

	if addr.Port != 0 {
		port = int64(addr.Port)
	}

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})
}

func run() error {
	newStore, err := store.NewStore(store.Config{
		Memory: &memory.Config{},
	})
	if err != nil {
		return err
	}

	newApplication := application.NewApplication(newStore)

	api := rest.NewServerAPI(newApplication, port)

	return api.Run()
}
