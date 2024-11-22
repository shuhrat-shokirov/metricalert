package file

import "metricalert/internal/server/infra/store/memory"

type Config struct {
	MemoryStore   *memory.Config
	FilePath      string
	StoreInterval int
	Restore       bool
}
