package httpClient

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
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

func BearerAuthorization(token string) Option {
	return func(cfg *optionConfig) {
		cfg.headers["Authorization"] = "Bearer " + token
	}
}

type optionConfig struct {
	headers map[string]string
}

func (h *Handler) PutStream(url string, body io.Reader, options ...Option) (Response, error) {
	return h.doStream("PUT", url, body, options...)
}

func (h *Handler) Put(url string, data []byte, options ...Option) (Response, error) {
	return h.do("PUT", url, data, options...)
}

func (h *Handler) Post(url string, data []byte, options ...Option) (Response, error) {
	return h.do("POST", url, data, options...)
}

func (h *Handler) Get(url string, options ...Option) (Response, error) {
	return h.do("GET", url, nil)
}

func (h *Handler) do(method string, url string, data []byte, options ...Option) (Response, error) {
	return h.doStream(method, url, bytes.NewReader(data), options...)
}

func (h *Handler) doStream(method string, url string, body io.Reader, options ...Option) (Response, error) {
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
	req, err := http.NewRequest(method, url, body)
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
	}, nil
}
