package secrets

import "strings"

// Scope is the first segment of a secret's key path. GlobalScope covers user-level secrets
// (the default token, the cached VPN keys). RepoScope wraps a stable repo identifier (a ULID
// written into the repo's profile.json) for per-repo overrides.
type Scope interface{ prefix() string }

type globalScope struct{}

func (globalScope) prefix() string { return "global" }

// GlobalScope identifies secrets owned by the user globally (no per-directory override).
var GlobalScope Scope = globalScope{}

type repoScope struct{ id string }

func (s repoScope) prefix() string { return "repo/" + s.id }

// RepoScope identifies secrets owned by a specific repo profile, keyed by the repo profile's
// stable Id field. The id is expected to be a ULID so it's filesystem and keyring friendly.
func RepoScope(id string) Scope { return repoScope{id: id} }

// Key composes a secret key from a Scope and one or more name parts. Parts are joined with "/".
// Examples:
//   Key(GlobalScope, "token")              -> "global/token"
//   Key(GlobalScope, "vpn", "<projectId>") -> "global/vpn/<projectId>"
//   Key(RepoScope(id), "token")            -> "repo/<id>/token"
//
// Centralising key construction here means renaming the layout is one place to change, not a
// scan across every caller for string literals.
func Key(scope Scope, parts ...string) string {
	if len(parts) == 0 {
		return scope.prefix()
	}
	return scope.prefix() + "/" + strings.Join(parts, "/")
}
