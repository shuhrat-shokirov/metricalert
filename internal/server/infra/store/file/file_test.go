package file

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"metricalert/internal/server/core/repositories"
	"metricalert/internal/server/infra/store/memory"
)

func ExampleNewStore() {
	_, _ = NewStore(&Config{})

	// Output:
}

func TestNewStore(t *testing.T) {
	{
		store, err := NewStore(&Config{})
		assert.Nil(t, store)
		assert.Error(t, err)
	}

	{
		store, err := NewStore(&Config{
			MemoryStore: &memory.Config{},
		})
		assert.Nil(t, store)
		assert.Error(t, err)
	}

	{
		store, err := NewStore(&Config{
			MemoryStore:   &memory.Config{},
			FilePath:      "test",
			StoreInterval: 1,
		})
		defer func() {
			err := store.Close()
			assert.NoError(t, err)
		}()
		defer func() {
			err := os.Remove("test")
			assert.NoError(t, err)
		}()
		assert.NotNil(t, store)
		assert.NoError(t, err)
	}

	{
		store, fileName := testInitModule()
		defer testDone(fileName)
		defer func() {
			_ = store.Close()
		}()

		assert.NotNil(t, store)

		err := store.UpdateCounters(context.Background(), map[string]int64{
			"test1": 1,
			"test2": 2,
		})
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)

		store2, err := NewStore(&Config{
			MemoryStore:   &memory.Config{},
			FilePath:      fileName,
			StoreInterval: 1,
		})
		defer func() {
			err := store2.Close()
			assert.NoError(t, err)
		}()

		assert.NotNil(t, store2)
		assert.NoError(t, err)

	}

}

func testInitModule() (*Store, string) {
	fileName := uuid.New().String()
	store, err := NewStore(&Config{
		MemoryStore:   &memory.Config{},
		FilePath:      fileName,
		StoreInterval: 1,
	})
	if err != nil {
		panic(err)
	}
	return store, fileName
}

func testDone(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		panic(err)
	}
}

func ExampleStore_UpdateGauge() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Обновление значения метрики
	_ = store.UpdateGauge(context.Background(), "test", 1.1)

	// Output:
}

func TestStore_UpdateGauge(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		err := store.UpdateGauge(context.Background(), "test", 1.1)
		assert.Nil(t, err)

		gauge, err := store.Store.GetGauge(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, 1.1, gauge)
	}

	{
		err := store.UpdateGauge(context.Background(), "test", 2.2)
		assert.Nil(t, err)

		gauge, err := store.Store.GetGauge(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, 2.2, gauge)
	}
}

func ExampleStore_UpdateGauges() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Обновление значений метрик
	_ = store.UpdateGauges(context.Background(), map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})

	// Output:
}

func TestStore_UpdateGauges(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		err := store.UpdateGauges(context.Background(), map[string]float64{
			"test1": 1.1,
			"test2": 2.2,
		})
		assert.Nil(t, err)

		gauge1, err := store.Store.GetGauge(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, 1.1, gauge1)

		gauge2, err := store.Store.GetGauge(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, 2.2, gauge2)
	}

	{
		err := store.UpdateGauges(context.Background(), map[string]float64{
			"test1": 3.3,
			"test2": 4.4,
		})
		assert.Nil(t, err)

		gauge1, err := store.Store.GetGauge(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, 3.3, gauge1)

		gauge2, err := store.Store.GetGauge(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, 4.4, gauge2)
	}
}

func ExampleStore_UpdateCounter() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Обновление значения метрики
	_ = store.UpdateCounter(context.Background(), "test", 1)

	// Output:
}

func TestStore_UpdateCounter(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		err := store.UpdateCounter(context.Background(), "test", 1)
		assert.Nil(t, err)

		counter, err := store.Store.GetCounter(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, int64(1), counter)
	}

	{
		err := store.UpdateCounter(context.Background(), "test", 2)
		assert.Nil(t, err)

		counter, err := store.Store.GetCounter(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, int64(3), counter)
	}
}

func ExampleStore_UpdateCounters() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Обновление значений метрик
	_ = store.UpdateCounters(context.Background(), map[string]int64{
		"test1": 1,
		"test2": 2,
	})

	// Output:
}

func TestStore_UpdateCounters(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		err := store.UpdateCounters(context.Background(), map[string]int64{
			"test1": 1,
			"test2": 2,
		})
		assert.Nil(t, err)

		counter1, err := store.Store.GetCounter(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, int64(1), counter1)

		counter2, err := store.Store.GetCounter(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, int64(2), counter2)
	}

	{
		err := store.UpdateCounters(context.Background(), map[string]int64{
			"test1": 3,
			"test2": 4,
		})
		assert.Nil(t, err)

		counter1, err := store.Store.GetCounter(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, int64(4), counter1)

		counter2, err := store.Store.GetCounter(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, int64(6), counter2)
	}
}

