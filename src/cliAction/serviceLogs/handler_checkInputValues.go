package serviceLogs

import (
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"strconv"
)

//(limit int, severity int, msgType string, format string, formatTemplate string, err error)
func (h *Handler) getLimit(config RunConfig) (limit uint32, err error) {
	limit = config.Limit

	if limit < 1 || limit > 1000 {
		err = fmt.Errorf("%s", i18n.LogLimitInvalid)
		return limit, err
	}

	return limit, nil
}

func (h *Handler) getMinSeverity(config RunConfig) (intVal int, err error) {
	ms := config.MinSeverity

	for key, val := range config.Levels {
		if ms == val[0] || ms == val[1] {
			return key, nil
		}
	}
	intVal, err = strconv.Atoi(ms)
	if err != nil {
		return 1, fmt.Errorf("%s %s", i18n.LogMinSeverityInvalid, i18n.LogMinSeverityStringLimitErr)
	}
	return 1, fmt.Errorf("%s %s", i18n.LogMinSeverityInvalid, i18n.LogMinSeverityNumLimitErr)
}

func (h *Handler) getFacility(config RunConfig) (int, error) {
	mt := config.MsgType
	if mt == "APPLICATION" {
		return 16, nil
	}
	if mt == "WEBSERVER" {
		return 17, nil
	}
	return 16, fmt.Errorf("%s", i18n.LogMsgTypeInvalid)
}

func (h *Handler) getFormat(config RunConfig) (string, string, error) {
	f, ft := config.Format, config.FormatTemplate
	formatValid := f == "FULL" || f == "SHORT" || f == "JSON"
	if !formatValid {
		return "", "", fmt.Errorf("%s", i18n.LogFormatTemplateInvalid)
	}
	if ft == "" {
		return f, ft, nil
	}
	if f != "FULL" {
		return "", "", fmt.Errorf("%s", i18n.LogFormatTemplateMismatch)
	}
	template, err := h.createFormat(ft)
	if err != nil {
		return "", "", err
	}
	return f, template, nil
}

//IF the custom template fails to be created, return the error
//     "Invalid --formatTemplate content. The custom template failed with following error: {error message returned by GoLang
//   template}"
func (h *Handler) createFormat(ft string) (string, error) {
	// TODO see https://pkg.go.dev/text/template
	// e.g. --formatTemplate="{{.timestamp}} {{.priority}} {{.facility}} {{.message}}"
	return ft, nil
}
