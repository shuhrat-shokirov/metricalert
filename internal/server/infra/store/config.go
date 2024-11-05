package store

import (
	"metricalert/internal/server/infra/store/memory"
)

type Config struct {
	Memory *memory.Config
}
