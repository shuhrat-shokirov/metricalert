package store

import (
	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
)

type Config struct {
	File *file.Config
	DB   *db.Config
}
