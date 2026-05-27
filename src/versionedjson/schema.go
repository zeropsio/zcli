// Package versionedjson loads a JSON document through a chain of schema versions, running each
// older version's Up step until it reaches the current target shape. The migration context
// type is a generic parameter on Schema so callers thread whatever per-step state they need
// (secrets backend, cwd, callbacks, loggers) without this package having to know about it.
//
// # Wiring
//
// Each document type owns one Schema instance, typically a package-level var:
//
//	var globalSchema = versionedjson.New[Data, migrationCtx](CurrentDataVersion).
//	    WithMissingVersion(0)
//
// The first type parameter is the schema's current target type - parameterise it on the
// exported alias (Data here) rather than introducing a separate "Current" name; the package's
// own conventions don't use a <type>Current alias layer.
//
// Every version's file registers its decoder in init():
//
//	func init() {
//	    globalSchema.Register(0, versionedjson.DecoderFor[dataV0]())
//	}
//
// Every non-current version implements Up on its struct, returning the next version's value:
//
//	func (d dataV0) Up(ctx migrationCtx) (any, error) {
//	    return dataV1{...}, nil
//	}
//
// The Load call decodes, walks the Up chain, and returns the migrated value plus the source
// version (so the caller can decide whether to persist the new shape).
package versionedjson

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Upgrader is implemented by every schema version that is NOT the current target. Each
// implementation's Up method takes the same concrete Ctx type the owning Schema was
// instantiated with, so the type-parameterized interface assertion succeeds at runtime.
type Upgrader[Ctx any] interface {
	Up(ctx Ctx) (any, error)
}

// Probe carries the on-disk "version" key as a *uint plus the StampVersion hook that
// Schema.Load / Migrate call after walking up to Current. Designed to be embedded by every
// versioned struct so neither the field nor the method has to be repeated per version:
//
//	type dataV2 struct {
//	    versionedjson.Probe
//	    Region region.Item `json:"region"`
//	}
//
// JSON marshaling of the outer struct produces "version" at the top level because
// encoding/json promotes embedded fields; the StampVersion method is similarly promoted, so
// the embedded type satisfies versionStamper without an explicit method on the outer struct.
type Probe struct {
	Version *uint `json:"version,omitempty"`
}

// StampVersion is the auto-stamp hook. Pointer receiver so the assignment is visible to the
// outer struct that embedded Probe.
func (p *Probe) StampVersion(v uint) { p.Version = &v }

// versionStamper is the unexported interface Schema.Load / Migrate use to dispatch the
// post-migration stamp. Satisfied automatically by any struct that embeds Probe - external
// code should not implement it directly; embed Probe instead.
type versionStamper interface {
	StampVersion(v uint)
}

// Decoder turns raw JSON bytes into the value of one schema version. Use DecoderFor[T] for
// the standard "json.Unmarshal into struct T" case.
type Decoder func([]byte) (any, error)

// DecoderFor builds the standard Decoder that unmarshals raw bytes into struct T. Returned
// by-value so each call produces an independent closure.
func DecoderFor[T any]() Decoder {
	return func(b []byte) (any, error) {
		var d T
		return d, json.Unmarshal(b, &d)
	}
}

// Schema holds the chain of registered decoders for one document type. Current is the latest
// target type (Load returns a value of this type); Ctx is the migration-context type the Up
// methods accept. Instantiate via New, configure via WithMissingVersion / Register.
type Schema[Current any, Ctx any] struct {
	current        uint
	missingVersion uint
	registry       map[uint]Decoder
}

// New constructs a Schema whose current target version is currentVersion. By default a
// document with no "version" key is treated as already at currentVersion - call
// WithMissingVersion to override for chains whose oldest pre-versioning shape used no key
// (typically v0).
func New[Current any, Ctx any](currentVersion uint) *Schema[Current, Ctx] {
	return &Schema[Current, Ctx]{
		current:        currentVersion,
		missingVersion: currentVersion,
		registry:       map[uint]Decoder{},
	}
}

// WithMissingVersion sets the source version assumed when the JSON has no "version" key.
// Returns the schema so it chains nicely with the constructor at package-var init.
func (s *Schema[Current, Ctx]) WithMissingVersion(v uint) *Schema[Current, Ctx] {
	s.missingVersion = v
	return s
}

// Register adds a decoder for one schema version. Panics on duplicate registration so a
// programming mistake surfaces at init() time rather than load.
func (s *Schema[Current, Ctx]) Register(version uint, d Decoder) *Schema[Current, Ctx] {
	if _, dup := s.registry[version]; dup {
		panic(errors.Errorf("versionedjson: duplicate registration for version %d", version))
	}
	s.registry[version] = d
	return s
}

// CurrentVersion returns the target version this schema migrates up to.
func (s *Schema[Current, Ctx]) CurrentVersion() uint { return s.current }

