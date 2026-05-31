package errorsx

import (
	"github.com/pkg/errors"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

// Code is the zcli-side mirror of a zerops-go error code. It lets callers outside
// the API boundary match on error codes without importing zerops-go's errorCode package.
type Code string

const (
	CodeProjectNotFound             = Code(errorCode.ProjectNotFound)
	CodeServiceStackNotFound        = Code(errorCode.ServiceStackNotFound)
	CodeServiceStackNameUnavailable = Code(errorCode.ServiceStackNameUnavailable)
	CodeInvalidUserInput            = Code(errorCode.InvalidUserInput)
	CodeZeropsYamlSetupNotFound     = Code(errorCode.ZeropsYamlSetupNotFound)
)

// ApiError is the boundary view of a zerops-go apiError.Error, exposing only what
// command code needs so callers never touch the zerops-go type directly.
type ApiError interface {
	Message() string
	Meta() any
	Code() Code
	HttpStatusCode() int
}

type apiErrorView struct {
	err apiError.Error
}

func (v apiErrorView) Message() string     { return v.err.GetMessage() }
func (v apiErrorView) Meta() any           { return v.err.GetMeta() }
func (v apiErrorView) Code() Code          { return Code(v.err.GetErrorCode()) }
func (v apiErrorView) HttpStatusCode() int { return v.err.GetHttpStatusCode() }

// AsApiError extracts the boundary view of an API error from err, reporting whether
// err (or anything it wraps) is an apiError.Error.
func AsApiError(err error) (ApiError, bool) {
	var apiErr apiError.Error
	if !errors.As(err, &apiErr) {
		return nil, false
	}
	return apiErrorView{err: apiErr}, true
}

// HasErrorCode reports whether err (or anything it wraps) is an API error with the given code.
func HasErrorCode(err error, code Code) bool {
	apiErr, ok := AsApiError(err)
	return ok && apiErr.Code() == code
}
