package errorsx

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

type convertor func(err error) error
type check func(err error, errMessage string) error

func Check(err error, checks ...check) bool {
	if err == nil {
		return false
	}

	for _, check := range checks {
		if err := check(err, ""); err != nil {
			return true
		}
	}

	return false
}

func Convert(err error, convertors ...convertor) error {
	if err == nil {
		return nil
	}

	for _, convertor := range convertors {
		if err := convertor(err); err != nil {
			return err
		}
	}

	return err
}

func CheckErrorCode(errorCode errorCode.ErrorCode) check {
	return func(err error, errMessage string) error {
		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}
		if string(errorCode) != apiErr.GetErrorCode() {
			return nil
		}

		return NewUserError(errMessage, err)
	}
}

func CheckInvalidUserInput(parameterName string) check {
	return func(err error, errMessage string) error {
		var apiErr apiError.Error
		if !errors.As(err, &apiErr) {
			return nil
		}

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
				}
			}
		}

		return nil
	}
}

func ConvertErrorCode(errorCode errorCode.ErrorCode, errMessage string) convertor {
	return func(err error) error {
		return CheckErrorCode(errorCode)(err, errMessage)
	}
}

func ConvertInvalidUserInput(parameterName string, errMessage string) convertor {
	return func(err error) error {
		return CheckInvalidUserInput(parameterName)(err, errMessage)
	}
}
