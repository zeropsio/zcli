package repository

import (
	"context"
	"encoding/base64"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/gn"
	"github.com/zeropsio/zcli/src/zeropsRestApiClient"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// RegisterProjectVpn registers publicKey with the project's VPN and returns the
// settings needed to build a WireGuard config. The returned type is consumed by
// src/wg (also an API-boundary package), so it stays a zerops-go output type.
func RegisterProjectVpn(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	publicKey string,
) (output.ProjectVpnItem, error) {
	response, err := restApiClient.PostProjectVpn(
		ctx,
		path.ProjectId{Id: uuid.ProjectId(projectId)},
		body.PostProjectVpn{PublicKey: types.NewString(publicKey)},
	)
	if err != nil {
		return output.ProjectVpnItem{}, err
	}
	return response.Output()
}

// DeleteProjectVpnKey removes a registered VPN public key from the project.
func DeleteProjectVpnKey(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	publicKey string,
) error {
	response, err := restApiClient.DeleteProjectVpn(
		ctx,
		path.ProjectId{Id: uuid.ProjectId(projectId)},
		body.PostProjectVpn{PublicKey: types.NewString(publicKey)},
	)
	if err != nil {
		return err
	}
	return response.Err()
}

// ListProjectVpnPublicKeys returns the public keys of all VPN peers registered with the project.
func ListProjectVpnPublicKeys(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
) ([]string, error) {
	response, err := restApiClient.GetProjectVpnList(
		ctx,
		path.ProjectId{Id: uuid.ProjectId(projectId)},
		query.GetProjectVpn{},
	)
	if err != nil {
		return nil, err
	}
	list, err := response.Output()
	if err != nil {
		return nil, err
	}
	return gn.TransformSlice(list.Peers, func(peer output.ProjectVpnPeerSetup) string {
		return peer.PublicKey.Native()
	}), nil
}

// ProjectVpnKeyExists reports whether publicKey is registered with the project's VPN.
func ProjectVpnKeyExists(
	ctx context.Context,
	restApiClient *zeropsRestApiClient.Handler,
	projectId entity.ProjectId,
	publicKey string,
) (bool, error) {
	response, err := restApiClient.GetProjectVpn(
		ctx,
		path.ProjectIdBase64PublicKey{
			Id:              uuid.ProjectId(projectId),
			Base64PublicKey: types.NewString(base64.URLEncoding.EncodeToString([]byte(publicKey))),
		},
		query.GetProjectVpn{},
	)
	if err != nil {
		return false, err
	}
	return response.Err() == nil, nil
}
