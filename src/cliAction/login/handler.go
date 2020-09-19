package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/zerops-io/zcli/src/cliStorage"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

type Config struct {
	ApiAddress string
}

type RunConfig struct {
	ZeropsLogin    string
	ZeropsPassword string
}

type Handler struct {
	config     Config
	storage    *cliStorage.Handler
	httpClient *httpClient.Handler
}

func New(
	config Config,
	storage *cliStorage.Handler,
	httpClient *httpClient.Handler,
) *Handler {
	return &Handler{
		config:     config,
		storage:    storage,
		httpClient: httpClient,
	}
}

func (h *Handler) Run(_ context.Context, runConfig RunConfig) error {

	if runConfig.ZeropsLogin == "" {
		return errors.New(i18n.LoginZeropsLoginMissing)
	}
	if runConfig.ZeropsPassword == "" {
		return errors.New(i18n.LoginZeropsPasswordMissing)
	}

	loginData, err := json.Marshal(struct {
		Email    string
		Password string
	}{
		Email:    runConfig.ZeropsLogin,
		Password: runConfig.ZeropsPassword,
	})
	if err != nil {
		return err
	}

	loginResponse, err := h.httpClient.Post(h.config.ApiAddress+"/api/rest/public/auth/login", loginData)
	if err != nil {
		return err
	}

	var loginResponseObject struct {
		Auth struct {
			AccessToken string
		}
	}

	if loginResponse.StatusCode < http.StatusBadRequest {
		err := json.Unmarshal(loginResponse.Body, &loginResponseObject)
		if err != nil {
			return err
		}
	} else {
		return parseRestApiError(loginResponse.Body)
	}

	os := ""
	switch runtime.GOOS {
	case "windows":
		os = "WINDOWS"
	case "darwin":
		os = "MAC"
	case "linux":
		os = "LINUX"
	}

	cliData, err := json.Marshal(struct {
		Os string
	}{
		Os: os,
	})
	if err != nil {
		return err
	}

	cliResponse, err := h.httpClient.Post(
		h.config.ApiAddress+"/api/rest/public/cli/certificate",
		cliData,
		httpClient.BearerAuthorization(loginResponseObject.Auth.AccessToken),
	)
	if err != nil {
		return err
	}

	if cliResponse.StatusCode >= http.StatusBadRequest {
		return parseRestApiError(cliResponse.Body)
	}

	token := string(cliResponse.Body)

	data := h.storage.Data()
	data.Token = token
	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	fmt.Println(i18n.LoginSuccess)

	return nil
}
