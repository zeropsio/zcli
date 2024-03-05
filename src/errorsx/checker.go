package errorsx

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

type check func(err error) error

func CheckErrorCode(errorCode errorCode.ErrorCode, errMessage string) check {
	return func(err error) error {
		var apiErr apiError.Error
		if errors.As(err, &apiErr) {
			if string(errorCode) != apiErr.GetErrorCode() {
				return nil
			}

			return NewUserError(errMessage, err)
		}

		return nil
	}
}

func CheckInvalidUserInput(parameterName string, errMessage string) check {
	return func(err error) error {
		var apiErr apiError.Error
		if errors.As(err, &apiErr) {
			if string(errorCode.InvalidUserInput) != apiErr.GetErrorCode() {
				return nil
			}

			meta, ok := apiErr.GetMeta().([]interface{})
			if !ok {
				return nil
			}

			for _, metaItem := range meta {
				if metaItemTyped, ok := metaItem.(map[string]interface{}); ok {
					if parameterValue, ok := metaItemTyped["parameter"]; ok {
						if parameterValue == parameterName {
							return NewUserError(fmt.Sprintf(errMessage, metaItemTyped["message"]), err)
						}
						return nil
					}
				}
			}
		}

		return nil
	}
}

func Check(err error, checks ...check) error {
	if err == nil {
		return nil
	}

	for _, check := range checks {
		if err := check(err); err != nil {
			return err
		}
	}

	return err
}
