package cmd

import (
	"errors"
	"os"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func bucketS3Cmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("s3").
		Short(i18n.T(i18n.CmdBucketS3)).
		AddChildrenCmd(bucketS3CreateCmd()).
		AddChildrenCmd(bucketS3DeleteCmd())
}

// TODO - janhajek better place?
const (
	s3ServerRegion      = "us-east-1"
	accessKeyIdName     = "accessKeyId"
	secretAccessKeyName = "secretAccessKey"
)

func getAccessKeys(accessKeyId string, secretAccessKey string, serviceName string) (string, string, error) {
	// only one of the flags is set
	if (accessKeyId == "" && secretAccessKey != "") || (accessKeyId != "" && secretAccessKey == "") {
		return "", "", errors.New(i18n.T(i18n.BucketS3FlagBothMandatory))
	}

	if accessKeyId == "" && secretAccessKey == "" {
		if val, ok := os.LookupEnv(serviceName + "_" + accessKeyIdName); ok {
			accessKeyId = val
		}
		if val, ok := os.LookupEnv(serviceName + "_" + secretAccessKeyName); ok {
			secretAccessKey = val
		}

		// only one of the env variables was set
		if (accessKeyId == "" && secretAccessKey != "") || (accessKeyId != "" && secretAccessKey == "") {
			return "", "", errors.New(i18n.T(i18n.BucketS3EnvBothMandatory))
		}
	}

	if accessKeyId == "" || secretAccessKey == "" {
		return "", "", errors.New(i18n.T(i18n.BucketS3FlagBothMandatory))
	}

	return accessKeyId, secretAccessKey, nil
}
