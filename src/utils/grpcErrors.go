package utils

import (
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
	"github.com/zerops-io/zcli/src/zeropsVpnProtocol"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WithCustomTimeoutMessage(message string) HandleGrpcErrorOption {
	return func(config *handleGrpcErrorConfig) {
		config.customTimeoutMessage = message
	}
}

type HandleGrpcErrorOption func(*handleGrpcErrorConfig)

type handleGrpcErrorConfig struct {
	customTimeoutMessage string
}

func HandleGrpcApiError(
	response interface {
		GetError() *zeropsApiProtocol.Error
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
	if response.GetError().GetCode() != zeropsApiProtocol.ErrorCode_NO_ERROR {

		return errors.New(fmt.Sprintf("%s [%s]", response.GetError().GetMessage(), string(response.GetError().GetMeta())))
	}

	return nil
}

func HandleVpnApiError(
	response interface {
		GetError() *zeropsVpnProtocol.Error
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
	if response.GetError().GetCode() != zeropsVpnProtocol.ErrorCode_NO_ERROR {
		return errors.New(response.GetError().GetMessage())
	}

	return nil
}

func HandleDaemonError(
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
