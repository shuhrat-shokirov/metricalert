package model

type Metric struct {
	Name  string
	Value any
	Type  string
}

type MetricData struct {
	Name  string
	Value string
}

type MetricRequest struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}
