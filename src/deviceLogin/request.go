package deviceLogin

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/errorCode"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

type request struct {
	id         uuid.DeviceLoginRequestId
	pollSecret types.String
	userCode   string
	url        string
	expiresIn  time.Duration
	interval   time.Duration
	// checkTimeout bounds one check call, the client has no HTTP timeout and a black-holed request would block until the OS gives up.
	checkTimeout time.Duration
}

func initiate(ctx context.Context, f flow, cfg Config) (request, error) {
	out, err := f.start(ctx, body.DeviceLoginInitiate{AppCode: types.NewString(cfg.AppCode)})
	if err != nil {
		return request{}, err
	}
	interval := time.Duration(max(out.Interval.Native(), 1)) * cfg.unit
	return request{
		id:           out.RequestId,
		pollSecret:   out.PollSecret,
		userCode:     out.UserCode.Native(),
		url:          out.VerificationUrl.Native(),
		expiresIn:    time.Duration(out.ExpiresIn.Native()) * cfg.unit,
		interval:     interval,
		checkTimeout: max(3*interval, 10*cfg.unit),
	}, nil
}

type pollOutcome int

const (
	pollPending pollOutcome = iota
	// pollRetry is a 5xx or any error without a decoded api error (transport, timeout, non-JSON body), polling continues.
	pollRetry
	pollApproved
	pollDenied
	pollExpired
	// pollFailed ends polling with err.
	pollFailed
)

type pollResult struct {
	outcome pollOutcome
	auth    output.AuthLogin
	err     error
}

func check(ctx context.Context, f flow, req request) pollResult {
	callCtx, cancel := context.WithTimeout(ctx, req.checkTimeout)
	defer cancel()
	resp, err := f.checker.PostAuthDeviceLoginCheck(callCtx, body.DeviceLoginCheck{RequestId: req.id, PollSecret: req.pollSecret})
	if err != nil {
		if ctx.Err() != nil {
			return pollResult{outcome: pollFailed, err: ctx.Err()}
		}
		return pollResult{outcome: pollRetry, err: err}
	}
	out, err := resp.Output()
	if err != nil {
		switch {
		case apiError.HasErrorCode(err, errorCode.DeviceLoginNotFound):
			return pollResult{outcome: pollExpired}
		case isServerError(err):
			return pollResult{outcome: pollRetry, err: err}
		default:
			return pollResult{outcome: pollFailed, err: err}
		}
	}
	switch out.Status {
	case enum.DeviceLoginStatusEnumPending:
		return pollResult{outcome: pollPending}
	case enum.DeviceLoginStatusEnumApproved:
		if f.sudo {
			return pollResult{outcome: pollApproved}
		}
		if out.Auth == nil {
			return pollResult{outcome: pollFailed, err: errors.New("device login approved without auth")}
		}
		return pollResult{outcome: pollApproved, auth: *out.Auth}
	case enum.DeviceLoginStatusEnumDenied:
		return pollResult{outcome: pollDenied}
	default:
		return pollResult{outcome: pollFailed, err: errors.Errorf("unexpected device login status %q", out.Status)}
	}
}

func isServerError(err error) bool {
	var apiErr apiError.Error
	return errors.As(err, &apiErr) && apiErr.GetHttpStatusCode() >= http.StatusInternalServerError
}
