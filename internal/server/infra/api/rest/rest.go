package rest

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/core/model"
)

type ServerService interface {
	UpdateMetric(metricName, metricType string, value any) error
	GetMetric(metricName, metricType string) (string, error)
	GetMetrics() []model.MetricData
}

type API struct {
	srv *http.Server
}

func NewServerAPI(server ServerService, port int64, sugar zap.SugaredLogger) *API {
	h := handler{
		server: server,
		sugar:  sugar,
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(h.MwLog())
	router.Use(h.mwDecompress())
	router.Use(h.responseGzipMiddleware())

	router.POST("/update/:type/:name/:value", h.update)

	router.POST("/update/", h.updateWithBody)

	router.GET("/value/:type/:name", h.get)

	router.POST("/value/", h.getMetricValue)

	router.GET("/", h.metrics)

	log.Printf("Server started on port %d", port)

	return &API{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
	}
}

func (h *handler) mwDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			c.Next()
			return
		}

		gzipReader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			log.Printf("failed to create gzip reader: %v", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			err := gzipReader.Close()
			if err != nil {
				log.Printf("failed to close gzip reader: %v", err)
			}
		}()

		c.Request.Body = io.NopCloser(gzipReader)

		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Accept-Encoding", "gzip")

		c.Next()
	}
}

func (h *handler) MwLog() gin.HandlerFunc {
	return func(c *gin.Context) {

		now := time.Now()

		c.Next()

		h.sugar.Infoln(
			"URI: ", c.Request.RequestURI,
			"Method: ", c.Request.Method,
			"Latency: ", time.Since(now).String(),
			"Status: ", c.Writer.Status(),
			"Size: ", c.Writer.Size(),
		)

	}

}

func (a *API) Run() error {
	return a.srv.ListenAndServe()
}

type handler struct {
	server ServerService
	sugar  zap.SugaredLogger
}

func (h *handler) update(ginCtx *gin.Context) {

	var (
		metricType  = ginCtx.Param("type")
		metricName  = ginCtx.Param("name")
		metricValue = ginCtx.Param("value")
	)

	var value any

	switch metricType {
	case "counter":
		v, err := strconv.Atoi(metricValue)
		if err != nil {
			log.Printf("failed to parse counter value: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		value = int64(v)
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Printf("failed to parse gauge value: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		value = v
	default:
		log.Printf("unknown metric type")
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.server.UpdateMetric(metricName, metricType, value)
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

func (h *handler) updateWithBody(ginCtx *gin.Context) {

	var metric metrics

	err := ginCtx.BindJSON(&metric)
	if err != nil {
		log.Printf("failed to bind json: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var value any
	switch metric.MType {
	case "counter":
		if metric.Delta == nil {
			log.Printf("delta is nil")
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		value = *metric.Delta
	case "gauge":
		if metric.Value == nil {
			log.Printf("value is nil")
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		value = *metric.Value
	default:
		log.Printf("unknown metric type")
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.server.UpdateMetric(metric.ID, metric.MType, value)
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

type metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

func (h *handler) getMetricValue(ginCtx *gin.Context) {

	var request metrics

	err := ginCtx.BindJSON(&request)
	if err != nil {
		log.Printf("failed to bind json: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := h.server.GetMetric(request.ID, request.MType)
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

	response := metrics{
		ID:    request.ID,
		MType: request.MType,
	}

	switch request.MType {
	case "counter":
		v, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("failed to parse counter value: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricValue := int64(v)

		response.Delta = &metricValue
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("failed to parse gauge value: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Value = &v
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
	ginCtx.Header("Content-Type", "application/json")
	_, err = ginCtx.Writer.Write(bytes)
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

func (h *handler) responseGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, поддерживает ли клиент gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Выполняем обработку запроса и сохраняем ответ
		c.Writer.Header().Set("Content-Encoding", "gzip")
		c.Writer.Header().Set("Accept-Encoding", "gzip")

		// Перенаправляем вывод в gzip.Writer
		gz := gzip.NewWriter(c.Writer)
		defer func() {
			err := gz.Close()
			if err != nil {
				log.Printf("failed to close gzip writer: %v", err)
			}
		}()

		// Заменяем Writer на обертку для gzip
		c.Writer = &gzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}

		c.Next()
	}
}

type gzipResponseWriter struct {
	io.Writer
	gin.ResponseWriter
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	// Проверяем тип контента и выполняем сжатие только для JSON и HTML
	contentType := w.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		return w.Writer.Write(data)
	}
	return w.ResponseWriter.Write(data)
}