func ExampleStore_GetGauge() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Получение значения метрики
	_, _ = store.GetGauge(context.Background(), "test")

	// Output:
}

func TestStore_GetGauge(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	_, err := store.GetGauge(context.Background(), "test")
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Errorf("can't get gauge: %w", repositories.ErrNotFound).Error())

	{
		err := store.UpdateGauge(context.Background(), "test", 1.1)
		assert.Nil(t, err)

		value, err := store.GetGauge(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, 1.1, value)
	}
}

func ExampleStore_GetCounter() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Получение значения метрики
	_, _ = store.GetCounter(context.Background(), "test")

	// Output:
}

func TestStore_GetCounter(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	_, err := store.GetCounter(context.Background(), "test")
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Errorf("can't get counter: %w", repositories.ErrNotFound).Error())

	{
		err := store.UpdateCounter(context.Background(), "test", 1)
		assert.Nil(t, err)

		value, err := store.GetCounter(context.Background(), "test")
		assert.Nil(t, err)

		assert.Equal(t, int64(1), value)
	}
}

func ExampleStore_GetGaugeList() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Получение списка метрик
	_, _ = store.GetGaugeList(context.Background())

	// Output:
}

func TestStore_GetGaugeList(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		list, err := store.GetGaugeList(context.Background())
		assert.Nil(t, err)
		assert.Empty(t, list)
	}

	{
		err := store.UpdateGauge(context.Background(), "test1", 1.1)
		assert.Nil(t, err)

		err = store.UpdateGauge(context.Background(), "test2", 2.2)
		assert.Nil(t, err)

		list, err := store.GetGaugeList(context.Background())
		assert.Nil(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, 1.1, list["test1"])
		assert.Equal(t, 2.2, list["test2"])
	}
}

func ExampleStore_GetCounterList() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Получение списка метрик
	_, _ = store.GetCounterList(context.Background())

	// Output:
}

func TestStore_GetCounterList(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		list, err := store.GetCounterList(context.Background())
		assert.Nil(t, err)
		assert.Empty(t, list)
	}

	{
		err := store.UpdateCounter(context.Background(), "test1", 1)
		assert.Nil(t, err)

		err = store.UpdateCounter(context.Background(), "test2", 2)
		assert.Nil(t, err)

		list, err := store.GetCounterList(context.Background())
		assert.Nil(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int64(1), list["test1"])
		assert.Equal(t, int64(2), list["test2"])
	}
}

func ExampleStore_RestoreGauges() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Восстановление значений метрик
	store.RestoreGauges(map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})

	// Output:
}

func TestStore_RestoreGauges(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		store.RestoreGauges(map[string]float64{
			"test1": 1.1,
			"test2": 2.2,
		})

		gauge1, err := store.Store.GetGauge(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, 1.1, gauge1)

		gauge2, err := store.Store.GetGauge(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, 2.2, gauge2)
	}

	{
		store.RestoreGauges(map[string]float64{
			"test1": 3.3,
			"test2": 4.4,
		})

		gauge1, err := store.Store.GetGauge(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, 3.3, gauge1)

		gauge2, err := store.Store.GetGauge(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, 4.4, gauge2)
	}
}

func ExampleStore_RestoreCounters() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Восстановление значений метрик
	store.RestoreCounters(map[string]int64{
		"test1": 1,
		"test2": 2,
	})

	// Output:
}

func TestStore_RestoreCounters(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	{
		store.RestoreCounters(map[string]int64{
			"test1": 1,
			"test2": 2,
		})

		counter1, err := store.Store.GetCounter(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, int64(1), counter1)

		counter2, err := store.Store.GetCounter(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, int64(2), counter2)
	}

	{
		store.RestoreCounters(map[string]int64{
			"test1": 3,
			"test2": 4,
		})

		counter1, err := store.Store.GetCounter(context.Background(), "test1")
		assert.Nil(t, err)

		assert.Equal(t, int64(3), counter1)

		counter2, err := store.Store.GetCounter(context.Background(), "test2")
		assert.Nil(t, err)

		assert.Equal(t, int64(4), counter2)
	}
}

func ExampleStore_Close() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)

	// Закрытие хранилища
	_ = store.Close()

	// Output:
}

func TestStore_Close(t *testing.T) {
	store, fileName := testInitModule()
	defer testDone(fileName)

	err := store.Close()
	assert.Nil(t, err)
}

func ExampleStore_Ping() {
	// Инициализация хранилища
	store, fileName := testInitModule()
	defer testDone(fileName)
	defer func() {
		_ = store.Close()
	}()

	// Проверка доступности хранилища
	_ = store.Ping(context.Background())

	// Output:
}
