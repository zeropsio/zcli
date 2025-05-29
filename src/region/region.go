package region

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/httpClient"
	"github.com/zeropsio/zcli/src/optional"
)

type Item struct {
	Name       string                `json:"name"`
	IsDefault  bool                  `json:"isDefault"`
	Address    string                `json:"address"`
	GuiAddress optional.Null[string] `json:"guiAddress"`
}

type Handler struct {
	client *httpClient.Handler
}

func New(client *httpClient.Handler) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) RetrieveAllFromURL(ctx context.Context, regionURL string) ([]Item, error) {
	resp, err := h.client.Get(ctx, regionURL)
	if err != nil {
		return nil, errors.Wrap(err, "retrieve regions")
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

func readRegions(regionFile json.RawMessage) ([]Item, error) {
	var regionItemsResponse response
	err := json.Unmarshal(regionFile, &regionItemsResponse)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal regions")
	}
	return regionItemsResponse.Items, err
}

type response struct {
	Items []Item `json:"items"`
}
