package store

import (
	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
)

// Config инициализация конфигурации для хранилища.
type Config struct {
	File *file.Config
	DB   *db.Config
}
