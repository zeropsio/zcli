package serviceLogs

import (
	"context"
	"fmt"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetLogs(ctx context.Context, method, url string) error {
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

	fmt.Println(string(body))

	if err != nil {
		return err
	}
	return nil
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

func makeQueryParams(limit, facility, minSeverity int, logServiceId, containerId string) string {
	if containerId != "" {
		return fmt.Sprintf("&limit=%d&desc=1&facility=%d&serviceStackId=%s&containerId=%s&minimumSeverity=%d",
			limit, facility, logServiceId, containerId, minSeverity)
	}
	return fmt.Sprintf("&limit=%d&desc=1&facility=%d&serviceStackId=%s&minimumSeverity=%d",
		limit, facility, logServiceId, minSeverity)
}
