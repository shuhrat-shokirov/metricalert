package rest

import (
	"errors"
	"net/http"
	"strings"

	"metricalert/internal/core/model"
)

type ServerService interface {
	UpdateMetric(metricName, metricType, value string) error
}

type API struct {
	srv *http.Server
}

func NewServerAPI(server ServerService) *API {
	h := handler{
		server: server,
	}

	router := http.NewServeMux()

	router.HandleFunc("/update/", h.update)

	return &API{
		srv: &http.Server{
			Addr:    ":8080",
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

	err := h.server.UpdateMetric(metricName, metricType, metricValue)
	if err != nil {
		h.handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) handleErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrorBadRequest):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, model.ErrorNotFound):
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
