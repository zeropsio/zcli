package serviceLogs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	Items []Data `json:"items"`
}

type Data struct {
	Timestamp      string `json:"timestamp"`
	Version        int    `json:"version"`
	Hostname       string `json:"hostname"`
	Content        string `json:"content"`
	Client         string `json:"client"`
	Facility       int    `json:"facility"`
	FacilityLabel  string `json:"facilityLabel"`
	Id             string `json:"id"`
	MsgId          string `json:"msgId"`
	Priority       int    `json:"priority"`
	ProcId         string `json:"procId"`
	Severity       int    `json:"severity"`
	SeverityLabel  string `json:"severityLabel"`
	StructuredData string `json:"structuredData"`
	Tag            string `json:"tag"`
	TlsPeer        string `json:"tlsPeer"`
	AppName        string `json:"appName"`
	Message        string `json:"message"`
}

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrLogResponse     = errors.New("log response error")
	ErrInvalidResponse = errors.New("invalid response")
)

type InvalidRequestError struct {
	FuncName string
	Msg      string
	Err      error
}

func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.FuncName, e.Msg, e.Err)
}

func NewInvalidRequestError(funcName, msg string, err error) error {
	return &InvalidRequestError{FuncName: funcName, Msg: msg, Err: err}
}

type LogResponseError struct {
	StatusCode int
	Msg        string
	Err        error
}

func (e *LogResponseError) Error() string {
	return fmt.Sprintf("status code: %d: %s: %v", e.StatusCode, e.Msg, e.Err)
}

func NewLogResponseError(statusCode int, msg string, err error) error {
	return &LogResponseError{StatusCode: statusCode, Msg: msg, Err: err}
}

func getLogs(ctx context.Context, method, url, format, formatTemplate, mode string) error {
	c := http.Client{Timeout: time.Duration(1) * time.Minute}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return NewInvalidRequestError("getLogs", "failed to create request", err)
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return NewInvalidRequestError("getLogs", "failed to execute request", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewLogResponseError(resp.StatusCode, "failed to read response body", err)
	}

	if resp.StatusCode != http.StatusOK {
		return NewLogResponseError(resp.StatusCode, fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	jsonData, err := parseResponse(body)
	if err != nil {
		return NewLogResponseError(resp.StatusCode, "failed to parse response", err)
	}

	return parseResponseByFormat(jsonData, format, formatTemplate, mode)
}

func parseResponse(body []byte) (Response, error) {
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, NewLogResponseError(0, "failed to unmarshal response", err)
	}
	return response, nil
}
