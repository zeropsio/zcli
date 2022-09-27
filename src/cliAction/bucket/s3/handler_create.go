package bucketS3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/i18n"
)

func (h Handler) Create(ctx context.Context, config RunConfig) error {
	awsConf := aws.NewConfig().
		WithEndpoint(h.config.S3StorageAddress).
		WithRegion(s3ServerRegion).
		WithS3ForcePathStyle(true).
		WithCredentials(
			credentials.NewStaticCredentials(config.AccessKeyId, config.SecretAccessKey, ""),
		)

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return err
	}

	bucketInput := (&s3.CreateBucketInput{}).
		SetACL(config.XAmzAcl).
		SetBucket(config.BucketName)

	if _, err := s3.New(sess).CreateBucketWithContext(ctx, bucketInput); err != nil {
		var s3Err s3.RequestFailure
		if errors.As(err, &s3Err) {
			if s3Err.Code() == s3.ErrCodeBucketAlreadyExists {
				return errors.New(i18n.BucketS3BucketAlreadyExists)
			}
			return fmt.Errorf(i18n.BucketS3RequestFailed, s3Err)
		}
		return err
	}

	fmt.Println(constants.Success + i18n.BucketCreated + i18n.Success)

	return nil
}
