package services

import "metricalert/internal/core/model"

type Store interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGaugeList() map[string]string
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}
type Repo struct {
	store Store
}

func NewRepo(store Store) *Repo {
	return &Repo{
		store: store,
	}
}

func (r *Repo) UpdateGauge(name string, value float64) error {
	r.store.UpdateGauge(name, value)

	return nil
}

func (r *Repo) UpdateCounter(name string, value int64) error {
	r.store.UpdateCounter(name, value)

	return nil
}

func (r *Repo) GetGauge(name string) (float64, error) {
	value, ok := r.store.GetGauge(name)
	if !ok {
		return 0, model.ErrorNotFound
	}

	return value, nil
}

func (r *Repo) GetCounter(name string) (int64, error) {
	value, ok := r.store.GetCounter(name)
	if !ok {
		return 0, model.ErrorNotFound
	}

	return value, nil
}

func (r *Repo) GetGaugeList() map[string]string {
	return r.store.GetGaugeList()
}
