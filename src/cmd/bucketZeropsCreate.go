package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
)

func bucketZeropsCreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short(i18n.T(i18n.CmdBucketCreate)).
		ScopeLevel(cmdBuilder.Service).
		Arg("bucketName").
		StringFlag(xAmzAclName, "", i18n.T(i18n.BucketGenericXAmzAcl)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			xAmzAcl := cmdData.Params.GetString(xAmzAclName)
			err := checkXAmzAcl(xAmzAcl)
			if err != nil {
				return err
			}

			if cmdData.Service.ServiceTypeCategory != enum.ServiceStackTypeCategoryEnumObjectStorage {
				return errors.New(i18n.T(i18n.BucketGenericOnlyForObjectStorage))
			}

			serviceId := cmdData.Service.ID
			bucketName := fmt.Sprintf("%s.%s", strings.ToLower(serviceId.Native()), cmdData.Args["bucketName"][0])

			uxBlocks.PrintLine(i18n.T(i18n.BucketCreateCreatingZeropsApi, bucketName))
			uxBlocks.PrintLine(i18n.T(i18n.BucketGenericBucketNamePrefixed))

			bucketBody := body.PostS3Bucket{
				Name: types.NewString(bucketName),
			}
			if xAmzAcl != "" {
				bucketBody.XAmzAcl = types.NewStringNull(xAmzAcl)
			}

			resp, err := cmdData.RestApiClient.PostS3Bucket(
				ctx,
				path.ServiceStackIdNamed{ServiceStackId: serviceId},
				bucketBody,
			)
			if err != nil {
				return err
			}
			if _, err := resp.Output(); err != nil {
				return err
			}

			uxBlocks.PrintSuccessLine(i18n.T(i18n.BucketCreated))

			return nil
		})
}
