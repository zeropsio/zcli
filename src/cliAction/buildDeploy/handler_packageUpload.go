package buildDeploy

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/zerops-io/zcli/src/proto/business"

	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

func (h *Handler) packageUpload(appVersion *business.PostAppVersionResponseDto, reader io.Reader) error {
	fmt.Println(i18n.BuildDeployUploadingPackageStart)

	cephResponse, err := h.httpClient.PutStream(appVersion.GetUploadUrl(), reader, httpClient.ContentType("application/gzip"))
	if err != nil {
		return err
	}
	if cephResponse.StatusCode != http.StatusCreated {
		return errors.New(i18n.BuildDeployUploadPackageFailed)
	}

	fmt.Println(i18n.BuildDeployUploadingPackageDone)
	return nil
}
