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
	"github.com/zeropsio/zerops-go/types"
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

	logs := jsonData.Items
	ascLogs := reverseLogs(logs)

	if format == FULL {
		if formatTemplate != "" {
			if err = getFullWithTemplate(ascLogs, templateFix(formatTemplate)); err != nil {
				return err
			}
			return nil
		} else {
			// TODO get rfc from config when implemented as flag
			getFullByRfc(ascLogs, RFC5424)
			return nil
		}
	} else if format == SHORT {
		for _, o := range ascLogs {
			fmt.Printf("%v %s \n", o.Timestamp, o.Content)
		}
	} else if format == JSONSTREAM {
		for _, o := range ascLogs {
			val, err := json.Marshal(o)
			if err != nil {
				return err
			}
			fmt.Println(string(val))
		}
	} else {
		val, err := json.Marshal(ascLogs)
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

func getLogRequestData(resOutput output.ProjectLog) (string, string, types.DateTime) {
	outputUrl := string(resOutput.Url)
	urlData := strings.Split(outputUrl, " ")
	method, url := urlData[0], urlData[1]

	// TODO enable token when websocket is used and return it
	// accessToken := resOutput.AccessToken
	expiration := resOutput.Expiration

	return method, HTTP + url, expiration
}

func makeQueryParams(limit, facility, minSeverity int, logServiceId, containerId string) string {
	query := fmt.Sprintf("&limit=%d&desc=1&facility=%d&serviceStackId=%s",
		limit, facility, logServiceId)

	if minSeverity != -1 {
		query += fmt.Sprintf("&minimumSeverity=%d", minSeverity)
	}

	if containerId != "" {
		query += fmt.Sprintf("&containerId=%s", containerId)
	}

	return query
}
