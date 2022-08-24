package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/i18n"
)

func bucketS3Cmd() *cobra.Command {
	cmd := &cobra.Command{Use: "s3", Short: i18n.CmdBucketS3}

	cmd.AddCommand(bucketS3CreateCmd(), bucketS3DeleteCmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}

func getAccessKeys(cmd *cobra.Command, serviceName string) (string, string, error) {
	accessKeyId := params.GetString(cmd, "accessKeyId")
	secretAccessKey := params.GetString(cmd, "secretAccessKey")

	// only one of the flags is set
	if (accessKeyId == "" && secretAccessKey != "") || (accessKeyId != "" && secretAccessKey == "") {
		return "", "", errors.New(i18n.BucketS3FlagBothMandatory)
	}

	if accessKeyId == "" && secretAccessKey == "" {
		if val, ok := os.LookupEnv(serviceName + "_accessKeyId"); ok {
			accessKeyId = val
		}
		if val, ok := os.LookupEnv(serviceName + "_secretAccessKey"); ok {
			secretAccessKey = val
		}

		// only one of the env variables was set
		if (accessKeyId == "" && secretAccessKey != "") || (accessKeyId != "" && secretAccessKey == "") {
			return "", "", errors.New(i18n.BucketS3EnvBothMandatory)
		}
	}

	if accessKeyId == "" || secretAccessKey == "" {
		return "", "", errors.New(i18n.BucketS3FlagBothMandatory)
	}

	return accessKeyId, secretAccessKey, nil
}
