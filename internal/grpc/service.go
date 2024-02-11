package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vorotislav/alert-service/internal/model"

	"github.com/vorotislav/alert-service/internal/repository"
	pb "github.com/vorotislav/alert-service/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MetricServer struct {
	pb.UnimplementedMetricsServer

	logger  *zap.Logger
	repo    repository.Repository
	address string
	s       *grpc.Server
}

func NewMetricServer(log *zap.Logger, repo repository.Repository, address string) *MetricServer {

	ms := &MetricServer{
		logger:  log.With(zap.String("package", "grpc service")),
		repo:    repo,
		address: address,
	}

	ms.s = grpc.NewServer()

	pb.RegisterMetricsServer(ms.s, ms)

	return ms
}

func (ms *MetricServer) AddMetric(ctx context.Context, in *pb.AddMetricRequest) (*pb.AddMetricResponse, error) {

	m := model.Metrics{
		ID:    in.Metric.Id,
		MType: convertPBMetricType(in.Metric.Type),
		Delta: in.Metric.Delta,
		Value: in.Metric.Value,
	}

	resp, err := ms.repo.UpdateMetric(ctx, m)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update metrics: %s", err.Error())
	}

	return &pb.AddMetricResponse{
		Metric: &pb.Metric{
			Id:    resp.ID,
			Type:  convertMetricTypeToPB(resp.MType),
			Delta: resp.Delta,
			Value: resp.Value,
		},
	}, nil
}

func (ms *MetricServer) Run() error {
	ms.logger.Info("Running grpc server on", zap.String("address", ms.address))

	listen, err := net.Listen("tcp", ms.address)
	if err != nil {
		return fmt.Errorf("cannot listen address: %w", err)
	}

	return ms.s.Serve(listen)
}

func (ms *MetricServer) Stop(_ context.Context) error {
	ms.logger.Info("Stopping grpc server")

	ms.s.GracefulStop()

	return nil
}

func convertPBMetricType(metricType pb.Metric_MetricType) string {
	switch metricType {
	case pb.Metric_COUNTER:
		return model.MetricCounter
	case pb.Metric_GAUGE:
		return model.MetricGauge
	}

	return ""
}

func convertMetricTypeToPB(mt string) pb.Metric_MetricType {
	switch mt {
	case model.MetricCounter:
		return pb.Metric_COUNTER
	case model.MetricGauge:
		return pb.Metric_GAUGE
	}

	return pb.Metric_UNSPECIFIED
}
