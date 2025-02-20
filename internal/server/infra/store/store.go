// Package store реализует интерфейс для работы с метриками.
// Пакет содержит функцию NewStore, которая создает новый экземпляр Store.
// В зависимости от конфигурации создается экземпляр хранилища в памяти, файловое хранилище или база данных.
// В случае, если конфигурация не передана, возвращается ошибка.
//
// В данном этапе реализована работа с базой данных, файловым хранилищем и хранилищем в памяти.
// В случае, если передана конфигурация для базы данных, создается экземпляр хранилища db.Store.
// В случае, если передана конфигурация для файлового хранилища, создается экземпляр хранилища file.Store.
// В случае, если конфигурация не передана, создается экземпляр хранилища в памяти memory.Store.
// В случае, если конфигурация не передана, возвращается ошибка.
//
// Для работы с porstgresql передается конфигурация db.Config.
//
//	dbConfig = &db.Config{
//		DSN: conf.databaseDsn,
//	}
//
// Для работы с файловым хранилищем передается конфигурация file.Config.
//
//	fileConfig := &file.Config{
//			StoreInterval: conf.storeInterval,
//			Restore:       conf.restore,
//			FilePath:      conf.fileStorePath,
//			MemoryStore:   &memory.Config{},
//		}
//
// Если Restore = true, то используется файловое хранилище, иначе хранилище в памяти.
//
// В случае, если конфигурация не передана, возвращается ошибка.
package store

import (
	"context"
	"fmt"

	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
	"metricalert/internal/server/infra/store/memory"
)

// Store интерфейс для работы с метриками.
type Store interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateGauges(ctx context.Context, gauges map[string]float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	UpdateCounters(ctx context.Context, counters map[string]int64) error
	GetGaugeList(context.Context) (map[string]float64, error)
	GetCounterList(context.Context) (map[string]int64, error)
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	Close() error
	Ping(ctx context.Context) error
}

// NewStore создает новый экземпляр Store.
// В зависимости от конфигурации создается экземпляр хранилища в памяти, файловое хранилище или база данных.
// В случае, если конфигурация не передана, возвращается ошибка.
func NewStore(conf Config) (Store, error) {
	switch {
	case conf.DB != nil:
		store, err := db.New(conf.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("can't create db store: %w", err)
		}

		return store, nil
	case conf.File != nil:
		if !conf.File.Restore {
			return memory.NewStore(conf.File.MemoryStore), nil
		}

		store, err := file.NewStore(conf.File)
		if err != nil {
			return nil, fmt.Errorf("can't create file store: %w", err)
		}

		return store, nil
	default:
		return nil, fmt.Errorf("unknown store type, config: %+v", conf)
	}
}
