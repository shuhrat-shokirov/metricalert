package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type configParams struct {
	Addr string `env:"ADDRESS"`
}

func main() {
	var defaultParams configParams

	err := env.Parse(&defaultParams)
	if err != nil {
		log.Printf("can't parse env: %v\n", err)
		os.Exit(1)
	}

	serverAddr := flag.String("a", "localhost:8080", "server address")

	flag.Parse()

	if defaultParams.Addr != "" {
		serverAddr = &defaultParams.Addr
	}

	portService := func() int64 {
		split := strings.Split(*serverAddr, ":")
		if len(split) != 2 {
			log.Printf("Invalid address: %s\n", *serverAddr)
			os.Exit(1)
		}
		port, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			log.Printf("Invalid port: %s\n", split[1])
			os.Exit(1)
		}

		return port
	}()

	// Проверка на неизвестные флаги
	flag.VisitAll(func(f *flag.Flag) {
		if !flag.Parsed() {
			log.Printf("Unknown flag: %s\n", f.Name)
			os.Exit(1)
		}
	})

	if err := run(portService); err != nil {
		log.Printf("can't run server: %v", err)
		os.Exit(1)
	}
}
