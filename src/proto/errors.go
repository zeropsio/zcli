package proto

import (
	"errors"
	"fmt"

	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/proto/vpnproxy"

	"github.com/zerops-io/zcli/src/i18n"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandleGrpcErrorOption func(*handleGrpcErrorConfig)

type handleGrpcErrorConfig struct {
	customTimeoutMessage string
}

type errorCode interface {
	GetCodeInt() int
	GetMeta() []byte
	GetMessage() string
}

type response[E errorCode] interface {
	GetError() E
}

func BusinessError[R response[*business.Error]](
	resp R,
	err error,
	options ...HandleGrpcErrorOption,
) error {
	return GrpcError[*business.Error](resp, err, options...)
}

func VpnError[R response[*vpnproxy.Error]](
	resp R,
	err error,
	options ...HandleGrpcErrorOption,
) error {
	return GrpcError[*vpnproxy.Error](resp, err, options...)
}

func GrpcError[T errorCode, R response[T]](
	resp R,
	err error,
	options ...HandleGrpcErrorOption,
) error {
	config := handleGrpcErrorConfig{
		customTimeoutMessage: i18n.GrpcApiTimeout,
	}
	for _, o := range options {
		o(&config)
	}

	if err := handleGrpcError(err, config); err != nil {
		return err
	}

	noErrorCode := 0
	if resp.GetError().GetCodeInt() != noErrorCode {
		err := resp.GetError()
		if len(err.GetMeta()) == 0 {
			return fmt.Errorf("%s [%s]", err.GetMessage(), string(err.GetMeta()))
		}
		return fmt.Errorf("%s", err.GetMessage())
	}

	return nil
}

type Err struct {
	Msg string
	*status.Status
}

func (e Err) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Status.Message()
}

func IsUnauthenticated(err error) bool {
	var e Err
	if errors.As(err, &e) {
		return e.Status.Code() == codes.Unauthenticated
	}
	return false
}

func handleGrpcError(err error, config handleGrpcErrorConfig) error {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			err := Err{Status: s}
			if s.Code() == codes.DeadlineExceeded {
				err.Msg = config.customTimeoutMessage
			}
			return err
		}
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
