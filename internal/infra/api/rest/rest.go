package rest

import (
	"net/http"
	"strconv"
	"strings"
)

type ServerService interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

type API struct {
	srv *http.Server
}

func NewServerApi(server ServerService) *API {
	h := handler{
		server: server,
	}

	router := http.NewServeMux()

	router.HandleFunc("/health", h.health)
	router.HandleFunc("/update/", h.update)

	return &API{
		srv: &http.Server{
			Addr:    ":8081",
			Handler: router,
		},
	}
}

func (a *API) Run() error {
	return a.srv.ListenAndServe()
}

type handler struct {
	server ServerService
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid request format", http.StatusNotFound)
		return
	}

	metricType := parts[0]
	metricName := parts[1]
	metricValue := parts[2]

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch MetricType(metricType) {
	case CounterType:
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.server.UpdateCounter(metricName, int64(value))
	case GaugeType:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.server.UpdateGauge(metricName, value)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
