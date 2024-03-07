package serviceLogs

import (
	"encoding/json"
	"fmt"
)

func parseResponseByFormat(jsonData Response, format, formatTemplate, mode string) error {
	var err error

	logs := jsonData.Items
	if mode == RESPONSE {
		logs = reverseLogs(logs)
	}

	switch format {
	case FULL:
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
	case SHORT:
		for _, o := range logs {
			fmt.Printf("%v %s \n", o.Timestamp, o.Content)
		}
	case JSONSTREAM:
		for _, o := range logs {
			val, err := json.Marshal(o)
			if err != nil {
				return err
			}
			fmt.Println(string(val))
		}
	default:
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
