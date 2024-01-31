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

func bucketS3DeleteCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("delete").
		Short(i18n.T(i18n.CmdBucketDelete)).
		ScopeLevel(cmdBuilder.Service).
		Arg("bucketName").
		StringFlag(accessKeyIdName, "", i18n.T(i18n.BucketS3AccessKeyId)).
		StringFlag(secretAccessKeyName, "", i18n.T(i18n.BucketS3SecretAccessKey)).
		BoolFlag("confirm", false, i18n.T(i18n.ConfirmFlag)).
		LoggedUserRunFunc(func(ctx context.Context, cmdData *cmdBuilder.LoggedUserCmdData) error {
			uxBlocks := cmdData.UxBlocks

			accessKeyId, secretAccessKey, err := getAccessKeys(
				cmdData.Params.GetString(accessKeyIdName),
				cmdData.Params.GetString(secretAccessKeyName),
				cmdData.Service.Name.String(),
			)
			if err != nil {
				return err
			}

			bucketName := fmt.Sprintf("%s.%s", strings.ToLower(accessKeyId), cmdData.Args["bucketName"][0])

			if !cmdData.Params.GetBool("confirm") {
				err = YesNoPromptDestructive(ctx, cmdData, i18n.T(i18n.BucketDeleteConfirm, bucketName))
				if err != nil {
					return err
				}
			}

			uxBlocks.PrintLine(i18n.T(i18n.BucketDeleteDeletingDirect, bucketName))
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

			bucketInput := (&s3.DeleteBucketInput{}).
				SetBucket(bucketName).
				SetExpectedBucketOwner(accessKeyId)

			if _, err := s3.New(sess).DeleteBucketWithContext(ctx, bucketInput); err != nil {
				var s3Err s3.RequestFailure
				if errors.As(err, &s3Err) {
					return errors.Errorf(i18n.T(i18n.BucketS3RequestFailed), s3Err)
				}
				return err
			}

			uxBlocks.PrintSuccessLine(i18n.T(i18n.BucketDeleted))

			return nil
		})
}
