package proto

import (
	"errors"
	"fmt"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/proto/vpnproxy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandleGrpcErrorOption func(*handleGrpcErrorConfig)

type handleGrpcErrorConfig struct {
	customTimeoutMessage string
}

func BusinessError(
	response interface {
		GetError() *business.Error
	},
	err error,
	options ...HandleGrpcErrorOption,
) error {
	config := handleGrpcErrorConfig{
		customTimeoutMessage: i18n.GrpcApiTimeout,
	}
	for _, o := range options {
		o(&config)
	}

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.DeadlineExceeded && config.customTimeoutMessage != "" {
				return errors.New(config.customTimeoutMessage)
			} else {
				return errors.New(s.Message())
			}
		}
		return err
	}
	if response.GetError().GetCode() != business.ErrorCode_NO_ERROR {
		return fmt.Errorf("%s [%s]", response.GetError().GetMessage(), string(response.GetError().GetMeta()))
	}

	return nil
}

func VpnError(
	response interface {
		GetError() *vpnproxy.Error
	},
	err error,
	options ...HandleGrpcErrorOption,

) error {
	config := handleGrpcErrorConfig{
		customTimeoutMessage: i18n.GrpcVpnApiTimeout,
	}
	for _, o := range options {
		o(&config)
	}

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.DeadlineExceeded && config.customTimeoutMessage != "" {
				return errors.New(config.customTimeoutMessage)
			} else {
				return errors.New(s.Message())
			}
		}
	}
	if response.GetError().GetCode() != vpnproxy.ErrorCode_NO_ERROR {
		return errors.New(response.GetError().GetMessage())
	}

	return nil
}

func DaemonError(
	err error,
) (daemonInstalled bool, _ error) {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unavailable {
				return false, nil
			}
			return true, errors.New(s.Message())
		}
		return true, err
	}
	return true, nil
}
