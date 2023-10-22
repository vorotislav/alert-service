package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/vorotislav/alert-service/internal/utils"
	"io"
	"net/http"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/agent"

	"go.uber.org/zap"
)

var (
	ErrSendMetrics = errors.New("cannot send metrics")
)

const (
	maxRetryAttempt = 4
	retryDelay      = 2
)

type Client struct {
	dc        *http.Client
	logger    *zap.Logger
	set       *agent.Settings
	serverURL string
}

func NewClient(logger *zap.Logger, set *agent.Settings) *Client {
	c := &Client{
		dc: &http.Client{
			Timeout: time.Millisecond * 500,
		},
		logger:    logger,
		set:       set,
		serverURL: fmt.Sprintf("http://%s/update", set.ServerAddress),
	}

	return c
}

func (c *Client) SendMetrics(metrics map[string]*model.Metrics) error {

	//newMetrics := make([]model.Metrics, 0, len(metrics))
	//for _, m := range metrics {
	//	m := m
	//	newMetrics = append(newMetrics, *m)
	//}

	//if err := c.sendMetrics(newMetrics); err != nil {
	//	return fmt.Errorf("cannot send metrics: %w", err)
	//}

	for _, m := range metrics {
		m := m
		if err := c.sendMetricRetry(m); err != nil {
			return err
		}
	}
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

func (c *Client) sendMetricRetry(metric *model.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

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

	err = retry.Do(
		func() error {
			resp, err := c.dc.Do(req)
			if resp != nil {
				defer resp.Body.Close()
			}

			if err != nil {
				return err
			}

			return nil
		},
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
		retry.Attempts(4),
		retry.Context(ctx),
	)

	if err != nil {
		return err
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

	resp, err := c.retryDo(req)
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

func (c *Client) retryDo(req *http.Request) (*http.Response, error) {
	var (
		originalBody []byte
		err          error
	)

	if req != nil && req.Body != nil {
		originalBody, err = copyBody(req.Body)
		resetBody(req, originalBody)
	}

	if err != nil {
		return nil, err
	}

	var resp *http.Response

	delay := time.Second

	for i := 1; i <= maxRetryAttempt; i++ {
		c.logger.Debug("Send request", zap.Int("attempt", i))
		resp, err = c.dc.Do(req)

		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}

		if err != nil {
			c.logger.Debug("error sending", zap.String("", err.Error()))
		}

		if resp != nil {
			c.logger.Debug("error sending", zap.Int("status code", resp.StatusCode))

			resp.Body.Close()
		}

		if req.Body != nil {
			resetBody(req, originalBody)
		}

		if i == maxRetryAttempt {
			break
		}

		newDelay := delay + (retryDelay * time.Second)

		c.logger.Debug("next attempt in", zap.String("sec", delay.String()))
		time.Sleep(delay)
		delay = newDelay
	}

	return resp, err
}

func copyBody(src io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("reading request body")
	}
	src.Close()
	return b, nil
}

func resetBody(request *http.Request, originalBody []byte) {
	request.Body = io.NopCloser(bytes.NewBuffer(originalBody))
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(originalBody)), nil
	}
}
