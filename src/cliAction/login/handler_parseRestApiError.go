package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zeropsio/zcli/src/i18n"
)

type failResponse struct {
	Error *failResponseError `json:"error"`
}

func (f failResponse) IsUnauthenticated() bool {
	code := f.Error.ErrorCode
	return code == userNotFound || code == incorrectUserCredentials
}

type failResponseError struct {
	ErrorCode string      `json:"code"`
	Message   string      `json:"message"`
	Meta      interface{} `json:"meta,omitempty"`
}

func parseRestApiError(body []byte) error {
	var errorResponse failResponse
	err := json.Unmarshal(body, &errorResponse)
	if err != nil {
		return err
	}
	if errorResponse.Error.ErrorCode == "invalidUserInput" {
		var errorList []string
		if metaList, ok := errorResponse.Error.Meta.([]interface{}); ok {
			for _, meta := range metaList {
				if metaItem, ok := meta.(map[string]interface{}); ok {
					if parameter, exists := metaItem["parameter"]; exists {
						if message, exists := metaItem["message"]; exists {
							if p, ok := parameter.(string); ok {
								if m, ok := message.(string); ok {
									errorList = append(errorList, fmt.Sprintf("'%s': %s", p, m))
								}
							}
						}
					}
				}
			}
		}

		return errors.New(strings.Join(errorList, ", "))
	} else {
		err := errors.New(errorResponse.Error.Message)
		if errorResponse.IsUnauthenticated() {
			return i18n.AddHintChangeRegion(err)
		}
		return err
	}
}

const userNotFound = "userNotFound"
const incorrectUserCredentials = "incorrectUserCredentials"
