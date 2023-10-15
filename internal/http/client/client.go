package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vorotislav/alert-service/internal/utils"
	"net/http"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/agent"

	"go.uber.org/zap"
)

var (
	ErrSendMetrics = errors.New("cannot send metrics")
)

type Client struct {
	dc        *http.Client
	logger    *zap.Logger
	set       *agent.Settings
	serverURL string
}

func NewClient(logger *zap.Logger, set *agent.Settings) *Client {
	c := &Client{
		dc:        http.DefaultClient,
		logger:    logger.With(zap.String("package", "client")),
		set:       set,
		serverURL: fmt.Sprintf("http://%s/updates", set.ServerAddress),
	}

	return c
}

func (c *Client) SendMetrics(metrics map[string]*model.Metrics) error {

	newMetrics := make([]model.Metrics, 0, len(metrics))
	for _, m := range metrics {
		m := m
		newMetrics = append(newMetrics, *m)
	}

	if err := c.sendMetrics(newMetrics); err != nil {
		return fmt.Errorf("cannot send metrics: %w", err)
	}

	//for _, m := range metrics {
	//	m := m
	//	if err := c.sendMetric(m); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (c *Client) sendMetrics(metrics []model.Metrics) error {
	raw, err := json.Marshal(metrics)
	if err != nil {
		c.logger.Error("cannot metric marshal", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	compressRaw, err := utils.Compress(raw)
	if err != nil {
		c.logger.Error("cannot compress data", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.serverURL,
		bytes.NewBuffer(compressRaw),
	)

	if err != nil {
		c.logger.Error("cannot request prepare", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := c.dc.Do(req)
	if err != nil {
		c.logger.Error("cannot send metrics", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	if resp != nil {
		resp.Body.Close()
	}

	return nil
}

func (c *Client) sendMetric(metric *model.Metrics) error {
	raw, err := json.Marshal(metric)
	if err != nil {
		c.logger.Error("cannot metric marshal", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	compressRaw, err := utils.Compress(raw)
	if err != nil {
		c.logger.Error("cannot compress data", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.serverURL,
		bytes.NewBuffer(compressRaw),
	)

	if err != nil {
		c.logger.Error("cannot request prepare", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := c.dc.Do(req)
	if err != nil {
		c.logger.Error("cannot send metric", zap.Error(err))

		return fmt.Errorf("%w: %w", ErrSendMetrics, err)
	}

	if metric.Value != nil {
		c.logger.Debug("send metric",
			zap.String("name", metric.ID),
			zap.Float64("value", *metric.Value))
	} else if metric.Delta != nil {
		c.logger.Debug("send metric",
			zap.String("name", metric.ID),
			zap.Int64("value", *metric.Delta))
	}

	if resp != nil {
		resp.Body.Close()
	}

	return nil
}
