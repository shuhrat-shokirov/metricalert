package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"metricalert/internal/server/core/repositories"
)

func TestMemStorage_UpdateGauge1(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "first test",
			args: args{
				name:  "test",
				value: 1.1,
			},
			want: 1.1,
		},
		{
			name: "second test",
			args: args{
				name:  "test",
				value: 0.9,
			},
			want: 0.9,
		},
		{
			name: "third test",
			args: args{
				name:  "test",
				value: -1,
			},
			want: -1,
		},
	}

	s := NewStore(&Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateGauge(context.Background(), tt.args.name, tt.args.value)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, s.gauges[tt.args.name])
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "first test",
			args: args{
				name:  "test",
				value: 1,
			},
			want: 1,
		},
		{
			name: "second test",
			args: args{
				name:  "test",
				value: 1000012,
			},
			want: 1000013,
		},
		{
			name: "third test",
			args: args{
				name:  "test",
				value: -1,
			},
			want: 1000012,
		},
		{
			name: "invalid test",
			args: args{
				name:  "test",
				value: 10001,
			},
			wantErr: true,
			want:    1000012,
		},
	}

	s := NewStore(&Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateCounter(context.Background(), tt.args.name, tt.args.value)
			assert.Nil(t, err)

			if tt.wantErr {
				assert.NotEqual(t, tt.want, s.counters[tt.args.name])
				return
			}

			assert.Equal(t, tt.want, s.counters[tt.args.name])
		})
	}
}

func BenchmarkMemStorage_UpdateGauge1(b *testing.B) {
	s := NewStore(&Config{})

	for range b.N {
		_ = s.UpdateGauge(context.Background(), "test", 1.1)
	}
}

func ExampleNewStore() {
	_ = NewStore(&Config{})
	// Output:
}

func ExampleStore_UpdateGauge() {
	s := NewStore(&Config{}) // Инициализация хранилища

	_ = s.UpdateGauge(context.Background(), "test", 1.1)

	// Output:
}

func ExampleStore_UpdateGauges() {
	s := NewStore(&Config{}) // Инициализация хранилища

	_ = s.UpdateGauges(context.Background(), map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})

	// Output:
}

func ExampleStore_UpdateCounter() {
	s := NewStore(&Config{}) // Инициализация хранилища

	_ = s.UpdateCounter(context.Background(), "test", 1)

	// Output:
}

func ExampleStore_UpdateCounters() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Обновление значений метрик
	_ = s.UpdateCounters(context.Background(), map[string]int64{
		"test1": 1,
		"test2": 2,
	})
	// Output:
}

func ExampleStore_GetGauge() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Получение значения метрики
	_, _ = s.GetGauge(context.Background(), "test")

	// Output:
}

func TestStore_GetGauge(t *testing.T) {
	s := NewStore(&Config{})

	_, err := s.GetGauge(context.Background(), "test")
	assert.NotNil(t, err)
	assert.EqualError(t, err, repositories.ErrNotFound.Error())

	s.gauges["test"] = 1.1

	value, err := s.GetGauge(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, 1.1, value)
}

func ExampleStore_GetCounter() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Получение значения метрики
	_, _ = s.GetCounter(context.Background(), "test")

	// Output:
}

func TestStore_GetCounter(t *testing.T) {
	s := NewStore(&Config{})

	_, err := s.GetCounter(context.Background(), "test")
	assert.NotNil(t, err)
	assert.EqualError(t, err, repositories.ErrNotFound.Error())

	s.counters["test"] = 1

	value, err := s.GetCounter(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), value)
}

func ExampleStore_GetGaugeList() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Получение списка метрик
	_, _ = s.GetGaugeList(context.Background())

	// Output:
}

func TestStore_GetGaugeList(t *testing.T) {
	s := NewStore(&Config{})

	list, err := s.GetGaugeList(context.Background())
	assert.Nil(t, err)
	assert.Empty(t, list)

	s.gauges["test1"] = 1.1
	s.gauges["test2"] = 2.2

	list, err = s.GetGaugeList(context.Background())
	assert.Nil(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, 1.1, list["test1"])
	assert.Equal(t, 2.2, list["test2"])
}

func ExampleStore_GetCounterList() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Получение списка метрик
	_, _ = s.GetCounterList(context.Background())

	// Output:
}

func TestStore_GetCounterList(t *testing.T) {
	s := NewStore(&Config{})

	list, err := s.GetCounterList(context.Background())
	assert.Nil(t, err)
	assert.Empty(t, list)

	s.counters["test1"] = 1
	s.counters["test2"] = 2

	list, err = s.GetCounterList(context.Background())
	assert.Nil(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, int64(1), list["test1"])
	assert.Equal(t, int64(2), list["test2"])
}

func ExampleStore_RestoreGauges() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Восстановление значений метрик
	s.RestoreGauges(map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})

	// Output:
}

func TestStore_RestoreGauges(t *testing.T) {
	s := NewStore(&Config{})

	s.RestoreGauges(map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})

	assert.Equal(t, 1.1, s.gauges["test1"])
	assert.Equal(t, 2.2, s.gauges["test2"])
}

func ExampleStore_RestoreCounters() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Восстановление значений метрик
	s.RestoreCounters(map[string]int64{
		"test1": 1,
		"test2": 2,
	})

	// Output:
}

func TestStore_RestoreCounters(t *testing.T) {
	s := NewStore(&Config{})

	s.RestoreCounters(map[string]int64{
		"test1": 1,
		"test2": 2,
	})

	assert.Equal(t, int64(1), s.counters["test1"])
	assert.Equal(t, int64(2), s.counters["test2"])
}

func ExampleStore_Ping() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Проверка доступности хранилища
	_ = s.Ping(context.Background())

	// Output:
}

func ExampleStore_Close() {
	s := NewStore(&Config{}) // Инициализация хранилища

	// Закрытие хранилища
	_ = s.Close()

	// Output:
}
