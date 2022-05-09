package prolongVpn

import (
	"context"
	"time"

	"github.com/zerops-io/zcli/src/daemonStorage"
	"github.com/zerops-io/zcli/src/proto"
	"github.com/zerops-io/zcli/src/proto/business"
	"github.com/zerops-io/zcli/src/proto/vpnproxy"
	"github.com/zerops-io/zcli/src/utils/logger"
)

const (
	cronInterval      = time.Minute
	thresholdInterval = 10 * time.Minute
)

type Handler struct {
	storage *daemonStorage.Handler
	log     *logger.Handler
}

func New(storage *daemonStorage.Handler, log *logger.Handler) *Handler {
	return &Handler{
		storage: storage,
		log:     log,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	t := time.NewTicker(cronInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			h.log.Debug("checking prolong")
			err := h.prolong(ctx)
			if err != nil {
				h.log.Warning("prolong: ", err)
			}
		}
	}
}

func (h *Handler) prolong(ctx context.Context) error {
	data := h.storage.Data()
	if !data.VpnStarted {
		return nil
	}
	if data.Expiry.Sub(time.Now()) > thresholdInterval {
		h.log.Debug("prolong threshold not met")
		return nil
	}
	apiClientFactory := business.New(business.Config{CaCertificateUrl: data.CaCertificateUrl})
	apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(ctx, data.GrpcApiAddress, data.Token)
	if err != nil {
		return err
	}
	defer closeFunc()

	businessResp, err := apiGrpcClient.PostVpnRequest(ctx, &business.PostVpnRequestRequest{
		Id:              data.ProjectId,
		ClientPublicKey: data.PublicKey,
	})
	if err := proto.BusinessError(businessResp, err); err != nil {
		return err
	}
	expiry := businessResp.GetOutput().GetExpiry()
	accessToken := businessResp.GetOutput().GetAccessToken()

	vpnClient, closeFn, err := vpnproxy.CreateClient(ctx, data.GrpcVpnAddress)
	if err != nil {
		return err
	}
	vpnResp, err := vpnClient.ProlongVpn(ctx, &vpnproxy.ProlongVpnRequest{
		AccessToken: accessToken,
	})
	closeFn()
	if err := proto.VpnError(vpnResp, err); err != nil {
		return err
	}

	data.Expiry = business.FromProtoTimestamp(expiry)

	return h.storage.Save(data)
}
