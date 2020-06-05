package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	FilePath string
}

type Handler struct {
	config Config

	Data *Data
}

type Data struct {
	ProjectId string
	ServerIp  string
	Token     string
}

func New(config Config) (*Handler, error) {

	h := &Handler{
		config: config,
		Data:   &Data{},
	}

	if fileExists(config.FilePath) {
		f, err := os.Open(config.FilePath)
		if err != nil {
			return nil, err
		}

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(bytes, &h.Data)
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (h *Handler) Save() error {
	data, err := json.Marshal(h.Data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(h.config.FilePath, data, 0644)
	if err != nil {
		return err
	}
	return err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
