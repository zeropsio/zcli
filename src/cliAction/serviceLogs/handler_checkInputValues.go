package serviceLogs

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zerops-io/zcli/src/i18n"
)

type InputValues struct {
	limit          int
	minSeverity    int
	facility       int
	format         string
	formatTemplate string
	mode           string
}

func (h *Handler) checkInputValues(config RunConfig) (inputValues InputValues, err error) {
	limit, err := h.getLimit(config)
	if err != nil {
		return inputValues, err
	}
	severity, err := h.getMinSeverity(config)
	if err != nil {
		return inputValues, err
	}
	facility, err := h.getFacility(config)
	if err != nil {
		return inputValues, err
	}
	format, formatTemplate, err := h.getFormat(config)
	if err != nil {
		return inputValues, err
	}

	mode := RESPONSE
	if config.Follow {
		mode = STREAM
	}
	return InputValues{
		limit:          int(limit),
		minSeverity:    severity,
		facility:       facility,
		format:         format,
		formatTemplate: formatTemplate,
		mode:           mode,
	}, nil
}

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
	if ms == "" {
		// -1 for min.severity not required by user, used to make query
		return -1, nil
	}

	for key, val := range config.Levels {
		if strings.ToUpper(ms) == val[0] || ms == val[1] {
			return key, nil
		}
	}
	intVal, err = strconv.Atoi(ms)
	if err != nil {
		return 1, fmt.Errorf("%s %s", i18n.LogMinSeverityInvalid, i18n.LogMinSeverityStringLimitErr)
	}
	return 1, fmt.Errorf("%s %s", i18n.LogMinSeverityInvalid, i18n.LogMinSeverityNumLimitErr)
}

// getFacility returns facility number based on msgType
func (h *Handler) getFacility(config RunConfig) (int, error) {
	mt := strings.ToUpper(config.MsgType)
	if mt == APPLICATION {
		return 16, nil
	}
	if mt == WEBSERVER {
		return 17, nil
	}
	return 16, fmt.Errorf("%s", i18n.LogMsgTypeInvalid)
}

func (h *Handler) getFormat(config RunConfig) (string, string, error) {
	f, ft := strings.ToUpper(config.Format), config.FormatTemplate
	formatValid := f == FULL || f == SHORT || f == JSON || f == JSONSTREAM
	if !formatValid {
		return "", "", fmt.Errorf("%s", i18n.LogFormatInvalid)
	}
	if ft == "" {
		return f, ft, nil
	}
	if f != FULL {
		return "", "", fmt.Errorf("%s", i18n.LogFormatTemplateMismatch)
	}
	formatTemplate, err := h.checkFormat(ft)
	if err != nil {
		return "", "", err
	}
	return f, formatTemplate, nil
}

// e.g. --formatTemplate="{{.timestamp}} {{.priority}} {{.facility}} {{.message}}"
func (h *Handler) checkFormat(ft string) (string, error) {
	if err := validateTemplate(ft); err != nil {
		return "", fmt.Errorf("%s %s", i18n.LogFormatTemplateInvalid, err)
	}
	return ft, nil
}

func validateTemplate(s string) error {
	_, err := template.New("").Parse(s)
	return err
}
