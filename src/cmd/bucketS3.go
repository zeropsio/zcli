package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/region"
)

func bucketS3Cmd() *cobra.Command {
	cmd := &cobra.Command{Use: "s3", Short: i18n.CmdBucketS3}

	cmd.AddCommand(bucketS3CreateCmd(), bucketS3DeleteCmd())
	cmd.Flags().BoolP("help", "h", false, helpText(i18n.GroupHelp))
	return cmd
}

func getRegion(ctx context.Context, cmd *cobra.Command) (region.Data, error) {
	retriever, err := createRegionRetriever(ctx)
	if err != nil {
		return region.Data{}, err
	}

	// prefer region from command parameter or env
	selectedRegion := params.GetString(cmd, "region")

	// if not provided, try to use the region user is logged to
	if selectedRegion == "" {
		reg, err := retriever.RetrieveFromFile()
		if err != nil {
			return region.Data{}, err
		}
		if reg.Name != "" {
			return reg, nil
		}
	}

	// if no region is found, get a default region
	regionURL := params.GetString(cmd, "regionURL")
	return retriever.RetrieveFromURL(regionURL, selectedRegion)
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
