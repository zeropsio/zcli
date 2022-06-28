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
}

func GetLogs(ctx context.Context, method, url, format string) error {
	c := http.Client{Timeout: time.Duration(1) * time.Minute}
	fmt.Println("req: ", url)
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

	fmt.Println(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = parseResponseByFormat(body, format)
	if err != nil {
		return err
	}
	return nil
}

func parseResponseByFormat(body []byte, format string) error {
	var err error

	var jsonData Response
	err = json.Unmarshal(body, &jsonData)

	if format == "FULL" {
		// TODO format the data
		fmt.Println(jsonData.Items)
	} else if format == "SHORT" {
		for _, o := range jsonData.Items {
			fmt.Printf("%s %s \n", o.Timestamp, o.Content)
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

	return err
}
func getLogRequestData(resOutput output.ProjectLog) (string, string, types.DateTime) {
	outputUrl := string(resOutput.Url)
	urlData := strings.Split(outputUrl, " ")
	method, url := urlData[0], urlData[1]

	accessToken := resOutput.AccessToken
	expiration := resOutput.Expiration

	fmt.Println(url, accessToken, expiration)
	return method, HTTP + url, expiration
}

func makeQueryParams(limit, facility, minSeverity int, logServiceId, containerId, mode string) string {
	var desc = 1
	if mode == "RESPONSE" {
		desc = 0
	}

	if containerId != "" { // FIXME severity is broken on API side, debug level won't work (looks as reverse numbers used)
		// FIXME serviceStackId can be probably removed
		// example from UI "&limit=100&containerId=PfA02HAkTvSShzvrirmYew&desc=1"
		return fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s&containerId=%s&minimumSeverity=%d",
			limit, desc, facility, logServiceId, containerId, minSeverity)
	}
	// example from UI &serviceStackId=N05ZvyGUSjAyxBN6E8zBog&limit=6000&desc=0
	return fmt.Sprintf("&limit=%d&desc=%d&facility=%d&serviceStackId=%s&minimumSeverity=%d",
		limit, desc, facility, logServiceId, minSeverity,
	)
}
