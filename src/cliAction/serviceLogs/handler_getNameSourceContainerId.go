package serviceLogs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zeropsio/zcli/src/i18n"
)

func (h *Handler) getNameSourceContainerId(config RunConfig) (serviceName, source string, containerId int, err error) {
	sn := config.ServiceName
	source = RUNTIME

	if !strings.Contains(sn, AT) {
		return sn, source, 0, nil
	}
	split := strings.Split(sn, AT)
	if len(split) > 2 {
		return "", "", 0, fmt.Errorf("%s", i18n.LogServiceNameInvalid)
	}
	sn = split[0]
	suffix := split[1]

	if strings.Contains(suffix, AT) {
		return "", "", 0, fmt.Errorf("%s", i18n.LogServiceNameInvalid)
	}

	if suffix == "" {
		return sn, source, 0, nil
	}

	containerIndex, err := strconv.Atoi(suffix)
	if err == nil {
		if containerIndex < 1 {
			return "", "", 0, fmt.Errorf("%s", i18n.LogSuffixInvalid)
		}
		return sn, source, containerIndex, nil
	}

	if strings.ToUpper(suffix) != BUILD {
		return "", "", 0, fmt.Errorf("%s", i18n.LogSuffixInvalid)
	}
	source = BUILD
	return sn, source, 0, nil
}
