//nolint:dupl,nolintlint,gocritic
package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ExampleNew() {
	_, _ = New("dns")
	// Output:
}

func TestNew(t *testing.T) {
	store, err := New("dns")
	assert.Nil(t, store)
	assert.NotNil(t, err)
}

func TestStore_UpdateGauge(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Exec", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 1"), nil)

	store := &Store{pool: mockPool}

	err := store.UpdateGauge(context.Background(), "test", 1.1)
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
}

func TestStore_UpdateGauges(t *testing.T) {
	mockPool := new(MockPool)
	mockBatchResults := new(MockBatchResults)

	mockPool.On("SendBatch", mock.Anything, mock.Anything).Return(mockBatchResults, nil)
	mockBatchResults.On("Close").Return(nil)
	mockBatchResults.On("Exec").Return(pgconn.NewCommandTag("INSERT 1"), nil)

	store := &Store{pool: mockPool}

	err := store.UpdateGauges(context.Background(), map[string]float64{
		"test1": 1.1,
		"test2": 2.2,
	})
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockBatchResults.AssertExpectations(t)
}

func TestStore_UpdateCounter(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Exec", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(pgconn.NewCommandTag("INSERT 1"), nil)

	store := &Store{pool: mockPool}

	err := store.UpdateCounter(context.Background(), "test", 1)
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
}

func TestStore_UpdateCounters(t *testing.T) {
	mockPool := new(MockPool)
	mockBatchResults := new(MockBatchResults)

	mockPool.On("SendBatch", mock.Anything, mock.Anything).Return(mockBatchResults, nil)
	mockBatchResults.On("Close").Return(nil)
	mockBatchResults.On("Exec").Return(pgconn.NewCommandTag("INSERT 1"), nil)

	store := &Store{pool: mockPool}

	err := store.UpdateCounters(context.Background(), map[string]int64{
		"test1": 1,
		"test2": 2,
	})
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockBatchResults.AssertExpectations(t)
}

func TestStore_GetGauge(t *testing.T) {
	mockPool := new(MockPool)
	mockRow := new(MockRow)

	mockPool.On("QueryRow", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil)

	store := &Store{pool: mockPool}

	_, err := store.GetGauge(context.Background(), "test")
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestStore_GetGaugeError(t *testing.T) {
	mockPool := new(MockPool)
	mockRow := new(MockRow)

	mockPool.On("QueryRow", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(assert.AnError)

	store := &Store{pool: mockPool}

	_, err := store.GetGauge(context.Background(), "test")
	assert.NotNil(t, err)

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestStore_GetCounter(t *testing.T) {
	mockPool := new(MockPool)
	mockRow := new(MockRow)

	mockPool.On("QueryRow", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil)

	store := &Store{pool: mockPool}

	_, err := store.GetCounter(context.Background(), "test")
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestStore_GetCounterError(t *testing.T) {
	mockPool := new(MockPool)
	mockRow := new(MockRow)

	mockPool.On("QueryRow", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(assert.AnError)

	store := &Store{pool: mockPool}

	_, err := store.GetCounter(context.Background(), "test")
	assert.NotNil(t, err)

	mockPool.AssertExpectations(t)
	mockRow.AssertExpectations(t)
}

func TestStore_Close(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Close").Return(nil)

	store := &Store{pool: mockPool}

	err := store.Close()
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
}

func TestStore_GetGaugeList(t *testing.T) {
	mockPool := new(MockPool)
	mockRows := new(MockRow)

	mockPool.On("Query", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRows, nil)
	mockRows.On("Close").Return(nil)
	mockRows.On("Next").Return(true).Once()
	mockRows.On("Next").Return(false).Once()
	mockRows.On("Scan", mock.Anything, mock.Anything).Return(nil)

	store := &Store{pool: mockPool}

	_, err := store.GetGaugeList(context.Background())
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestStore_GetGaugeListError(t *testing.T) {
	mockPool := new(MockPool)
	mockRows := new(MockRow)

	mockPool.On("Query", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRows, assert.AnError)

	store := &Store{pool: mockPool}

	_, err := store.GetGaugeList(context.Background())
	assert.NotNil(t, err)

	mockPool.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestStore_GetCounterList(t *testing.T) {
	mockPool := new(MockPool)
	mockRows := new(MockRow)

	mockPool.On("Query", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRows, nil)
	mockRows.On("Close").Return(nil)
	mockRows.On("Next").Return(true).Once()
	mockRows.On("Next").Return(false).Once()
	mockRows.On("Scan", mock.Anything, mock.Anything).Return(nil)

	store := &Store{pool: mockPool}

	_, err := store.GetCounterList(context.Background())
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestStore_GetCounterListError(t *testing.T) {
	mockPool := new(MockPool)
	mockRows := new(MockRow)

	mockPool.On("Query", mock.Anything, mock.Anything,
		mock.Anything).Return(mockRows, assert.AnError)

	store := &Store{pool: mockPool}

	_, err := store.GetCounterList(context.Background())
	assert.NotNil(t, err)

	mockPool.AssertExpectations(t)
	mockRows.AssertExpectations(t)
}

func TestStore_Ping(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Ping", mock.Anything).Return(nil)

	store := &Store{pool: mockPool}

	err := store.Ping(context.Background())
	assert.Nil(t, err)

	mockPool.AssertExpectations(t)
}

func TestStore_PingError(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Ping", mock.Anything).Return(assert.AnError)

	store := &Store{pool: mockPool}

	err := store.Ping(context.Background())
	assert.NotNil(t, err)

	mockPool.AssertExpectations(t)
}
