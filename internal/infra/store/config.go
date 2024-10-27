package store

import "metricalert/internal/infra/store/memory"

type Config struct {
	Memory *memory.Config
}
