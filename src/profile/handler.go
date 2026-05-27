// Package profile is the CLI's persisted per-user state: login token, active region, scoped project, and VPN key registry.
// Backed by a single JSON file on disk via the generic storage.Handler. The on-disk schema is versioned so it can evolve without breaking older installs.
//
// # Layout
//
// Each historical schema lives in its own data_vN.go file as a pure struct (dataV0, dataV1, ...). The file registers a decoder for its version via init().
// If the struct is not the current target it also provides an Up() method returning the next version's value.
// The current shape is reached via the dataCurrent alias and exposed to callers as the distinct Data type.
//
// Reads (UnmarshalJSON) probe the on-disk "version" key (missing means v0), decode into the matching dataVN, then walk Up() until reaching dataCurrent.
// Writes (MarshalJSON) always stamp CurrentDataVersion, so the on-disk file is always at the current schema.
//
// # Adding a new migration
//
//  1. Create data_vN.go with the new struct and its init() decoder.
//  2. Add an Up() method on the previous version's struct returning the new dataVN. v0 → v1 was a field rename, but Up() is free to reshape fields, drop them, or compute defaults.
//  3. In handler.go bump CurrentDataVersion to N and flip the dataCurrent alias from dataV(N-1) to dataVN.
//     Data, the type assertion in UnmarshalJSON, and the MarshalJSON cast all follow automatically through the alias.
//
// Finally add a migration test in handler_test.go that loads a representative v(N-1) file and asserts fields land where they should after migration.
package profile

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/storage"
)

const CurrentDataVersion = 1

type Handler = storage.Handler[Data]

func NewHandler(config storage.Config) (*Handler, error) {
	return storage.New[Data](config)
}

// dataCurrent is an alias for the latest dataVN. Bumping a version means changing only this
// line: Data, the terminal type assertion in UnmarshalJSON, and the MarshalJSON cast all resolve
// through this alias and follow automatically.
type dataCurrent = dataV1

// Data is the current on-disk shape, defined as a distinct type over dataCurrent so serialization
// methods live on Data itself without inheriting any methods from the underlying dataVN.
type Data dataCurrent

// upgrader is implemented by every dataVN that is NOT the current target. Up returns the next
// version's value, type-erased because each version is its own Go type.
type upgrader interface {
	Up() any
}

// registry maps on-disk version to a decoder for that version's shape. Populated by init() in
// each data_vN.go file; adding a new version is purely additive.
var registry = map[int]func([]byte) (any, error){}

func register(version int, decode func([]byte) (any, error)) {
	if _, dup := registry[version]; dup {
		panic(errors.Errorf("profile: duplicate registration for version %d", version))
	}
	registry[version] = decode
}

// UnmarshalJSON probes the on-disk "version" key (missing means v0), decodes into the matching
// historical struct, then walks .Up() until the value is the current target type.
func (d *Data) UnmarshalJSON(b []byte) error {
	var probe struct {
		Version *int `json:"version,omitempty"`
	}
	if err := json.Unmarshal(b, &probe); err != nil {
		return err
	}
	v := 0
	if probe.Version != nil {
		v = *probe.Version
	}

	decode, ok := registry[v]
	if !ok {
		return errors.Errorf("profile: no decoder for version %d", v)
	}
	current, err := decode(b)
	if err != nil {
		return err
	}

	for {
		if final, ok := current.(dataCurrent); ok {
			v := CurrentDataVersion
			final.Version = &v
			*d = Data(final)
			return nil
		}
		up, ok := current.(upgrader)
		if !ok {
			return errors.Errorf("profile: %T is not the current version and has no Up()", current)
		}
		current = up.Up()
	}
}

// MarshalJSON forces CurrentDataVersion on every write, regardless of the in-memory value.
// Pairs with UnmarshalJSON so the on-disk format is always at the current schema and every
// write produces a "version" key (omitempty on Version applies to other marshal contexts,
// but on the storage path this always stamps).
// Value receiver (asymmetric with UnmarshalJSON) is required: storage.Handler.save passes the value through json.Encode, and a non-addressable value only finds methods in its value method set.
func (d Data) MarshalJSON() ([]byte, error) {
	out := dataCurrent(d)
	v := CurrentDataVersion
	out.Version = &v
	return json.Marshal(out)
}
