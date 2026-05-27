// Fixtures and tests for the "parent has no migration chain yet" composition. The parent
// schema's current target is also its only registered version, so its Load is always a
// no-op pass-through; the embedded child has its own multi-version chain. Exercises the
// case where one side of the nesting is stable while the other evolves - common during the
// window when an entry shape is being iterated but the file wrapper hasn't needed a bump.

package versionedjson_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/versionedjson"
)

// stableParentV1 has no Up method - it IS the current target. The static child slot still
// pins the on-disk shape to childV1, so callers Migrate it independently.
type stableParentV1 struct {
	versionedjson.Probe
	Title string  `json:"title"`
	Child childV1 `json:"child"`
}

func newStableParentSchema() *versionedjson.Schema[stableParentV1, chainCtx] {
	return versionedjson.New[stableParentV1, chainCtx](1).
		Register(1, versionedjson.DecoderFor[stableParentV1]())
}

// TestStableParent_ChildMigratesAlone: source stableParentV1 + child V1. Loading the parent
// schema is a no-op (no Up registered); the caller runs only the child schema's Migrate to
// bring the embedded child up to v2. Demonstrates that a parent with no chain is the natural
// model for "child migrates while parent stays" - no schema bypass needed.
func TestStableParent_ChildMigratesAlone(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":1,"title":"hi","child":{"version":1,"note":"hello"}}`)

	parentRes, err := newStableParentSchema().Load(raw, ctx)
	require.NoError(t, err)
	child, childMigrated, err := versionedjson.Migrate(newChildSchema(), parentRes.Value.Child, ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"child v1->v2"}, *ctx.steps, "no parent Up exists; only child migrates")
	assert.False(t, parentRes.Migrated(1), "stable parent never reports migration")
	assert.True(t, childMigrated)
	assert.Equal(t, "hi", parentRes.Value.Title)
	assert.Equal(t, "hello", parentRes.Value.Child.Note, "child slot in parent stays as childV1")
	assert.Equal(t, "annotated:hello", child.Annotated, "migrated child sits alongside the unchanged parent")
}

// TestStableParent_ChildAlreadyCurrent: source stableParentV1 + child V1 typed slot, but with
// the child JSON carrying childV2 fields that the typed slot drops. The caller's Migrate of
// the typed (childV1) child still walks v1->v2 producing a synthetic Annotated - which is the
// expected behavior, since the typed slot is the source of truth, not the embedded version
// key. Documents the trade-off of typed children: the parent's static slot dictates the
// effective shape, not the JSON.
func TestStableParent_ChildAlreadyCurrent(t *testing.T) {
	ctx := newCtx()
	// Note: child JSON has "version":2 and an "annotated" field, but the typed slot
	// childV1 only picks up Note + Version - the rest is dropped silently. This is a
	// known consequence of typed children documented for callers.
	raw := []byte(`{"version":1,"title":"hi","child":{"version":2,"note":"hello","annotated":"ignored-by-typing"}}`)

	parentRes, err := newStableParentSchema().Load(raw, ctx)
	require.NoError(t, err)

	// The typed slot kept only the v1-compatible fields; the embedded "version":2 marker is
	// retained in the slot's Version field but doesn't change decoding. Running the child
	// schema's Migrate walks v1->v2 because the static type is childV1.
	child, childMigrated, err := versionedjson.Migrate(newChildSchema(), parentRes.Value.Child, ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"child v1->v2"}, *ctx.steps)
	assert.True(t, childMigrated)
	assert.Equal(t, "annotated:hello", child.Annotated, "synthetic value from v1->v2 step wins over the discarded JSON field")
}
