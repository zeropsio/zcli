package serviceLogs

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/zeropsio/zcli/src/i18n"
)

func getFullWithTemplate(logData []Data, formatTemplate string) error {
	if len(logData) == 0 {
		return NewInvalidRequestError("getFullWithTemplate", "no log data available", nil)
	}

	ft, err := fixTemplate(formatTemplate)
	if err != nil {
		return err
	}
	for _, o := range logData {
		err := formatDataByTemplate(o, ft)
		if err != nil {
			return errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), err)
		}
	}
	return nil
}

func formatDataByTemplate(data Data, formatTemplate string) error {
	var b bytes.Buffer
	t, err := template.New("").Parse(formatTemplate)
	if err != nil {
		return NewInvalidRequestError("formatDataByTemplate", "failed to parse template", err)
	}
	
	if err = t.Execute(&b, data); err != nil {
		return NewInvalidRequestError("formatDataByTemplate", "failed to execute template", err)
	}
	
	fmt.Println(b.String())
	return nil
}

// test if there are any merged template items and return error
func testTokens(tokens []string) error {
	for _, token := range tokens {
		// if any `{` characters are left, it means the items were not split by correctly
		rmLeft := strings.Replace(token, "{", "", 2)
		if strings.Contains(rmLeft, "{") {
			return errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), i18n.T(i18n.LogFormatTemplateNoSpace))
		}
	}
	return nil
}

// change lowercase first letters to uppercase to match the struct
func fixTemplate(template string) (string, error) {
	tokens := strings.Split(template, "} {")
	if err := testTokens(tokens); err != nil {
		return "", err
	}
	repl := strings.NewReplacer("{", "", "}", "", ".", "")
	keys := make([]string, len(tokens))

	for i, val := range tokens {
		out := strings.Trim(repl.Replace(val), " ")
		titleStr := cases.Title(language.Und, cases.NoLower).String(out)
		item := fmt.Sprintf("{{.%s}}", titleStr)
		keys[i] = item
	}

	return strings.Join(keys, " "), nil
}
