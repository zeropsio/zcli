package serviceLogs

import (
	"bytes"
	"fmt"
	"text/template"
)

func getFullWithTemplate(jsonData Response, formatTemplate string) error {
	for _, o := range jsonData.Items {
		err := formatDataByTemplate(o, formatTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func formatDataByTemplate(data Data, formatTemplate string) error {
	var b bytes.Buffer
	t, err := template.New("").Parse(formatTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&b, data)
	if err != nil {
		return err
	}

	fmt.Println(b.String())
	return nil
}