// LoadResult bundles the migrated value with the source version so callers can decide whether
// to persist the new shape (Migrated reports source < current).
type LoadResult[Current any] struct {
	Value         Current
	SourceVersion uint
}

// Migrated reports whether at least one Up step ran during Load. Callers typically use this
// to decide if they should rewrite the file in the new shape (with an optional backup of the
// original bytes).
func (r LoadResult[Current]) Migrated(currentVersion uint) bool {
	return r.SourceVersion < currentVersion
}

// Load decodes raw into the version indicated by the JSON's "version" key (or the missing-
// version default), then walks the Up chain until reaching Current. The ctx is threaded into
// each Up call - each version's Up method receives it as the Ctx type the schema was
// instantiated with.
func (s *Schema[Current, Ctx]) Load(raw []byte, ctx Ctx) (LoadResult[Current], error) {
	var probe Probe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return LoadResult[Current]{}, errors.Wrap(err, "versionedjson: probe version")
	}
	sourceVersion := s.missingVersion
	if probe.Version != nil {
		sourceVersion = *probe.Version
	}

	decode, ok := s.registry[sourceVersion]
	if !ok {
		return LoadResult[Current]{}, errors.Errorf("versionedjson: no decoder for version %d", sourceVersion)
	}
	decoded, err := decode(raw)
	if err != nil {
		return LoadResult[Current]{}, errors.Wrapf(err, "versionedjson: decode version %d", sourceVersion)
	}

	value, _, err := migrateAny(s, decoded, ctx)
	if err != nil {
		return LoadResult[Current]{}, err
	}
	return LoadResult[Current]{Value: value, SourceVersion: sourceVersion}, nil
}

// Chain is the family-marker interface every version in one schema chain implements with the
// same C type argument. C is an arbitrary discriminator type (typically an empty struct, one
// per chain) - its only job is to make compile-time mismatches between a schema and a value
// caught at call sites of Migrate. The runtime payload of Chain() is irrelevant; only the
// returned type matters to the type system.
//
//	type globalFamily struct{}
//	func (dataV0) Chain() globalFamily { return globalFamily{} }
//	func (dataV1) Chain() globalFamily { return globalFamily{} }
//	func (dataV2) Chain() globalFamily { return globalFamily{} }
//
// Once those methods exist, `Migrate(globalSchema, dataV0{}, ctx)` compiles and
// `Migrate(globalSchema, repoProfileV1{}, ctx)` does not - the From type's Chain[C] doesn't
// match the schema Current's Chain[C].
type Chain[C any] interface {
	Chain() C
}

// Migrate walks the Up chain on an already-decoded value until reaching Current. Useful when
// JSON unmarshaling produced a typed value directly as part of a parent struct (e.g. an array
// element with a static older-version type) and you need it brought up to the current schema
// without going back through raw bytes via Load. Returns the migrated Current value plus a
// bool reporting whether any Up step ran - callers commonly use the bool to decide whether
// the surrounding document needs to be rewritten to disk.
//
// If from is already a Current, Migrate stamps its Version (via embedded Probe) and returns
// it with migrated=false. If it's an older version, every intermediate Up runs and
// migrated=true on success.
//
// The From and Current type parameters both carry the Chain[C] bound for the same C, so the
// compiler enforces that the value passed in belongs to the same schema family as the
// schema's target type. Pass a value from a different chain and you get a compile-time
// mismatch, not a runtime "not the current version" panic from the wrong chain's Up.
//
// Package-level (not a method) so all type parameters can be inferred from the call - the
// caller never names them explicitly.
func Migrate[C any, From Chain[C], Current Chain[C], Ctx any](s *Schema[Current, Ctx], from From, ctx Ctx) (Current, bool, error) {
	return migrateAny(s, from, ctx)
}

// migrateAny is the unconstrained walk-up implementation, shared between Load (which has an
// `any` produced by a Decoder) and the public Migrate (which adds the Chain[C] type bound
// for external safety). Internal so the Chain contract isn't accidentally bypassed by
// callers outside the package.
func migrateAny[Current any, Ctx any](s *Schema[Current, Ctx], from any, ctx Ctx) (Current, bool, error) {
	current := from
	migrated := false
	for {
		if final, ok := current.(Current); ok {
			if vs, ok := any(&final).(versionStamper); ok {
				vs.StampVersion(s.current)
			}
			return final, migrated, nil
		}
		up, ok := current.(Upgrader[Ctx])
		if !ok {
			var zero Current
			return zero, false, errors.Errorf("versionedjson: %T is not the current version and has no Up(%T)", current, ctx)
		}
		next, err := up.Up(ctx)
		if err != nil {
			var zero Current
			return zero, false, err
		}
		current = next
		migrated = true
	}
}
