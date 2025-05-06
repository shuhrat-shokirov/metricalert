package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "metricalert/proto"

	"metricalert/internal/server/core/application"
	"metricalert/internal/server/core/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	app *application.Application
}

func NewMetricsServer(app *application.Application) *MetricsServer {
	return &MetricsServer{app: app}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context,
	req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var metrics = make([]model.MetricRequest, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		metrics = append(metrics, model.MetricRequest{
			ID:    m.GetId(),
			MType: m.GetType(),
			Value: &m.Value,
			Delta: &m.Delta,
		})
	}

	err := s.app.UpdateMetrics(ctx, metrics)
	if err != nil {
		return nil, fmt.Errorf("update metrics: %w", err)
	}

	return &pb.UpdateMetricsResponse{Status: "success"}, nil
}

func StartGRPCServer(app *application.Application, address string) error {
	server := grpc.NewServer()
	pb.RegisterMetricsServiceServer(server, NewMetricsServer(app))
	reflection.Register(server)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("grpc server listening on %s", address)

	err = server.Serve(lis)
	if err != nil {
		return fmt.Errorf("grpc server failed to serve: %w", err)
	}

	return nil
}
