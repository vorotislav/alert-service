// Пакет client скрывает работу с grpc протоколом предоставляя один метод для отправки метрик.
package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc/metadata"

	interLog "github.com/vorotislav/alert-service/internal/grpc/middlewares/log"
	"github.com/vorotislav/alert-service/internal/model"
	pb "github.com/vorotislav/alert-service/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client основная сущность для отправки метрик. Содержит в себе grpc соединение, логгер, настройки и URL сервера.
type Client struct {
	gc            pb.MetricsClient
	logger        *zap.Logger
	serverAddress string
	conn          *grpc.ClientConn
	localAddress  string
}

// NewClient конструктор для Client.
func NewClient(logger *zap.Logger, serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			interLog.NewClientLoggerInterceptor(logger.With(zap.String("package", "grpc interceptor")))),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to server: %w", err)
	}

	gc := pb.NewMetricsClient(conn)

	logger.Debug("grpc client created on:", zap.String("grpc server address", serverAddress))

	la, err := getIPAddress()
	if err != nil {
		logger.Error("Get ip address", zap.Error(err))
	}

	return &Client{
		gc:            gc,
		logger:        logger.With(zap.String("package", "grpc client")),
		serverAddress: serverAddress,
		conn:          conn,
		localAddress:  la,
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

	md := metadata.New(map[string]string{"X-Real-IP": c.localAddress})
	ctx = metadata.NewOutgoingContext(ctx, md)

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

func getIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("cannot get interface addresses: %w", err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("cannot get ip address")
}
