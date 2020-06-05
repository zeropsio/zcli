package login

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime"

	"github.com/zerops-io/zcli/src/service/httpClient"

	"github.com/zerops-io/zcli/src/service/logger"
	"github.com/zerops-io/zcli/src/service/storage"
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
	logger     *logger.Handler
	storage    *storage.Handler
	httpClient *httpClient.Handler
}

func New(
	config Config,
	logger *logger.Handler,
	storage *storage.Handler,
	httpClient *httpClient.Handler,
) *Handler {
	return &Handler{
		config:     config,
		logger:     logger,
		storage:    storage,
		httpClient: httpClient,
	}
}

func (h *Handler) Run(_ context.Context, runConfig RunConfig) error {
	if runConfig.ZeropsLogin == "" {
		return errors.New("param zeropsLogin must be set")
	}
	if runConfig.ZeropsPassword == "" {
		return errors.New("param ZeropsPassword must be set")
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

	if loginResponse.StatusCode < http.StatusMultipleChoices {
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

	if cliResponse.StatusCode >= http.StatusMultipleChoices {
		return parseRestApiError(cliResponse.Body)
	}

	token := string(cliResponse.Body)

	h.storage.Data.Token = token
	err = h.storage.Save()
	if err != nil {
		return err
	}

	h.logger.Info("you are logged")

	return nil
}
