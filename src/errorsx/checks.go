package errorsx

import (
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

type ErrorCodeExtractMessageFunc func(apiError.Error) string

func extractApiErrorMessage(err apiError.Error) string {
	return err.GetMessage()
}

type ErrorCodeExtractMessageMetaFunc func(apiError.Error, map[string]interface{}) string

func extractApiErrorMessageMeta(err apiError.Error, _ map[string]interface{}) string {
	return err.GetMessage()
}

type errorCodeConfig struct {
	errorMessage ErrorCodeExtractMessageFunc
}

var defaultErrorCodeConfig = errorCodeConfig{
	errorMessage: extractApiErrorMessage,
}

type ErrorCodeOption generic.Option[errorCodeConfig]

func ErrorCodeErrorMessage(f ErrorCodeExtractMessageFunc) ErrorCodeOption {
	return func(cfg *errorCodeConfig) {
		cfg.errorMessage = f
	}
}

func ErrorCode(errorCode errorCode.ErrorCode, auxOptions ...ErrorCodeOption) Check {
	return func(err error) error {
		cfg := generic.ApplyOptionsWithDefault(defaultErrorCodeConfig, auxOptions...)

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if string(errorCode) != apiErr.GetErrorCode() {
			return nil
		}

		return NewUserError(cfg.errorMessage(apiErr), err)
	}
}

type httpStatusCodeConfig struct {
	errorMessage ErrorCodeExtractMessageFunc
}

var defaultHttpStatusCodeConfig = httpStatusCodeConfig{
	errorMessage: extractApiErrorMessage,
}

func HttpStatusCodeErrorMessage(f ErrorCodeExtractMessageFunc) HttpStatusCodeOption {
	return func(cfg *httpStatusCodeConfig) {
		cfg.errorMessage = f
	}
}

type HttpStatusCodeOption generic.Option[httpStatusCodeConfig]

func HttpStatusCode(httpStatusCode int, auxOptions ...HttpStatusCodeOption) Check {
	return func(err error) error {
		cfg := generic.ApplyOptionsWithDefault(defaultHttpStatusCodeConfig, auxOptions...)

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if httpStatusCode != apiErr.GetHttpStatusCode() {
			return nil
		}

		return NewUserError(cfg.errorMessage(apiErr), err)
	}
}

type invalidUserInputConfig struct {
	errorMessage ErrorCodeExtractMessageMetaFunc
}

var defaultInvalidUserInputConfig = invalidUserInputConfig{
	errorMessage: extractApiErrorMessageMeta,
}

func InvalidUserInputErrorMessage(f ErrorCodeExtractMessageMetaFunc) InvalidUserInputOption {
	return func(cfg *invalidUserInputConfig) {
		cfg.errorMessage = f
	}
}

type InvalidUserInputOption generic.Option[invalidUserInputConfig]

func InvalidUserInput(parameterName string, auxOptions ...InvalidUserInputOption) Check {
	return func(err error) error {
		cfg := generic.ApplyOptionsWithDefault(defaultInvalidUserInputConfig, auxOptions...)

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
						return NewUserError(cfg.errorMessage(apiErr, metaItemTyped), err)
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
