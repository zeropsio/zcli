package userInfo

import (
	"context"
	"fmt"

	"github.com/zerops-io/zcli/src/helpers"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func Run(ctx context.Context, apiGrpcClient zeropsApiProtocol.ZeropsApiProtocolClient) error {

	userInfoResponse, err := apiGrpcClient.GetUserInfo(ctx, &zeropsApiProtocol.GetUserInfoRequest{})
	if err := helpers.HandleGrpcApiError(userInfoResponse, err); err != nil {
		return err
	}

	fmt.Println(userInfoResponse.GetOutput())

	return nil
}
