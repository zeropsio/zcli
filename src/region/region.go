package region

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/zeropsio/zcli/src/httpClient"
)

type Data struct {
	Name             string `json:"name"`
	IsDefault        bool   `json:"isDefault"`
	RestApiAddress   string `json:"restApiAddress"`
	GrpcApiAddress   string `json:"grpcApiAddress"`
	VpnApiAddress    string `json:"vpnApiAddress"`
	CaCertificateUrl string `json:"caCertificateUrl"`
	S3StorageAddress string `json:"s3StorageAddress"`
}

type Handler struct {
	client *httpClient.Handler
}

func New(client *httpClient.Handler) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) RetrieveAllFromURL(ctx context.Context, regionURL string) ([]Data, error) {
	resp, err := h.client.Get(ctx, regionURL)
	if err != nil {
		return nil, err
	}
	regions, err := readRegions(resp.Body)
	if err != nil {
		return nil, err
	}
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].IsDefault && !regions[j].IsDefault {
			return true
		}
		if regions[j].IsDefault && !regions[i].IsDefault {
			return false
		}
		return regions[i].Name < regions[j].Name
	})
	return regions, nil
}

func readRegions(regionFile json.RawMessage) ([]Data, error) {
	var regions []Data
	err := json.Unmarshal(regionFile, &regions)
	return regions, err
}
