package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func bucketS3CreateCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("create").
		Short(i18n.T(i18n.CmdBucketCreate)).
		Long(i18n.T(i18n.CmdBucketCreate)).
		ScopeLevel(cmdBuilder.Service).
		Arg("bucketName").
		StringFlag(xAmzAclName, "", i18n.T(i18n.BucketGenericXAmzAcl)).
		StringFlag(accessKeyIdName, "", i18n.T(i18n.BucketS3AccessKeyId)).
		StringFlag(secretAccessKeyName, "", i18n.T(i18n.BucketS3SecretAccessKey)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			xAmzAcl := cmdData.Params.GetString(xAmzAclName)
			err := checkXAmzAcl(xAmzAcl)
			if err != nil {
				return err
			}

			accessKeyId, secretAccessKey, err := getAccessKeys(
				cmdData.Params.GetString(accessKeyIdName),
				cmdData.Params.GetString(secretAccessKeyName),
				cmdData.Service.Name.String(),
			)
			if err != nil {
				return err
			}

			bucketName := fmt.Sprintf("%s.%s", strings.ToLower(accessKeyId), cmdData.Args["bucketName"][0])

			uxBlocks.PrintLine(i18n.T(i18n.BucketCreateCreatingDirect, bucketName))
			uxBlocks.PrintLine(i18n.T(i18n.BucketGenericBucketNamePrefixed))

			awsConf := aws.NewConfig().
				WithEndpoint(cmdData.CliStorage.Data().RegionData.S3StorageAddress).
				WithRegion(s3ServerRegion).
				WithS3ForcePathStyle(true).
				WithCredentials(
					credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
				)

			sess, err := session.NewSession(awsConf)
			if err != nil {
				return err
			}

			bucketInput := (&s3.CreateBucketInput{}).
				SetACL(xAmzAclName).
				SetBucket(bucketName)

			if _, err := s3.New(sess).CreateBucketWithContext(ctx, bucketInput); err != nil {
				var s3Err s3.RequestFailure
				if errors.As(err, &s3Err) {
					if s3Err.Code() == s3.ErrCodeBucketAlreadyExists {
						return errors.New(i18n.T(i18n.BucketS3BucketAlreadyExists))
					}
					return errors.Errorf(i18n.T(i18n.BucketS3RequestFailed), s3Err)
				}
				return err
			}

			uxBlocks.PrintSuccessLine(i18n.T(i18n.BucketCreated))

			return nil
		})
}
