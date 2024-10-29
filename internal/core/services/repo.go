package services

type Store interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
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
