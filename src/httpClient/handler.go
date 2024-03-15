package httpClient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/zeropsio/zcli/src/uuid"

	"github.com/zeropsio/zcli/src/support"
)

type Config struct {
	HttpTimeout time.Duration
}

type Response struct {
	StatusCode int
	Body       []byte
}

type Handler struct {
	config    Config
	supportID string
}

func New(ctx context.Context, config Config) *Handler {
	supportID, ok := support.GetID(ctx)
	if !ok {
		supportID = uuid.GetShort()
	}
	return &Handler{
		config:    config,
		supportID: supportID,
	}
}

type Option func(cfg *optionConfig)

func ContentType(contentType string) Option {
	return func(cfg *optionConfig) {
		cfg.headers["Content-Type"] = contentType
	}
}

func ContentLength(contentLength int64) Option {
	return func(cfg *optionConfig) {
		cfg.contentLength = contentLength
	}
}

type optionConfig struct {
	contentLength int64
	headers       map[string]string
}

func (h *Handler) PutStream(ctx context.Context, url string, body io.Reader, options ...Option) (Response, error) {
	return h.doStream(ctx, "PUT", url, body, options...)
}

func (h *Handler) Get(ctx context.Context, url string, options ...Option) (Response, error) {
	return h.do(ctx, "GET", url, nil, options...)
}

func (h *Handler) do(ctx context.Context, method string, url string, data []byte, options ...Option) (Response, error) {
	return h.doStream(ctx, method, url, bytes.NewReader(data), options...)
}

func (h *Handler) doStream(ctx context.Context, method string, url string, body io.Reader, options ...Option) (Response, error) {
	cfg := &optionConfig{
		headers: map[string]string{
			"Content-Type": "application/json",
			support.Key:    h.supportID,
		},
	}
	for _, o := range options {
		o(cfg)
	}

	client := &http.Client{Timeout: h.config.HttpTimeout}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	req.ContentLength = cfg.contentLength
	if err != nil {
		return Response{}, err
	}

	for key, value := range cfg.headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
	}, nil
}
