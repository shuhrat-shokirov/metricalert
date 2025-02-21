package file

import "metricalert/internal/server/infra/store/memory"

// Config Параметры конфигурации для хранилища в файле.
type Config struct {
	MemoryStore   *memory.Config
	FilePath      string
	StoreInterval int
	Restore       bool
}
