package rest

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/core/model"
)

type ServerService interface {
	UpdateMetric(metricName, metricType, value string) error
	GetMetric(metricName, metricType string) (string, error)
	GetMetrics() []model.MetricData
}

type API struct {
	srv *http.Server
}

func NewServerAPI(server ServerService, port int64) *API {
	h := handler{
		server: server,
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.POST("/update/:type/:name/:value", h.update)

	router.GET("/value/:type/:name", h.get)

	router.GET("/", h.metrics)

	log.Printf("Server started on port %d", port)

	return &API{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
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

func (h *handler) update(ginCtx *gin.Context) {

	var (
		metricType  = ginCtx.Param("type")
		metricName  = ginCtx.Param("name")
		metricValue = ginCtx.Param("value")
	)

	err := h.server.UpdateMetric(metricName, metricType, metricValue)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			log.Printf("failed to update metric: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
}

func (h *handler) get(ginCtx *gin.Context) {

	var (
		metricType = ginCtx.Param("type")
		metricName = ginCtx.Param("name")
	)

	value, err := h.server.GetMetric(metricName, metricType)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			log.Printf("failed to get metric: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
	_, err = ginCtx.Writer.Write([]byte(value))
	if err != nil {
		log.Printf("failed to write response: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) metrics(ginCtx *gin.Context) {
	ginCtx.Writer.WriteHeader(http.StatusOK)

	metrics := h.server.GetMetrics()
	ginCtx.Writer.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("").Parse(metricsTemplate)
	if err != nil {
		log.Printf("failed to parse template: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(ginCtx.Writer, metrics)
	if err != nil {
		log.Printf("failed to execute template: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const metricsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <table border="1">
        <tr>
            <th>Name</th>
            <th>Value</th>
        </tr>
        {{ range . }}
        <tr>
            <td>{{ .Name }}</td>
            <td>{{ .Value }}</td>
        </tr>
        {{ end }}
    </table>
</body>
</html>
`
