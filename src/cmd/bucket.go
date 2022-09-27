package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/i18n"
)

func bucketCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "bucket", Short: i18n.CmdBucket}

	cmd.AddCommand(bucketZeropsCmd(), bucketS3Cmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}

func getXAmzAcl(cmd *cobra.Command) (string, error) {
	xAmzAcl := params.GetString(cmd, "x-amz-acl")
	switch xAmzAcl {
	case "", "private", "public-read", "public-read-write", "authenticated-read":
		return xAmzAcl, nil
	}

	return "", errors.New(i18n.BucketGenericXAmzAclInvalid)
}
