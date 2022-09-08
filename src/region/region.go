package region

import (
	"encoding/json"
	"errors"
	"sort"

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
	S3StorageAddress string `json:"s3StorageAddress"`
}

type Handler struct {
	client  *httpClient.Handler
	storage *storage.Handler[Data]
}

func New(client *httpClient.Handler, storage *storage.Handler[Data]) *Handler {
	return &Handler{storage: storage, client: client}
}

// RetrieveFromURL retrieves the region from URL, if region is empty, returns a default region
func (h *Handler) RetrieveFromURL(regionURL, region string) (Data, error) {
	resp, err := h.client.Get(regionURL)
	if err != nil {
		return Data{}, err
	}
	reg, err := readRegion(region, resp.Body)
	if err != nil {
		return Data{}, err
	}
	return reg, nil
}

// RetrieveFromURLAndSave retrieves the region using RetrieveFromURL and stores it into the file
func (h *Handler) RetrieveFromURLAndSave(regionURL, region string) (Data, error) {
	reg, err := h.RetrieveFromURL(regionURL, region)
	if err != nil {
		return Data{}, err
	}
	return reg, h.storage.Save(&reg)
}

func (h *Handler) RetrieveAllFromURL(regionURL string) ([]Data, error) {
	resp, err := h.client.Get(regionURL)
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

func (h *Handler) RetrieveFromFile() (Data, error) {
	return *h.storage.Data(), nil
}

func readRegions(regionFile json.RawMessage) ([]Data, error) {
	var regions []Data
	err := json.Unmarshal(regionFile, &regions)
	return regions, err
}

func readRegion(region string, regionFile json.RawMessage) (Data, error) {
	regions, err := readRegions(regionFile)
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
