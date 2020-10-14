package buildDeploy

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"
)

func (h *Handler) packageUpload(appVersion *zeropsApiProtocol.PostAppVersionResponseDto, buff *bytes.Buffer) error {
	fmt.Println(i18n.BuildDeployUploadingPackageStart)

	cephResponse, err := h.httpClient.Put(appVersion.GetUploadUrl(), buff.Bytes(), httpClient.ContentType("application/zip"))
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusCreated {
		return errors.New(i18n.BuildDeployUploadPackageFailed)
	}

	fmt.Println(i18n.BuildDeployUploadingPackageDone)
	return nil
}
