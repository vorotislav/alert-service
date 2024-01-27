// Пакет client скрывает работу с http протоколом предоставляя один метод для отправки метрик.
package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vorotislav/alert-service/internal/model"
	"github.com/vorotislav/alert-service/internal/settings/agent"
	"github.com/vorotislav/alert-service/internal/utils"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

// ErrSendMetrics ошибка, в случае неудачи отправки.
var (
	ErrSendMetrics = errors.New("cannot send metrics")
)

const (
	maxRetryAttempt = 4
)

const (
	defaultClientTimeout = time.Millisecond * 700
)

// Client основная сущность для отправки метрик. Содержит в себе http.Client, логгер, настройки и URL сервера.
type Client struct {
	dc        *http.Client
	logger    *zap.Logger
	set       *agent.Settings
	serverURL string
}

// NewClient конструктор для Client.
func NewClient(logger *zap.Logger, set *agent.Settings) *Client {
	c := &Client{
		dc: &http.Client{
			Timeout: defaultClientTimeout,
		},
		logger:    logger,
		set:       set,
		serverURL: fmt.Sprintf("http://%s/update", set.ServerAddress),
	}

	return c
}

// SendMetrics метод отправки метрик на сервер. Принимает карту с метриками и возвращает ошибку.
func (c *Client) SendMetrics(metrics map[string]*model.Metrics) error {
	ms := c.convertMetricsToSlice(metrics)

	jobs := make(chan *model.Metrics, c.set.RateLimit)
	results := make(chan error, c.set.RateLimit)

	for w := 1; w <= c.set.RateLimit; w++ {
		go c.sendWorker(w, jobs, results)
	}

	for _, m := range ms {
		m := m
		jobs <- m
	}

	close(jobs)

	return nil
}

func (c *Client) sendWorker(id int, jobs <-chan *model.Metrics, results chan<- error) {
	for j := range jobs {
		c.logger.Debug(fmt.Sprintf("worker %d started job: %s", id, j.ID))

		err := c.sendMetricRetry(j)
		if err != nil {
			c.logger.Debug(fmt.Sprintf("worker %d failed job: %s", id, j.ID))
			results <- err
		}
	}

	c.logger.Debug(fmt.Sprintf("worker %d done", id))
}

func (c *Client) convertMetricsToSlice(metrics map[string]*model.Metrics) []*model.Metrics {
	m := make([]*model.Metrics, 0, len(metrics))

	for _, v := range metrics {
		m = append(m, v)
	}

	return m
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

	if c.set.HashKey != "" {
		hash, err := utils.GetHash(raw, []byte(c.set.HashKey))
		if err != nil {
			c.logger.Error("cannot get hash of metric", zap.Error(err))
		}

		req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(hash))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	err = retry.Do(
		func() error {
			resp, err := c.dc.Do(req)
			if resp != nil {
				defer func() {
					_ = resp.Body.Close()
				}()
			}

			if err != nil || resp.StatusCode >= http.StatusInternalServerError {
				return fmt.Errorf("cannot do request: %w", err)
			}

			return nil
		},
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
		retry.Attempts(maxRetryAttempt),
		retry.Context(ctx),
	)

	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
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
