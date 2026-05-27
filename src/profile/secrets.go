package profile

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/entity"
	"github.com/zeropsio/zcli/src/secrets"
	zuuid "github.com/zeropsio/zerops-go/types/uuid"
)

// ValueSource indicates which profile layer produced an effective value.
type ValueSource string

const (
	// SourceNone value is not set in any layer.
	SourceNone ValueSource = ""
	// SourceGlobal value comes from the global profile / global-scope secrets.
	SourceGlobal ValueSource = "global"
	// SourceRepo value comes from the repo profile / repo-scope secrets (overrides global).
	SourceRepo ValueSource = "repo"
)

// Token returns the effective token. Repo-scoped token (when set) wins over the global one.
// Returns ("", false, nil) when neither layer has a token.
func (m *Manager) Token() (string, bool, error) {
	v, src, err := m.TokenWithSource()
	return v, src != SourceNone, err
}

// TokenWithSource returns the effective token along with which layer produced it. Used by
// introspection surfaces that want to attribute the value (`zcli profile where`); regular
// callers should use Token, which discards the source.
func (m *Manager) TokenWithSource() (string, ValueSource, error) {
	m.mu.RLock()
	repoId := ""
	if p, ok := m.activeProfile(); ok {
		repoId = p.Id
	}
	m.mu.RUnlock()

	if repoId != "" {
		v, ok, err := m.secrets.Get(secrets.Key(secrets.RepoScope(repoId), "token"))
		if err != nil {
			return "", SourceNone, err
		}
		if ok {
			return v, SourceRepo, nil
		}
	}
	v, ok, err := m.secrets.Get(secrets.Key(secrets.GlobalScope, "token"))
	if err != nil {
		return "", SourceNone, err
	}
	if ok {
		return v, SourceGlobal, nil
	}
	return "", SourceNone, nil
}

// SetToken writes the auth token to the global secrets scope. Repo-scoped tokens are still
// read by TokenWithSource if present in the backend, but writing one is not exposed today -
// add a dedicated method (e.g. SetRepoToken) when a caller actually needs it.
func (m *Manager) SetToken(token string) error {
	return m.secrets.Set(secrets.Key(secrets.GlobalScope, "token"), token)
}

// Clear wipes the global profile file plus every secret under the global scope (token, all
// VPN keys). Repo profiles are left alone - they belong to the repo, not the user session.
func (m *Manager) Clear() (Data, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prior := m.global
	m.global = Data{}
	if err := m.saveGlobal(m.global); err != nil {
		return prior, err
	}
	keys, err := m.secrets.List(secrets.Prefix(secrets.GlobalScope))
	if err != nil {
		return prior, errors.Wrap(err, "list global secrets")
	}
	for _, k := range keys {
		if err := m.secrets.Delete(k); err != nil {
			return prior, errors.Wrapf(err, "delete secret %s", k)
		}
	}
	return prior, nil
}

// VpnKey returns the cached VPN key for a project. VPN keys are global only - a project's key
// is the same regardless of which repo asks for it.
func (m *Manager) VpnKey(projectId zuuid.ProjectId) (entity.VpnKey, bool, error) {
	raw, ok, err := m.secrets.Get(vpnKeyPath(projectId))
	if err != nil || !ok {
		return entity.VpnKey{}, ok, err
	}
	var k entity.VpnKey
	if err := json.Unmarshal([]byte(raw), &k); err != nil {
		return entity.VpnKey{}, false, errors.Wrapf(err, "decode cached vpn key for project %s", projectId.Native())
	}
	return k, true, nil
}

// SetVpnKey writes (or overwrites) the cached VPN key for a project.
func (m *Manager) SetVpnKey(projectId zuuid.ProjectId, key entity.VpnKey) error {
	payload, err := json.Marshal(key)
	if err != nil {
		return errors.WithStack(err)
	}
	return m.secrets.Set(vpnKeyPath(projectId), string(payload))
}

// DeleteVpnKey removes the cached VPN key for a project. Idempotent.
func (m *Manager) DeleteVpnKey(projectId zuuid.ProjectId) error {
	return m.secrets.Delete(vpnKeyPath(projectId))
}

func vpnKeyPath(projectId zuuid.ProjectId) string {
	return secrets.Key(secrets.GlobalScope, "vpn", projectId.Native())
}
