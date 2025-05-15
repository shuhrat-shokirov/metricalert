package grpcclient

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/credentials/insecure"

	"metricalert/internal/server/core/model"
	pb "metricalert/proto"

	"google.golang.org/grpc"
)

type Client interface {
	SendMetrics(ctx context.Context, metrics []model.Metric, ip string) error
}

type MetricsClient struct {
	client pb.MetricsServiceClient
}

func NewMetricsClient(address string) (Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	log.Printf("grpc client created")

	return &MetricsClient{
		client: pb.NewMetricsServiceClient(conn),
	}, nil
}

func (c *MetricsClient) SendMetrics(ctx context.Context, metrics []model.Metric, ip string) error {
	var grpcMetrics = make([]*pb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		var m = &pb.Metric{
			Id:   metric.Name,
			Type: metric.Type,
		}

		switch m.GetType() {
		case "counter":
			v, ok := metric.Value.(int64)
			if !ok {
				return fmt.Errorf("invalid counter value type, type: %T, value: %v", metric.Value, metric.Value)
			}
			m.Delta = v
		case "gauge":
			v, ok := metric.Value.(float64)
			if !ok {
				return fmt.Errorf("invalid gauge value type, type: %T, value: %v", metric.Value, metric.Value)
			}
			m.Value = v
		}

		grpcMetrics = append(grpcMetrics, m)
	}

	_, err := c.client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: grpcMetrics})
	if err != nil {
		return fmt.Errorf("failed to send metrics to grpc server: %w", err)
	}

	return nil
}
