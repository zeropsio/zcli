package serviceLogs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/zeropsio/zerops-go/dto/output"
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

	err = parseResponseByFormat(body, format, formatTemplate, mode)
	if err != nil {
		return err
	}
	return nil
}

func parseResponseByFormat(body []byte, format, formatTemplate, mode string) error {
	var err error

	var jsonData Response
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return err
	}

	logs := jsonData.Items
	if mode == RESPONSE {
		logs = reverseLogs(logs)
	}

	if format == FULL {
		if formatTemplate != "" {
			if err = getFullWithTemplate(logs, formatTemplate); err != nil {
				return err
			}
			return nil
		} else {
			// TODO get rfc from config when implemented as flag
			getFullByRfc(logs, RFC5424)
			return nil
		}
	} else if format == SHORT {
		for _, o := range logs {
			fmt.Printf("%v %s \n", o.Timestamp, o.Content)
		}
	} else if format == JSONSTREAM {
		for _, o := range logs {
			val, err := json.Marshal(o)
			if err != nil {
				return err
			}
			fmt.Println(string(val))
		}
	} else {
		val, err := json.Marshal(logs)
		if err != nil {
			return err
		}
		fmt.Println(string(val))
	}

	return nil
}

// reverseLogs makes log order ASC to get the last logs of given limit
func reverseLogs(data []Data) []Data {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

func getLogRequestData(resOutput output.ProjectLog) (string, string) {
	outputUrl := string(resOutput.Url)
	urlData := strings.Split(outputUrl, " ")
	method, url := urlData[0], urlData[1]

	return method, url
}

func makeQueryParams(inputs InputValues, logServiceId, containerId string) string {
	query := fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s",
		inputs.limit, getDesc(inputs.mode), inputs.facility, logServiceId)

	if inputs.minSeverity != -1 {
		query += fmt.Sprintf("&minimumSeverity=%d", inputs.minSeverity)
	}

	if containerId != "" {
		query += fmt.Sprintf("&containerId=%s", containerId)
	}

	return query
}

func getDesc(mode string) int {
	if mode == RESPONSE {
		return 1
	}
	return 0
}
