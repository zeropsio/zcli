package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/cliAction/bucket/s3"
	"github.com/zerops-io/zcli/src/i18n"
)

func bucketS3DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete serviceName bucketName [flags]",
		Short:        i18n.CmdBucketDelete,
		Args:         ExactNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			accessKeyId, secretAccessKey, err := getAccessKeys(cmd, args[0])
			if err != nil {
				return err
			}

			region, err := createRegionRetriever(ctx)
			if err != nil {
				return err
			}

			reg, err := region.RetrieveFromFile()
			if err != nil {
				return err
			}

			bucketName := fmt.Sprintf("%s.%s", strings.ToLower(accessKeyId), args[1])

			fmt.Printf(i18n.BucketDeleteDeletingDirect, bucketName)
			fmt.Println(i18n.BucketGenericBucketNamePrefixed)

			b := bucketS3.New(bucketS3.Config{
				S3StorageAddress: reg.S3StorageAddress,
			})
			return b.Delete(ctx, bucketS3.RunConfig{
				ServiceStackName: args[0],
				BucketName:       bucketName,
				AccessKeyId:      accessKeyId,
				SecretAccessKey:  secretAccessKey,
			})
		},
	}
	params.RegisterString(cmd, "x-amz-acl", "", i18n.BucketGenericXAmzAcl)
	params.RegisterString(cmd, "accessKeyId", "", i18n.BucketS3AccessKeyId)
	params.RegisterString(cmd, "secretAccessKey", "", i18n.BucketS3SecretAccessKey)

	cmd.Flags().BoolP("help", "h", false, helpText(i18n.BucketDeleteHelp))
	return cmd
}
