package serviceLogs

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/i18n"
)

type InputValues struct {
	limit          int
	minSeverity    int
	facility       int
	format         string
	formatTemplate string
	mode           string
	tags           []string
}

func (h *Handler) checkInputValues(config RunConfig) (InputValues, error) {
	limit, err := h.getLimit(config)
	if err != nil {
		return InputValues{}, err
	}
	severity, err := h.getMinSeverity(config)
	if err != nil {
		return InputValues{}, err
	}
	facility, err := h.getFacility(config)
	if err != nil {
		return InputValues{}, err
	}
	format, formatTemplate, err := h.getFormat(config)
	if err != nil {
		return InputValues{}, err
	}

	mode := RESPONSE
	if config.Follow {
		mode = STREAM
		if format == JSON {
			return InputValues{}, errors.New(i18n.T(i18n.LogFormatStreamMismatch))
		}
	}
	return InputValues{
		limit:          limit,
		minSeverity:    severity,
		facility:       facility,
		format:         format,
		formatTemplate: formatTemplate,
		mode:           mode,
		tags:           config.Tags,
	}, nil
}

func (h *Handler) getLimit(config RunConfig) (int, error) {
	if config.Limit < 1 || config.Limit > 1000 {
		return 0, errors.New(i18n.T(i18n.LogLimitInvalid))
	}
	return config.Limit, nil
}

// severityLevels maps severity numbers (0..7) to their RFC 5424 keyword.
// Indexed by severity number; the keyword and the string form of the
// index are both accepted as input.
var severityLevels = [...]string{
	"EMERGENCY",
	"ALERT",
	"CRITICAL",
	"ERROR",
	"WARNING",
	"NOTICE",
	"INFORMATIONAL",
	"DEBUG",
}

func (h *Handler) getMinSeverity(config RunConfig) (int, error) {
	ms := config.MinSeverity
	if ms == "" {
		// -1 signals "no minimum severity"; the query param is then omitted.
		return -1, nil
	}
	upper := strings.ToUpper(ms)
	for i, name := range severityLevels {
		if upper == name || ms == strconv.Itoa(i) {
			return i, nil
		}
	}
	if _, err := strconv.Atoi(ms); err == nil {
		return 0, errors.Errorf("%s %s", i18n.T(i18n.LogMinSeverityInvalid), i18n.T(i18n.LogMinSeverityNumLimitErr))
	}
	return 0, errors.Errorf("%s %s", i18n.T(i18n.LogMinSeverityInvalid), i18n.T(i18n.LogMinSeverityStringLimitErr))
}

// getFacility returns the syslog facility number for the given message type.
func (h *Handler) getFacility(config RunConfig) (int, error) {
	switch strings.ToUpper(config.MsgType) {
	case APPLICATION:
		return 16, nil
	case WEBSERVER:
		return 17, nil
	}
	return 0, errors.New(i18n.T(i18n.LogMsgTypeInvalid))
}

func (h *Handler) getFormat(config RunConfig) (string, string, error) {
	f, ft := strings.ToUpper(config.Format), config.FormatTemplate
	switch f {
	case FULL, SHORT, JSON, JSONSTREAM:
	default:
		return "", "", errors.New(i18n.T(i18n.LogFormatInvalid))
	}
	if ft == "" {
		return f, ft, nil
	}
	if f != FULL {
		return "", "", errors.New(i18n.T(i18n.LogFormatTemplateMismatch))
	}
	if _, err := template.New("").Parse(ft); err != nil {
		return "", "", errors.Errorf("%s %s", i18n.T(i18n.LogFormatTemplateInvalid), err)
	}
	return f, ft, nil
}
