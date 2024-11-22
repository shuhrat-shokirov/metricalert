package store

import (
	"metricalert/internal/server/infra/store/file"
)

type Config struct {
	File *file.Config
}
