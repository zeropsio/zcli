package httpClient

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

type Config struct {
	HttpTimeout time.Duration
}

type Response struct {
	StatusCode int
	Body       []byte
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
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
	cfg := &optionConfig{
		headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	for _, o := range options {
		o(cfg)
	}

	client := &http.Client{Timeout: h.config.HttpTimeout}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
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
