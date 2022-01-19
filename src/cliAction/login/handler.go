package login

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"

	"github.com/zerops-io/zcli/src/grpcDaemonClientFactory"

	"github.com/zerops-io/zcli/src/utils"
	"github.com/zerops-io/zcli/src/zeropsApiProtocol"

	"github.com/zerops-io/zcli/src/grpcApiClientFactory"

	"github.com/zerops-io/zcli/src/cliStorage"
	"github.com/zerops-io/zcli/src/i18n"
	"github.com/zerops-io/zcli/src/utils/httpClient"
)

type Config struct {
	RestApiAddress string
	GrpcApiAddress string
}

type RunConfig struct {
	ZeropsEmail    string
	ZeropsPassword string
	ZeropsToken    string
}

type Handler struct {
	config                    Config
	storage                   *cliStorage.Handler
	httpClient                *httpClient.Handler
	grpcApiClientFactory      *grpcApiClientFactory.Handler
	zeropsDaemonClientFactory *grpcDaemonClientFactory.Handler
}

func New(
	config Config,
	storage *cliStorage.Handler,
	httpClient *httpClient.Handler,
	grpcApiClientFactory *grpcApiClientFactory.Handler,
	zeropsDaemonClientFactory *grpcDaemonClientFactory.Handler,
) *Handler {
	return &Handler{
		config:                    config,
		storage:                   storage,
		httpClient:                httpClient,
		grpcApiClientFactory:      grpcApiClientFactory,
		zeropsDaemonClientFactory: zeropsDaemonClientFactory,
	}
}

func (h *Handler) Run(ctx context.Context, runConfig RunConfig) error {

	if runConfig.ZeropsPassword == "" &&
		runConfig.ZeropsEmail == "" &&
		runConfig.ZeropsToken == "" {
		return errors.New(i18n.LoginParamsMissing)
	}

	var err error
	if runConfig.ZeropsToken != "" {
		err = h.loginWithToken(ctx, runConfig.ZeropsToken)
	} else {
		err = h.loginWithPassword(ctx, runConfig.ZeropsEmail, runConfig.ZeropsPassword)
	}
	if err != nil {
		return err
	}

	daemonClient, closeFunc, err := h.zeropsDaemonClientFactory.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer closeFunc()

	response, err := daemonClient.StopVpn(ctx, &zeropsDaemonProtocol.StopVpnRequest{})
	daemonInstalled, err := utils.HandleDaemonError(err)
	if err != nil {
		return err
	}

	if daemonInstalled && response.GetActiveBefore() {
		fmt.Println(i18n.LoginVpnClosed)
	}

	fmt.Println(i18n.LoginSuccess)
	return nil
}

func (h *Handler) loginWithPassword(_ context.Context, login, password string) error {
	loginData, err := json.Marshal(struct {
		Email    string
		Password string
	}{
		Email:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	loginResponse, err := h.httpClient.Post(h.config.RestApiAddress+"/api/rest/public/auth/login", loginData)
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

	cliResponse, err := h.httpClient.Post(
		h.config.RestApiAddress+"/api/rest/public/user-token",
		nil,
		httpClient.BearerAuthorization(loginResponseObject.Auth.AccessToken),
	)
	if err != nil {
		return err
	}

	if cliResponse.StatusCode >= http.StatusBadRequest {
		return parseRestApiError(cliResponse.Body)
	}

	var tokenResponseObject struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(cliResponse.Body, &tokenResponseObject)
	if err != nil {
		return err
	}

	data := h.storage.Data()
	data.Token = tokenResponseObject.Token
	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) loginWithToken(ctx context.Context, token string) error {

	grpcApiClient, closeFunc, err := h.grpcApiClientFactory.CreateClient(ctx, h.config.GrpcApiAddress, token)
	if err != nil {
		return err
	}
	defer closeFunc()

	resp, err := grpcApiClient.GetUserInfo(ctx, &zeropsApiProtocol.GetUserInfoRequest{})
	if err := utils.HandleGrpcApiError(resp, err); err != nil {
		return err
	}

	data := h.storage.Data()
	data.Token = token
	err = h.storage.Save(data)
	if err != nil {
		return err
	}

	return nil
}
