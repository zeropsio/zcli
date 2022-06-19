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
