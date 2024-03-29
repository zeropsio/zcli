package errorsx

import (
	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

type errorCodeConfig struct {
	errorMessage func(apiError.Error) string
}

func ErrorCodeErrorMessage(f func(apiErr apiError.Error) string) errorCodeOption {
	return func(cfg *errorCodeConfig) {
		cfg.errorMessage = f
	}
}

type errorCodeOption = func(cfg *errorCodeConfig)

func ErrorCode(errorCode errorCode.ErrorCode, auxOptions ...errorCodeOption) Check {
	return func(err error) error {
		cfg := errorCodeConfig{}
		for _, opt := range auxOptions {
			opt(&cfg)
		}

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if string(errorCode) != apiErr.GetErrorCode() {
			return nil
		}

		if cfg.errorMessage != nil {
			return NewUserError(cfg.errorMessage(apiErr), err)
		} else {
			return NewUserError(apiErr.GetMessage(), err)
		}
	}
}

type httpStatusCodeConfig struct {
	errorMessage func(apiError.Error) string
}

func HttpStatusCodeErrorMessage(f func(apiErr apiError.Error) string) httpStatusCodeOption {
	return func(cfg *httpStatusCodeConfig) {
		cfg.errorMessage = f
	}
}

type httpStatusCodeOption = func(cfg *httpStatusCodeConfig)

func HttpStatusCode(httpStatusCode int, auxOptions ...httpStatusCodeOption) Check {
	return func(err error) error {
		cfg := httpStatusCodeConfig{}
		for _, opt := range auxOptions {
			opt(&cfg)
		}

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if httpStatusCode != apiErr.GetHttpStatusCode() {
			return nil
		}

		if cfg.errorMessage != nil {
			return NewUserError(cfg.errorMessage(apiErr), err)
		} else {
			return NewUserError(apiErr.GetMessage(), err)
		}
	}
}

type invalidUserInputConfig struct {
	errorMessage func(apiError.Error, map[string]interface{}) string
}

func InvalidUserInputErrorMessage(
	f func(apiErr apiError.Error, metaItem map[string]interface{}) string,
) invalidUserInputOption {
	return func(cfg *invalidUserInputConfig) {
		cfg.errorMessage = f
	}
}

type invalidUserInputOption = func(cfg *invalidUserInputConfig)

func InvalidUserInput(parameterName string, auxOptions ...invalidUserInputOption) Check {
	return func(err error) error {
		cfg := invalidUserInputConfig{}
		for _, opt := range auxOptions {
			opt(&cfg)
		}

		if err := ErrorCode(errorCode.InvalidUserInput)(err); err == nil {
			return nil
		}

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		meta, ok := apiErr.GetMeta().([]interface{})
		if !ok {
			return nil
		}

		for _, metaItem := range meta {
			if metaItemTyped, ok := metaItem.(map[string]interface{}); ok {
				if metaParameterName, ok := metaItemTyped["parameter"]; ok {
					if metaParameterName == parameterName {
						if cfg.errorMessage != nil {
							return NewUserError(cfg.errorMessage(apiErr, metaItemTyped), err)
						} else {
							return NewUserError(apiErr.GetMessage(), err)
						}
					}
				}
			}
		}

		return nil
	}
}

func Meta(
	errorMessage func(apiErr apiError.Error, metaItem map[string]interface{}) string,
) Check {
	return func(err error) error {
		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if metaTyped, ok := apiErr.GetMeta().(map[string]interface{}); ok {
			return NewUserError(errorMessage(apiErr, metaTyped), err)
		}

		return nil
	}
}
