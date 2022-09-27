package buildDeploy

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/httpClient"
)

func (h *Handler) packageUpload(appVersion *zBusinessZeropsApiProtocol.PostAppVersionResponseDto, reader io.Reader) error {
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
