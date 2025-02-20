package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"metricalert/internal/server/infra/store/db"
	"metricalert/internal/server/infra/store/file"
)

func ExampleNewStore() {
	_, _ = NewStore(Config{
		File: &file.Config{
			StoreInterval: 1,
			Restore:       true,
			FilePath:      "test",
		},
		DB: nil,
	})
	// Output:
}

func TestNewStore(t *testing.T) {
	{
		store, err := NewStore(Config{
			File: nil,
			DB:   nil,
		})
		assert.Nil(t, store)
		assert.NotNil(t, err)
	}

	{
		store, err := NewStore(Config{
			File: &file.Config{
				StoreInterval: 1,
				Restore:       false,
			},
		})

		assert.Nil(t, err)
		assert.NotNil(t, store)
	}

	{
		store, err := NewStore(Config{
			File: &file.Config{
				StoreInterval: 1,
				Restore:       true,
			},
		})

		assert.Nil(t, store)
		assert.NotNil(t, err)
	}

	{
		store, err := NewStore(Config{
			File: &file.Config{
				StoreInterval: 1,
				Restore:       true,
				FilePath:      "test",
			},
		})
		defer func() {
			_ = store.Close()
			_ = os.Remove("test")
		}()

		assert.NotNil(t, store)
		assert.Nil(t, err)
	}

	{
		store, err := NewStore(Config{
			DB: &db.Config{
				DSN: "test",
			},
		})

		assert.Nil(t, store)
		assert.NotNil(t, err)
	}
}
