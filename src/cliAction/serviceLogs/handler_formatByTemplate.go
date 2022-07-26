package serviceLogs

import (
	"bytes"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
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

// change lowercase first letters to uppercase to match the struct
func templateFix(template string) string {
	repl := strings.NewReplacer("{", "", "}", "", ".", "")
	out := repl.Replace(template)
	tokens := strings.Split(out, " ")

	keys := make([]string, len(tokens))

	for i, val := range tokens {
		titleStr := cases.Title(language.Und, cases.NoLower).String(val)
		item := fmt.Sprintf("{{.%s}}", titleStr)
		keys[i] = item
	}

	return strings.Join(keys, " ")
}
