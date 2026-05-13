package serviceLogs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func (h *Handler) getLogs(ctx context.Context, method, url, format, formatTemplate, mode string) error {
	c := http.Client{Timeout: time.Minute}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		// Surface a meaningful error instead of feeding a 4xx/5xx HTML page
		// to json.Unmarshal and reporting a parse failure.
		return errors.Errorf("log request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	jsonData, err := parseResponse(body)
	if err != nil {
		return err
	}
	return h.parseResponseByFormat(jsonData, format, formatTemplate, mode)
}

func parseResponse(body []byte) (Response, error) {
	var jsonData Response
	err := json.Unmarshal(body, &jsonData)
	if err != nil {
		return Response{}, err
	}
	return jsonData, nil
}
