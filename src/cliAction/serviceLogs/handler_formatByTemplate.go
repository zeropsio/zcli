package serviceLogs

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/zerops-io/zcli/src/i18n"
)

func getFullWithTemplate(logData []Data, formatTemplate string) error {
	for _, o := range logData {
		err := formatDataByTemplate(o, formatTemplate)
		if err != nil {
			return fmt.Errorf("%s %s", i18n.LogFormatTemplateInvalid, err)
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
