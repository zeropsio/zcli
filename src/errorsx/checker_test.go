package errorsx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

func TestCascading(t *testing.T) {
	inputError := apiError.Error{
		HttpStatusCode: 400,
		ErrorCode:      string(errorCode.InvalidUserInput),
		Message:        "apiErrorMessage",
	}

	tests := []struct {
		name           string
		inputErr       error
		check          Check
		convertedError error
	}{
		{
			name:           "No checks",
			inputErr:       inputError,
			check:          nil,
			convertedError: inputError,
		},
		{
			name:     "or",
			inputErr: inputError,
			check: Or(
				ErrorCode(CodeProjectNotFound),
				ErrorCode(CodeInvalidUserInput),
				ErrorCode(CodeServiceStackNotFound),
			),
			convertedError: NewUserError(inputError.Message, inputError),
		},
		{
			name:     "and",
			inputErr: inputError,
			check: And(
				HttpStatusCode(400),
				ErrorCode(CodeInvalidUserInput),
			),
			convertedError: NewUserError(inputError.Message, inputError),
		},
		{
			name:     "complex",
			inputErr: inputError,
			check: Or(
				And(ErrorCode(CodeServiceStackNameUnavailable)),
				And(HttpStatusCode(400), ErrorCode(CodeInvalidUserInput)),
				And(ErrorCode(CodeServiceStackNotFound)),
			),
			convertedError: NewUserError(inputError.Message, inputError),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			response := Convert(tt.inputErr, tt.check)
			require.Equal(t, tt.convertedError, response)
		})
	}
}
