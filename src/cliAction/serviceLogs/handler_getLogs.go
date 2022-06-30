package serviceLogs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"io/ioutil"
	"net/http"
	"strings"
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

func getLogs(ctx context.Context, method, url, format, formatTemplate string) error {
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

	err = parseResponseByFormat(body, format, formatTemplate)
	if err != nil {
		return err
	}
	return nil
}

func parseResponseByFormat(body []byte, format, formatTemplate string) error {
	var err error

	var jsonData Response
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return err
	}

	if format == "FULL" {
		if formatTemplate != "" {
			if err = getFullWithTemplate(jsonData, formatTemplate); err != nil {
				return err
			}
			return nil
		} else {
			// TODO get rfc from config when implemented as flag
			getFullByRfc(jsonData, RFC5424)
			return nil
		}
	} else if format == "SHORT" {
		for _, o := range jsonData.Items {
			fmt.Printf("%v %s \n", o.Timestamp, o.Content)
		}
	} else {
		for _, o := range jsonData.Items {
			val, err := json.Marshal(o)
			if err != nil {
				return err
			}
			fmt.Println(string(val))
		}
	}

	return nil
}

func getLogRequestData(resOutput output.ProjectLog) (string, string, types.DateTime) {
	outputUrl := string(resOutput.Url)
	urlData := strings.Split(outputUrl, " ")
	method, url := urlData[0], urlData[1]

	// TODO enable token when websocket is used and return it
	// accessToken := resOutput.AccessToken
	expiration := resOutput.Expiration

	return method, HTTP + url, expiration
}

func makeQueryParams(limit, facility, minSeverity int, logServiceId, containerId, mode string) string {
	var desc = 1
	if mode == "RESPONSE" {
		desc = 0
	}

	if containerId != "" {
		return fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s&containerId=%s&minimumSeverity=%d",
			limit, desc, facility, logServiceId, containerId, minSeverity)
	}

	return fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s&minimumSeverity=%d",
		limit, desc, facility, logServiceId, minSeverity,
	)
}
