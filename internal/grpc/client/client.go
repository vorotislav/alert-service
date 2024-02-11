// Пакет client скрывает работу с grpc протоколом предоставляя один метод для отправки метрик.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	pb "github.com/vorotislav/alert-service/proto"

	"github.com/vorotislav/alert-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client основная сущность для отправки метрик. Содержит в себе grpc соединение, логгер, настройки и URL сервера.
type Client struct {
	gc            proto.MetricsClient
	logger        *zap.Logger
	serverAddress string
	conn          *grpc.ClientConn
}

// NewClient конструктор для Client.
func NewClient(logger *zap.Logger, serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to server: %w", err)
	}

	gc := pb.NewMetricsClient(conn)

	logger.Debug("grpc client created on:", zap.String("grpc server address", serverAddress))
	return &Client{
		gc:            gc,
		logger:        logger.With(zap.String("package", "grpc client")),
		serverAddress: serverAddress,
		conn:          conn,
	}, nil
}

// Stop реализует интерфейс Client. Закрывает grpc соединение.
func (c *Client) Stop() error {
	c.logger.Info("stop grpc client")

	return c.conn.Close()
}

// SendMetrics метод отправки метрик на сервер. Принимает карту с метриками и возвращает ошибку.
func (c *Client) SendMetrics(metrics map[string]*model.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, v := range metrics {
		_, err := c.gc.AddMetric(ctx, &pb.AddMetricRequest{
			Metric: &pb.Metric{
				Id:    v.ID,
				Type:  convertMetricTypeToPB(v.MType),
				Delta: v.Delta,
				Value: v.Value,
			},
		})

		if err != nil {
			return fmt.Errorf("cannot update metric [%s]: %w", v.ID, err)
		}

		c.logger.Debug("successful update metric", zap.String("metric id", v.ID))
	}

	return nil
}

func convertMetricTypeToPB(mtype string) pb.Metric_MetricType {
	switch mtype {
	case model.MetricCounter:
		return pb.Metric_COUNTER
	case model.MetricGauge:
		return pb.Metric_GAUGE
	}

	return pb.Metric_UNSPECIFIED
}
