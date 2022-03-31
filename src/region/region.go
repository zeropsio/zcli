package region

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/zerops-io/zcli/src/i18n"

	"github.com/zerops-io/zcli/src/constants"

	"github.com/zerops-io/zcli/src/utils/httpClient"
)

func RetrieveFromURL(client *httpClient.Handler, regionURL, region string) (Config, error) {
	resp, err := client.Get(regionURL)
	if err != nil {
		return Config{}, err
	}
	reg, err := readRegion(region, resp.Body)
	if err != nil {
		return Config{}, err
	}
	regJson, err := json.Marshal(reg)
	if err != nil {
		return Config{}, err
	}
	filepath, err := constants.CliRegionData()
	if err != nil {
		return Config{}, err
	}
	err = os.WriteFile(filepath, regJson, 0666)
	return reg, err
}

func RetrieveFromFile() (Config, error) {
	filepath, err := constants.CliRegionData()
	if err != nil {
		return Config{}, err
	}
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}
	var reg Config
	err = json.Unmarshal(f, &reg)
	if err != nil {
		return Config{}, err
	}
	return reg, nil
}

func readRegion(region string, regionFile json.RawMessage) (Config, error) {
	var regions []Config

	err := json.Unmarshal(regionFile, &regions)
	if err != nil {
		return Config{}, err
	}

	var reg *Config
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
		return Config{}, errors.New(i18n.RegionNotFound)
	}

	return *reg, nil
}
