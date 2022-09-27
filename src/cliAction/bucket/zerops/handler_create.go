package bucketZerops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/projectService"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

func (h Handler) Create(ctx context.Context, config RunConfig) error {
	projectId, err := projectService.GetProjectId(ctx, h.apiGrpcClient, config.ProjectNameOrId, h.sdkConfig)
	if err != nil {
		return err
	}

	stack, err := projectService.GetServiceStack(ctx, h.apiGrpcClient, projectId, config.ServiceStackName)
	if err != nil {
		return err
	}
	if stack.GetServiceStackTypeInfo().GetServiceStackTypeCategory() != zBusinessZeropsApiProtocol.ServiceStackTypeCategory_SERVICE_STACK_TYPE_CATEGORY_OBJECT_STORAGE {
		return errors.New(i18n.BucketGenericOnlyForObjectStorage)
	}

	stackId := stack.GetId()
	bucketName := fmt.Sprintf("%s.%s", strings.ToLower(stackId), config.BucketName)

	fmt.Printf(i18n.BucketCreateCreatingZeropsApi, bucketName)
	fmt.Println(i18n.BucketGenericBucketNamePrefixed)

	zdk := sdk.New(
		sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(h.sdkConfig.RegionUrl)),
		&http.Client{Timeout: 1 * time.Minute},
	)
	authorizedSdk := sdk.AuthorizeSdk(zdk, h.sdkConfig.Token)

	bucketBody := body.PostS3Bucket{
		Name: types.NewString(bucketName),
	}
	if config.XAmzAcl != "" {
		bucketBody.XAmzAcl = types.NewStringNull(config.XAmzAcl)
	}

	resp, err := authorizedSdk.PostS3Bucket(
		ctx,
		path.ServiceStackIdNamed{ServiceStackId: uuid.ServiceStackId(stackId)},
		bucketBody,
	)
	if err != nil {
		return err
	}
	if _, err := resp.Output(); err != nil {
		return err
	}

	fmt.Println(constants.Success + i18n.BucketCreated + i18n.Success)

	return nil
}
