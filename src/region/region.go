package region

import (
	"encoding/json"
	"errors"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
	"github.com/zerops-io/zcli/src/utils/storage"
)

type Data struct {
	Name             string `json:"name"`
	IsDefault        bool   `json:"isDefault"`
	RestApiAddress   string `json:"restApiAddress"`
	GrpcApiAddress   string `json:"grpcApiAddress"`
	VpnApiAddress    string `json:"vpnApiAddress"`
	CaCertificateUrl string `json:"caCertificateUrl"`
}

type Handler struct {
	client *httpClient.Handler
	storage *storage.Handler[Data]
}

func New(client *httpClient.Handler, storage *storage.Handler[Data]) *Handler {
	return &Handler{storage: storage, client: client}
}

func (h *Handler) RetrieveFromURL(regionURL, region string) (Data, error) {
	resp, err := h.client.Get(regionURL)
	if err != nil {
		return Data{}, err
	}
	reg, err := readRegion(region, resp.Body)
	if err != nil {
		return Data{}, err
	}
	return reg, h.storage.Save(&reg)
}

func (h *Handler) RetrieveFromFile() (Data, error) {
	return *h.storage.Data(), nil
}

func readRegion(region string, regionFile json.RawMessage) (Data, error) {
	var regions []Data

	err := json.Unmarshal(regionFile, &regions)
	if err != nil {
		return Data{}, err
	}

	var reg *Data
	for _, r := range regions {
		r := r
		if r.IsDefault && region == "" {
			reg = &r
		}
		if r.Name == region {
			reg = &r
		}
	}

	if reg == nil {
		return Data{}, errors.New(i18n.RegionNotFound)
	}

	return *reg, nil
}
