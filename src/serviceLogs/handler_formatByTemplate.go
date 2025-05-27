package serviceLogs

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) getFullWithTemplate(logData []Data, formatTemplate string) error {
	for _, o := range logData {
		err := h.formatDataByTemplate(o, formatTemplate)
		if err != nil {
			return errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), err)
		}
	}
	return nil
}

func (h *Handler) formatDataByTemplate(data Data, formatTemplate string) error {
	var b bytes.Buffer
	t, err := template.New("").Parse(formatTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(&b, data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(h.out, b.String())
	return err
}
