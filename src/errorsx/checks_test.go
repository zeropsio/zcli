package errorsx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/errorCode"
)

func TestCheckInvalidUserInput(t *testing.T) {
	inputErr := apiError.Error{
		ErrorCode: string(errorCode.InvalidUserInput),
		Message:   "invalid user input",
		Meta: []interface{}{
			map[string]interface{}{"parameter": "name", "message": "name message"},
			map[string]interface{}{"parameter": "id", "message": "id message"},
		},
	}

	{
		check := Or(InvalidUserInput("id"))

		expectedError := NewUserError("invalid user input", inputErr)

		require.True(t, Is(inputErr, check))
		require.Equal(t, expectedError, Convert(inputErr, check))
	}
	{
		check := Or(InvalidUserInput(
			"id",
			InvalidUserInputErrorMessage(func(apiErr apiError.Error, metaItem map[string]interface{}) string {
				return fmt.Sprintf("%s:%s", apiErr.GetMessage(), metaItem["message"])
			}),
		))

		expectedError := NewUserError("invalid user input:id message", inputErr)

		require.True(t, Is(inputErr, check))
		require.Equal(t, expectedError, Convert(inputErr, check))
	}
	{
		check := Or(InvalidUserInput("age"))

		require.False(t, Is(inputErr, check))
		require.Equal(t, inputErr, Convert(inputErr, check))
	}
}

func TestErrorCode(t *testing.T) {
	inputErr := apiError.Error{
		ErrorCode: string(errorCode.InvalidUserInput),
		Message:   "invalid user input",
	}

	{
		check := Or(ErrorCode(errorCode.InvalidUserInput))

		expectedError := NewUserError("invalid user input", inputErr)

		require.True(t, Is(inputErr, check))
		require.Equal(t, expectedError, Convert(inputErr, check))
	}
	{
		check := Or(ErrorCode(errorCode.PaymentFailedError))

		require.False(t, Is(inputErr, check))
		require.Equal(t, inputErr, Convert(inputErr, check))
	}
}

func TestHttpStatusCode(t *testing.T) {
	inputErr := apiError.Error{
		HttpStatusCode: 400,
		ErrorCode:      string(errorCode.InvalidUserInput),
		Message:        "invalid user input",
	}

	{
		check := Or(HttpStatusCode(400))

		expectedError := NewUserError("invalid user input", inputErr)

		require.True(t, Is(inputErr, check))
		require.Equal(t, expectedError, Convert(inputErr, check))
	}
	{
		check := Or(HttpStatusCode(500))

		require.False(t, Is(inputErr, check))
		require.Equal(t, inputErr, Convert(inputErr, check))
	}
}
