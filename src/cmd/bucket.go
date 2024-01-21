package cmd

import (
	"errors"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/i18n"
)

func bucketCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("bucket").
		Short(i18n.T(i18n.CmdBucket)).
		AddChildrenCmd(bucketZeropsCmd()).
		AddChildrenCmd(bucketS3Cmd())
}

// FIXME - janhajek better place?
const (
	xAmzAclName = "x-amz-acl"
)

func checkXAmzAcl(xAmzAcl string) error {
	switch xAmzAcl {
	case "", "private", "public-read", "public-read-write", "authenticated-read":
		return nil
	}

	return errors.New(i18n.T(i18n.BucketGenericXAmzAclInvalid))
}
