package services

import (
	"testing"
)

func TestCollector_CollectMetrics(t *testing.T) {
	collector := NewCollector()

	metrics := collector.CollectMetrics()

	if len(metrics) == 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}

	for _, metric := range metrics {
		if metric.Name == "" {
			t.Errorf("Expected non-empty metric name, got empty")
		}
		if metric.Value == nil {
			t.Errorf("Expected non-nil metric value, got nil")
		}

		switch metric.Type {
		case "gauge":
			if _, ok := metric.Value.(float64); !ok {
				t.Errorf("Expected metric value type float64, got %T", metric.Value)
			}
		case "counter":
			if _, ok := metric.Value.(int64); !ok {
				t.Errorf("Expected metric value type int64, got %T", metric.Value)
			}
		default:
			t.Errorf("Expected metric type gauge or counter, got %s", metric.Type)
		}
	}
}

func TestCollector_CollectMemoryMetrics(t *testing.T) {
	collector := NewCollector()

	metrics := collector.CollectMemoryMetrics()

	if len(metrics) == 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}

	for _, metric := range metrics {
		if metric.Name == "" {
			t.Errorf("Expected non-empty metric name, got empty")
		}
		if metric.Value == nil {
			t.Errorf("Expected non-nil metric value, got nil")
		}

		switch metric.Type {
		case "gauge":
			if _, ok := metric.Value.(float64); !ok {
				t.Errorf("Expected metric value type float64, got %T", metric.Value)
			}
		default:
			t.Errorf("Expected metric type gauge or counter, got %s", metric.Type)
		}
	}
}

func TestCollector_ResetCounters(t *testing.T) {
	collector := NewCollector()

	if collector.pollCount != 0 {
		t.Errorf("Expected pollCount 0, got %d", collector.pollCount)
	}

	collector.CollectMetrics()

	if collector.pollCount != 1 {
		t.Errorf("Expected pollCount 1, got %d", collector.pollCount)
	}

	collector.ResetCounters()

	if collector.pollCount != 0 {
		t.Errorf("Expected pollCount 0, got %d", collector.pollCount)
	}
}
