package zeropsRestApiClient

import (
	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

type checkErrorConfig struct {
	checks []func(error) error
}

func CheckInvalidUserInput(paramName string, returnedErr error) CheckErrorOption {
	return func(cfg *checkErrorConfig) {
		cfg.checks = append(cfg.checks, func(err error) error {
			var apiErr apiError.Error
			if errors.As(err, &apiErr) {
				if apiErr.GetErrorCode() == string(errorCode.InvalidUserInput) {
					if typedMeta, ok := apiErr.GetMeta().([]interface{}); ok {
						if len(typedMeta) > 0 {
							if typed, ok := typedMeta[0].(map[string]interface{}); ok {
								if param, exists := typed["parameter"]; exists {
									if value, ok := param.(string); ok && value == paramName {
										return returnedErr
									}
								}
							}
						}
					}
				}
			}

			return nil
		})
	}
}

func CheckErrorCode(errorCode errorCode.ErrorCode, returnedErr error) CheckErrorOption {
	return func(cfg *checkErrorConfig) {
		cfg.checks = append(cfg.checks, func(err error) error {
			var apiErr apiError.Error
			if errors.As(err, &apiErr) {
				if apiErr.GetErrorCode() == string(errorCode) {
					return returnedErr
				}
			}

			return nil
		})
	}
}

type CheckErrorOption = func(cfg *checkErrorConfig)

func CheckError(err error, auxOptions ...CheckErrorOption) error {
	cfg := checkErrorConfig{}
	for _, opt := range auxOptions {
		opt(&cfg)
	}

	for _, check := range cfg.checks {
		if err := check(err); err != nil {
			return err
		}
	}

	return err
}
