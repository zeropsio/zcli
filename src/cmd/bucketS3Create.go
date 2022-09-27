package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/cliAction/bucket/s3"
	"github.com/zeropsio/zcli/src/i18n"
)

func bucketS3CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create serviceName bucketName [flags]",
		Short:        i18n.CmdBucketCreate,
		Args:         ExactNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			xAmzAcl, err := getXAmzAcl(cmd)
			if err != nil {
				return err
			}

			accessKeyId, secretAccessKey, err := getAccessKeys(cmd, args[0])
			if err != nil {
				return err
			}

			reg, err := getRegion(ctx, cmd)
			if err != nil {
				return err
			}

			bucketName := fmt.Sprintf("%s.%s", strings.ToLower(accessKeyId), args[1])

			fmt.Printf(i18n.BucketCreateCreatingDirect, bucketName)
			fmt.Println(i18n.BucketGenericBucketNamePrefixed)

			b := bucketS3.New(bucketS3.Config{
				S3StorageAddress: reg.S3StorageAddress,
			})
			return b.Create(ctx, bucketS3.RunConfig{
				ServiceStackName: args[0],
				BucketName:       bucketName,
				XAmzAcl:          xAmzAcl,
				AccessKeyId:      accessKeyId,
				SecretAccessKey:  secretAccessKey,
			})
		},
	}
	params.RegisterString(cmd, "x-amz-acl", "", i18n.BucketGenericXAmzAcl)
	params.RegisterString(cmd, "accessKeyId", "", i18n.BucketS3AccessKeyId)
	params.RegisterString(cmd, "secretAccessKey", "", i18n.BucketS3SecretAccessKey)
	params.RegisterString(cmd, "region", "", i18n.BucketS3Region)

	params.RegisterString(cmd, "regionURL", defaultRegionUrl, i18n.RegionUrlFlag)
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.BucketCreateHelp))
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		if err := command.Flags().MarkHidden("regionURL"); err != nil {
			return
		}
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}
