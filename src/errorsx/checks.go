package errorsx

import (
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zerops-go/apiError"
)

type ErrorCodeExtractMessageFunc func(ApiError) string

func extractApiErrorMessage(err ApiError) string {
	return err.Message()
}

type ErrorCodeExtractMessageMetaFunc func(ApiError, map[string]interface{}) string

func extractApiErrorMessageMeta(err ApiError, _ map[string]interface{}) string {
	return err.Message()
}

type errorCodeConfig struct {
	errorMessage ErrorCodeExtractMessageFunc
}

var defaultErrorCodeConfig = errorCodeConfig{
	errorMessage: extractApiErrorMessage,
}

type ErrorCodeOption gn.Option[errorCodeConfig]

func ErrorCodeErrorMessage(f ErrorCodeExtractMessageFunc) ErrorCodeOption {
	return func(cfg *errorCodeConfig) {
		cfg.errorMessage = f
	}
}

func ErrorCode(code Code, auxOptions ...ErrorCodeOption) Check {
	return func(err error) error {
		cfg := gn.ApplyOptionsWithDefault(defaultErrorCodeConfig, auxOptions...)

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if string(code) != apiErr.GetErrorCode() {
			return nil
		}

		return NewUserError(cfg.errorMessage(apiErrorView{err: apiErr}), err)
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

type HttpStatusCodeOption gn.Option[httpStatusCodeConfig]

func HttpStatusCode(httpStatusCode int, auxOptions ...HttpStatusCodeOption) Check {
	return func(err error) error {
		cfg := gn.ApplyOptionsWithDefault(defaultHttpStatusCodeConfig, auxOptions...)

		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if httpStatusCode != apiErr.GetHttpStatusCode() {
			return nil
		}

		return NewUserError(cfg.errorMessage(apiErrorView{err: apiErr}), err)
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

type InvalidUserInputOption gn.Option[invalidUserInputConfig]

func InvalidUserInput(parameterName string, auxOptions ...InvalidUserInputOption) Check {
	return func(err error) error {
		cfg := gn.ApplyOptionsWithDefault(defaultInvalidUserInputConfig, auxOptions...)

		if err := ErrorCode(CodeInvalidUserInput)(err); err == nil {
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
						return NewUserError(cfg.errorMessage(apiErrorView{err: apiErr}, metaItemTyped), err)
					}
				}
			}
		}

		return nil
	}
}

func Meta(
	errorMessage func(apiErr ApiError, metaItem map[string]interface{}) string,
) Check {
	return func(err error) error {
		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

		if metaTyped, ok := apiErr.GetMeta().(map[string]interface{}); ok {
			return NewUserError(errorMessage(apiErrorView{err: apiErr}, metaTyped), err)
		}

		return nil
	}
}
