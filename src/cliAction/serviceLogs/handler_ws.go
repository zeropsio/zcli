package serviceLogs

import (
	"context"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/types"
)

func getLogStream(_ context.Context, expiration types.DateTime, format, serviceId, url string) error {
	if format == JSON {
		return fmt.Errorf("%s", i18n.LogFormatStreamMismatch)
	}

	fmt.Printf("stream with:\n expiration %s\n for serviceId %s\n in format %s\n url is %s\n", expiration, serviceId, format, url)
	// todo add websocket
	return nil
}
