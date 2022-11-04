package prolongVpn

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/zeropsio/zcli/src/daemonStorage"
	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/vpnproxy"
	"github.com/zeropsio/zcli/src/proto/zBusinessZeropsApiProtocol"
	"github.com/zeropsio/zcli/src/utils/logger"
	"golang.zx2c4.com/wireguard/wgctrl"
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
	if data.InterfaceName == "" {
		return nil
	}
	if data.Expiry.Sub(time.Now()) > thresholdInterval {
		return nil
	}
	apiClientFactory := zBusinessZeropsApiProtocol.New(zBusinessZeropsApiProtocol.Config{CaCertificateUrl: data.CaCertificateUrl})
	apiGrpcClient, closeFunc, err := apiClientFactory.CreateClient(ctx, data.GrpcApiAddress, data.Token)
	if err != nil {
		return err
	}
	defer closeFunc()

	wgClient, err := wgctrl.New()
	if err != nil {
		return errors.New(i18n.VpnStatusWireguardNotAvailable)
	}
	defer wgClient.Close()

	device, err := wgClient.Device(data.InterfaceName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	businessResp, err := apiGrpcClient.PostVpnRequest(ctx, &zBusinessZeropsApiProtocol.PostVpnRequestRequest{
		Id:              data.ProjectId,
		ClientPublicKey: device.PublicKey.String(),
	})
	if err := proto.BusinessError(businessResp, err); err != nil {
		return err
	}
	expiry := businessResp.GetOutput().GetExpiry()
	accessToken := businessResp.GetOutput().GetAccessToken()

	vpnClient, closeFn, err := vpnproxy.CreateClient(ctx, data.GrpcTargetVpnAddress)
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

	h.storage.Update(func(data daemonStorage.Data) daemonStorage.Data {
		data.Expiry = zBusinessZeropsApiProtocol.FromProtoTimestamp(expiry)
		return data
	})

	return nil
}
