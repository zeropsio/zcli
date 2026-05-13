package serviceLogs

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) getFullWithTemplate(logData []Data, formatTemplate string) error {
	ft, err := fixTemplate(formatTemplate)
	if err != nil {
		return err
	}
	tmpl, err := template.New("").Parse(ft)
	if err != nil {
		return errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), err)
	}
	for _, o := range logData {
		if err := h.formatDataByTemplate(o, tmpl); err != nil {
			return errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), err)
		}
	}
	return nil
}

func (h *Handler) formatDataByTemplate(data Data, tmpl *template.Template) error {
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return err
	}
	_, err := fmt.Fprintln(h.out, b.String())
	return err
}

// test if there are any merged template items and return an error
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

// fixTemplate normalizes Data field names in the user's format string to
// match Go struct casing — e.g. `{{.message}}` becomes `{{.Message}}` —
// so the renderer doesn't have to require correctly-cased input.
func fixTemplate(tmpl string) (string, error) {
	tokens := strings.Split(tmpl, "} {")
	if err := testTokens(tokens); err != nil {
		return "", err
	}
	repl := strings.NewReplacer("{", "", "}", "", ".", "")
	keys := make([]string, len(tokens))
	for i, val := range tokens {
		out := strings.TrimSpace(repl.Replace(val))
		keys[i] = fmt.Sprintf("{{.%s}}", upperFirst(out))
	}
	return strings.Join(keys, " "), nil
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
