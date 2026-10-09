package deviceLogin

import (
	"context"
	"time"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/uxBlock/styles"
	"github.com/zeropsio/zerops-go/dto/output"
)

func runPlain(ctx context.Context, f flow, cfg Config) (output.AuthLogin, error) {
	req, err := initiate(ctx, f, cfg)
	if err != nil {
		return output.AuthLogin{}, err
	}
	cfg.UxBlocks.PrintInfoText(i18n.T(i18n.DeviceLoginConfirmCode))
	cfg.UxBlocks.PrintInfo(styles.NewLine(styles.InfoText(styles.InfoIcon), " ", styles.InfoPrefix(), " ", styles.SelectText(req.userCode)))
	cfg.UxBlocks.PrintInfoText(req.url)
	cfg.UxBlocks.PrintInfoText(openBrowser(cfg, req.url))

	timer := time.NewTimer(req.interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return output.AuthLogin{}, ctx.Err()
		case <-timer.C:
		}
		res := check(ctx, f, req)
		switch res.outcome {
		case pollApproved:
			return res.auth, nil
		case pollDenied:
			return output.AuthLogin{}, deniedError(f)
		case pollExpired:
			return output.AuthLogin{}, expiredError()
		case pollFailed:
			return output.AuthLogin{}, res.err
		case pollRetry:
			cfg.UxBlocks.LogDebug("device login: check failed, retrying: " + res.err.Error())
		case pollPending:
		}
		timer.Reset(req.interval)
	}
}
