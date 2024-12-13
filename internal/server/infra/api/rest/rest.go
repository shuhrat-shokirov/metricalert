package rest

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
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
	UpdateMetric(ctx context.Context, request model.MetricRequest) error
	UpdateMetrics(ctx context.Context, request []model.MetricRequest) error
	GetMetric(ctx context.Context, metricName, metricType string) (string, error)
	GetMetrics(ctx context.Context) ([]model.MetricData, error)
	Ping(ctx context.Context) error
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

	router.POST("/updates/", h.batchUpdate)

	router.GET("/value/:type/:name", h.get)

	router.POST("/value/", h.getMetricValue)

	router.GET("/ping", h.dbPing)

	router.GET("/", h.metrics)

	h.sugar.Infof("server started on port: %d", port)

	return &API{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
	}
}

func (h *handler) mwDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		const gzipScheme = "gzip"
		if !strings.Contains(c.GetHeader("Content-Encoding"), gzipScheme) {
			c.Next()
			return
		}

		gzipReader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			h.sugar.Errorf("failed to create gzip reader: %v", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			err := gzipReader.Close()
			if err != nil {
				h.sugar.Errorf("failed to close gzip reader: %v", err)
			}
		}()

		c.Request.Body = io.NopCloser(gzipReader)

		c.Writer.Header().Set("Content-Encoding", gzipScheme)
		c.Writer.Header().Set("Accept-Encoding", gzipScheme)

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
	if err := a.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}

	return nil
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

	request := model.MetricRequest{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case counterType:
		v, err := strconv.Atoi(metricValue)
		if err != nil {
			h.sugar.Errorf("failed to parse counter, value: %s, error: %v", metricValue, err)
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		value := int64(v)

		request.Delta = &value
	case gaugeType:
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			h.sugar.Errorf("failed to parse gauge, value: %s, error: %v", metricValue, err)
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		request.Value = &v
	default:
		h.sugar.Errorf("unknown metric type: %s", metricType)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.server.UpdateMetric(context.TODO(), request)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			h.sugar.Errorf("failed to update metric: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
}

func (h *handler) updateWithBody(ginCtx *gin.Context) {
	var metric model.MetricRequest

	err := ginCtx.BindJSON(&metric)
	if err != nil {
		h.sugar.Errorf("failed to bind json: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.server.UpdateMetric(context.TODO(), metric)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			h.sugar.Errorf("failed to update metric: %v", err)
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

	value, err := h.server.GetMetric(context.TODO(), metricName, metricType)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			h.sugar.Errorf("failed to get metric: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
	_, err = ginCtx.Writer.WriteString(value)
	if err != nil {
		h.sugar.Errorf("failed to write response: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) getMetricValue(ginCtx *gin.Context) {
	var request model.MetricRequest

	err := ginCtx.BindJSON(&request)
	if err != nil {
		h.sugar.Errorf("failed to bind json: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := h.server.GetMetric(context.TODO(), request.ID, request.MType)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			h.sugar.Errorf("failed to get metric: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response := model.MetricRequest{
		ID:    request.ID,
		MType: request.MType,
	}

	switch request.MType {
	case counterType:
		v, err := strconv.Atoi(value)
		if err != nil {
			h.sugar.Errorf("failed to parse counter, value: %s, error: %v", value, err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		metricValue := int64(v)

		response.Delta = &metricValue
	case gaugeType:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			h.sugar.Errorf("failed to parse gauge, value: %s, error: %v", value, err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Value = &v
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		h.sugar.Errorf("failed to marshal response: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
	ginCtx.Header("Content-Type", "application/json")
	_, err = ginCtx.Writer.Write(bytes)
	if err != nil {
		h.sugar.Errorf("failed to write response: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) metrics(ginCtx *gin.Context) {
	ginCtx.Writer.WriteHeader(http.StatusOK)

	metrics, err := h.server.GetMetrics(context.TODO())
	if err != nil {
		h.sugar.Errorf("failed to get metrics: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	ginCtx.Writer.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("").Parse(metricsTemplate)
	if err != nil {
		h.sugar.Errorf("failed to parse template: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(ginCtx.Writer, metrics)
	if err != nil {
		h.sugar.Errorf("failed to execute template: %v", err)
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
				h.sugar.Errorf("failed to close gzip writer: %v", err)
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
		n, err := w.Writer.Write(data)
		if err != nil {
			return n, fmt.Errorf("failed to write data: %w", err)
		}

		return n, nil
	}

	write, err := w.ResponseWriter.Write(data)
	if err != nil {
		return write, fmt.Errorf("failed to write data: %w", err)
	}

	return write, nil
}

func (h *handler) dbPing(ginCtx *gin.Context) {
	err := h.server.Ping(ginCtx.Request.Context())
	if err != nil {
		h.sugar.Errorf("failed to ping db: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
}

const (
	counterType = "counter"
	gaugeType   = "gauge"
)

func (h *handler) batchUpdate(ginCtx *gin.Context) {
	var request []model.MetricRequest

	err := ginCtx.BindJSON(&request)
	if err != nil {
		h.sugar.Errorf("failed to bind json: %v", err)
		ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.server.UpdateMetrics(context.TODO(), request)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBadRequest):
			ginCtx.Writer.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, application.ErrNotFound):
			ginCtx.Writer.WriteHeader(http.StatusNotFound)
		default:
			h.sugar.Errorf("failed to update metrics: %v", err)
			ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	ginCtx.Writer.WriteHeader(http.StatusOK)
}
