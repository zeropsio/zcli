package serviceLogs

import (
	"context"
	"encoding/json"
	"io/ioutil"
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

func getLogs(ctx context.Context, method, url, format, formatTemplate, mode string) error {
	c := http.Client{Timeout: time.Duration(1) * time.Minute}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	jsonData, err := parseResponse(body)
	if err != nil {
		return err
	}
	err = parseResponseByFormat(jsonData, format, formatTemplate, mode)
	if err != nil {
		return err
	}
	return nil
}

func parseResponse(body []byte) (Response, error) {
	var jsonData Response
	err := json.Unmarshal(body, &jsonData)
	if err != nil || len(jsonData.Items) == 0 {
		return Response{}, err
	}
	return jsonData, nil
}
