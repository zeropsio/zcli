package versionedjson_test

import (
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/versionedjson"
)

// The test chain models a three-version schema (chainV0 → chainV1 → chainV2) with each Up step
// recording itself in a ctx-supplied slice. Tests assert on the recorded steps to prove the
// migration walked the expected path, plus on the returned value for content shape.

type chainCtx struct {
	steps *[]string
}

func newCtx() chainCtx {
	steps := make([]string, 0)
	return chainCtx{steps: &steps}
}

// chainFamily is the Chain[C] marker for every version in the chainV0→V1→V2 schema. Reused by
// failingV0 and orphan in the error-path tests so they can be passed to versionedjson.Migrate
// against a *Schema[chainV2, chainCtx] without type-mismatch errors.
type chainFamily struct{}

type chainV0 struct {
	versionedjson.Probe
	Name string `json:"name"`
}

func (chainV0) Chain() chainFamily { return chainFamily{} }

func (d chainV0) Up(ctx chainCtx) (any, error) {
	*ctx.steps = append(*ctx.steps, "v0->v1")
	return chainV1{Name: d.Name, Greeting: "hi " + d.Name}, nil
}

type chainV1 struct {
	versionedjson.Probe
	Name     string `json:"name"`
	Greeting string `json:"greeting"`
}

func (chainV1) Chain() chainFamily { return chainFamily{} }

func (d chainV1) Up(ctx chainCtx) (any, error) {
	*ctx.steps = append(*ctx.steps, "v1->v2")
	return chainV2{Name: d.Name, Greeting: d.Greeting, MigratedAt: 2026}, nil
}

type chainV2 struct {
	versionedjson.Probe
	Name       string `json:"name"`
	Greeting   string `json:"greeting"`
	MigratedAt int    `json:"migrated_at"`
}

func (chainV2) Chain() chainFamily { return chainFamily{} }

func newChainSchema() *versionedjson.Schema[chainV2, chainCtx] {
	return versionedjson.New[chainV2, chainCtx](2).
		WithMissingVersion(0).
		Register(0, versionedjson.DecoderFor[chainV0]()).
		Register(1, versionedjson.DecoderFor[chainV1]()).
		Register(2, versionedjson.DecoderFor[chainV2]())
}

// TestLoad_AtCurrentVersion_NoMigration covers the cheap happy path where the on-disk version
// already matches Current: Load returns the decoded value verbatim, SourceVersion equals
// Current, no Up methods run.
func TestLoad_AtCurrentVersion_NoMigration(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":2,"name":"already-current","greeting":"hello","migrated_at":1}`)

	got, err := newChainSchema().Load(raw, ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, got.SourceVersion)
	assert.False(t, got.Migrated(2), "loading at current must not flag migration")
	assert.Equal(t, "already-current", got.Value.Name)
	assert.Empty(t, *ctx.steps, "no Up steps should have fired")
}

// TestLoad_WalksUpChain decodes a v0 document and asserts both that every intermediate Up ran,
// in order, and that the final value carries the data each step contributed.
func TestLoad_WalksUpChain(t *testing.T) {
	ctx := newCtx()
	// No "version" key → uses WithMissingVersion(0), enters at v0.
	raw := []byte(`{"name":"luke"}`)

	got, err := newChainSchema().Load(raw, ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 0, got.SourceVersion)
	assert.True(t, got.Migrated(2), "v0 → v2 must flag migration")
	assert.Equal(t, []string{"v0->v1", "v1->v2"}, *ctx.steps, "every intermediate Up step must run in order")

	assert.Equal(t, "luke", got.Value.Name)
	assert.Equal(t, "hi luke", got.Value.Greeting, "v0→v1 step contributed the greeting")
	assert.Equal(t, 2026, got.Value.MigratedAt, "v1→v2 step contributed MigratedAt")
}

// TestLoad_PartialChain enters at v1 (not the oldest) and walks just the remaining step. Covers
// the case where a user already migrated past v0 in a prior run.
func TestLoad_PartialChain(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":1,"name":"luke","greeting":"yo"}`)

	got, err := newChainSchema().Load(raw, ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 1, got.SourceVersion)
	assert.Equal(t, []string{"v1->v2"}, *ctx.steps, "v0->v1 must not run when entering at v1")
	assert.Equal(t, "yo", got.Value.Greeting, "v1 fields survive the v1→v2 step")
}

// TestLoad_MissingVersion_DefaultsToCurrent: without WithMissingVersion, a doc with no
// "version" key is assumed to already be at Current. Verifies the constructor's default.
func TestLoad_MissingVersion_DefaultsToCurrent(t *testing.T) {
	ctx := newCtx()
	schema := versionedjson.New[chainV2, chainCtx](2).
		Register(2, versionedjson.DecoderFor[chainV2]())

	raw := []byte(`{"name":"luke","greeting":"hi","migrated_at":7}`)
	got, err := schema.Load(raw, ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 2, got.SourceVersion, "missing version defaults to current when no override is set")
	assert.Empty(t, *ctx.steps)
}

// TestLoad_NoDecoder_Errors: a JSON version key pointing at a number with no registered
// decoder must surface a clear error rather than silently dropping data.
func TestLoad_NoDecoder_Errors(t *testing.T) {
	raw := []byte(`{"version":99,"name":"luke"}`)
	_, err := newChainSchema().Load(raw, newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no decoder for version 99")
}

// TestLoad_NoUp_Errors: an intermediate version registered without an Up method is a
// programming mistake. Load must fail with a descriptive error rather than infinite-loop or
// silently return the wrong type.
func TestLoad_NoUp_Errors(t *testing.T) {
	type strandedV0 struct {
		versionedjson.Probe
	}
	schema := versionedjson.New[chainV2, chainCtx](2).
		WithMissingVersion(0).
		Register(0, versionedjson.DecoderFor[strandedV0]()).
		Register(2, versionedjson.DecoderFor[chainV2]())

	raw := []byte(`{"version":0}`)
	_, err := schema.Load(raw, newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not the current version")
}

// TestLoad_UpErrorPropagates: an error returned by an Up method must surface unchanged so the
// caller sees the migration's own diagnostic, not a generic wrapper.
func TestLoad_UpErrorPropagates(t *testing.T) {
	schema := versionedjson.New[chainV2, chainCtx](2).
		WithMissingVersion(0).
		Register(0, versionedjson.DecoderFor[failingV0]()).
		Register(2, versionedjson.DecoderFor[chainV2]())

	_, err := schema.Load([]byte(`{"version":0}`), newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated migration failure")
}

type failingV0 struct {
	versionedjson.Probe
}

func (failingV0) Up(_ chainCtx) (any, error) {
	return nil, errors.New("simulated migration failure")
}

// failingV0 shares the chainFamily marker so it can be fed through the chainV2 schema in
// error-path tests without a family mismatch.
func (failingV0) Chain() chainFamily { return chainFamily{} }

// stampedV0 and stampedV2 model the auto-stamp test chain. stampedV2 is the current target
// and inherits the auto-stamp hook by embedding Probe; stampedV0 is the source with an Up
// step that yields stampedV2 directly (single-step migration is enough to exercise auto-stamp).

// stampedFamily isolates the auto-stamp chain from chainFamily / childFamily so the type
// system also catches accidental cross-schema migrations among the test fixtures.
type stampedFamily struct{}

type stampedV0 struct {
	versionedjson.Probe
}

func (stampedV0) Up(_ chainCtx) (any, error) {
	return stampedV2{}, nil
}

func (stampedV0) Chain() stampedFamily { return stampedFamily{} }

type stampedV2 struct {
	versionedjson.Probe
}

func (stampedV2) Chain() stampedFamily { return stampedFamily{} }

// Nested-schema fixtures: a parent document carrying its child as a TYPED field (not raw
// JSON). The static slot pins the on-disk child shape to whatever version was current at the
// time the parent shape was defined; the parent's Up never touches the child. The caller
// orchestrates the two chains by running the child schema's Migrate on the typed child after
// loading the parent. This mirrors the real composition in `profile` (repoFileV1.Profiles is
// []repoProfileV1, and loadRepo runs repoProfileSchema.Migrate on each entry).

// childFamily is the Chain[C] marker for every child version. Distinct from chainFamily so
// the compiler will reject `Migrate(newChildSchema(), chainV0{}, ctx)` (and vice versa).
type childFamily struct{}

type childV1 struct {
	versionedjson.Probe
	Note string `json:"note"`
}

func (childV1) Chain() childFamily { return childFamily{} }

func (c childV1) Up(ctx chainCtx) (any, error) {
	*ctx.steps = append(*ctx.steps, "child v1->v2")
	return childV2{Note: c.Note, Annotated: "annotated:" + c.Note}, nil
}

type childV2 struct {
	versionedjson.Probe
	Note      string `json:"note"`
	Annotated string `json:"annotated"`
}

func (childV2) Chain() childFamily { return childFamily{} }

type parentV1 struct {
	versionedjson.Probe
	Title string  `json:"title"`
	Child childV1 `json:"child"`
}

// parentV1.Up only touches parent fields. The child slot's static type stays childV1 across
// the parent bump, so callers cascade by handing parentRes.Value.Child to the child schema's
// Migrate themselves - keeps the two chains independent.
func (p parentV1) Up(ctx chainCtx) (any, error) {
	*ctx.steps = append(*ctx.steps, "parent v1->v2")
	return parentV2{Title: p.Title, TitleUpper: p.Title + "!", Child: p.Child}, nil
}

type parentV2 struct {
	versionedjson.Probe
	Title      string  `json:"title"`
	TitleUpper string  `json:"title_upper"`
	Child      childV1 `json:"child"`
}

func newParentSchema() *versionedjson.Schema[parentV2, chainCtx] {
	return versionedjson.New[parentV2, chainCtx](2).
		Register(1, versionedjson.DecoderFor[parentV1]()).
		Register(2, versionedjson.DecoderFor[parentV2]())
}

func newChildSchema() *versionedjson.Schema[childV2, chainCtx] {
	return versionedjson.New[childV2, chainCtx](2).
		Register(1, versionedjson.DecoderFor[childV1]()).
		Register(2, versionedjson.DecoderFor[childV2]())
}

// TestNestedSchemas_ParentUpgrades_ChildStays: source parent V1 + child V1. The caller loads
// the parent (parent.Up runs) and deliberately skips the child schema, leaving the typed
// child in the parent slot at v1. Demonstrates that the two chains are independently
// invokable - "keeping the subchain at vN" just means not running it.
func TestNestedSchemas_ParentUpgrades_ChildStays(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":1,"title":"hi","child":{"version":1,"note":"hello"}}`)

	parentRes, err := newParentSchema().Load(raw, ctx)
	require.NoError(t, err)
	// Intentionally do NOT call newChildSchema().Migrate - this is the "child stays" path.

	assert.Equal(t, []string{"parent v1->v2"}, *ctx.steps, "only parent migrates")
	assert.True(t, parentRes.Migrated(2))
	assert.Equal(t, "hi!", parentRes.Value.TitleUpper, "parent Up ran")
	assert.Equal(t, "hello", parentRes.Value.Child.Note, "child stays as childV1 in the parent slot")
}

// TestNestedSchemas_ChildUpgrades_ParentStays: source parent V2 + child V1. The parent file
// is already current so its Load is a no-op; the caller hands the typed child to the child
// schema's Migrate, which walks v1->v2.
func TestNestedSchemas_ChildUpgrades_ParentStays(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":2,"title":"hi","title_upper":"HI","child":{"version":1,"note":"hello"}}`)

	parentRes, err := newParentSchema().Load(raw, ctx)
	require.NoError(t, err)
	child, childMigrated, err := versionedjson.Migrate(newChildSchema(), parentRes.Value.Child, ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"child v1->v2"}, *ctx.steps, "only child migrates")
	assert.False(t, parentRes.Migrated(2))
	assert.True(t, childMigrated)
	assert.Equal(t, "HI", parentRes.Value.TitleUpper, "parent v2 fields pass through untouched")
	assert.Equal(t, "annotated:hello", child.Annotated, "child Up ran")
}

// TestNestedSchemas_OnlyChildMigrates_ParentBypassed: source parent V1 + child V1. The caller
// decodes the raw bytes directly into parentV1 (bypassing the parent schema entirely) and
// runs only the child schema's Migrate on the embedded typed child. The parent stays at v1 -
// no parent.Up runs - and only the child reaches its current version. This covers the
// "migrate child without committing parent forward" path; it's the symmetric counterpart of
// TestNestedSchemas_ParentUpgrades_ChildStays (which migrates only the parent).
func TestNestedSchemas_OnlyChildMigrates_ParentBypassed(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":1,"title":"hi","child":{"version":1,"note":"hello"}}`)

	var parent parentV1
	require.NoError(t, json.Unmarshal(raw, &parent))
	child, childMigrated, err := versionedjson.Migrate(newChildSchema(), parent.Child, ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"child v1->v2"}, *ctx.steps, "only child migrates, parent schema not invoked")
	assert.True(t, childMigrated)
	assert.Equal(t, "hi", parent.Title, "parent stays at v1, untouched by any Up step")
	assert.Equal(t, "annotated:hello", child.Annotated, "child Up ran independently")
}

// TestNestedSchemas_BothUpgrade: source parent V1 + child V1. The caller runs both - parent's
// Load (parent.Up) and the child schema's Migrate on the typed child (child.Up). Trace
// records the composite migration in caller-invocation order.
func TestNestedSchemas_BothUpgrade(t *testing.T) {
	ctx := newCtx()
	raw := []byte(`{"version":1,"title":"hi","child":{"version":1,"note":"hello"}}`)

	parentRes, err := newParentSchema().Load(raw, ctx)
	require.NoError(t, err)
	child, childMigrated, err := versionedjson.Migrate(newChildSchema(), parentRes.Value.Child, ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"parent v1->v2", "child v1->v2"}, *ctx.steps)
	assert.True(t, parentRes.Migrated(2))
	assert.True(t, childMigrated)
	assert.Equal(t, "hi!", parentRes.Value.TitleUpper)
	assert.Equal(t, "annotated:hello", child.Annotated)
}

// TestLoad_InvalidJSON_Errors covers the case where the bytes aren't even valid JSON: probing
// the version itself fails, which Load wraps with context.
func TestLoad_InvalidJSON_Errors(t *testing.T) {
	_, err := newChainSchema().Load([]byte(`not json`), newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "probe version")
}

// TestRegister_DuplicatePanics: registering the same version twice is a programming mistake
// (e.g. two init() blocks landing on the same number) and must fail loudly at wiring time.
func TestRegister_DuplicatePanics(t *testing.T) {
	schema := versionedjson.New[chainV2, chainCtx](2)
	schema.Register(1, versionedjson.DecoderFor[chainV1]())
	assert.Panics(t, func() {
		schema.Register(1, versionedjson.DecoderFor[chainV1]())
	})
}

// TestCurrentVersion_ReturnsConfigured is a small smoke test for the accessor used by callers
// that want to stamp the version field on save.
func TestCurrentVersion_ReturnsConfigured(t *testing.T) {
	schema := versionedjson.New[chainV2, chainCtx](2)
	assert.EqualValues(t, 2, schema.CurrentVersion())
}

// TestMigrated_ComparesAgainstArgument: a tiny sanity check on the LoadResult helper. The
// caller passes the target version it cares about, so the helper doesn't need to remember
// schema state.
func TestMigrated_ComparesAgainstArgument(t *testing.T) {
	r := versionedjson.LoadResult[chainV2]{SourceVersion: 1}
	assert.True(t, r.Migrated(2))
	assert.False(t, r.Migrated(1))
	assert.False(t, r.Migrated(0))
}

// TestLoad_AutoStampsVersion proves Schema.Load runs VersionStamper on the way out, both when
// no migration happened and after a walk-up. Saves callers from repeating
// `final.Version = &v` after every Load.
func TestLoad_AutoStampsVersion(t *testing.T) {
	schema := versionedjson.New[stampedV2, chainCtx](2).
		WithMissingVersion(0).
		Register(0, versionedjson.DecoderFor[stampedV0]()).
		Register(2, versionedjson.DecoderFor[stampedV2]())

	t.Run("walks up and stamps", func(t *testing.T) {
		got, err := schema.Load([]byte(`{}`), newCtx())
		require.NoError(t, err)
		require.NotNil(t, got.Value.Version, "auto-stamp must set Version even after migration")
		assert.EqualValues(t, 2, *got.Value.Version)
	})

	t.Run("stamps when already at current", func(t *testing.T) {
		// On-disk version 2 with no Version pointer; auto-stamp sets it after the no-op load.
		got, err := schema.Load([]byte(`{"version":2}`), newCtx())
		require.NoError(t, err)
		require.NotNil(t, got.Value.Version)
		assert.EqualValues(t, 2, *got.Value.Version)
	})
}

// TestMigrate_FromCurrent_NoOp: Migrate of a value already at Current returns it unchanged
// (modulo auto-stamp) and reports migrated=false. Covers the path used by callers that JSON-
// unmarshal a typed array element and just want auto-stamping + a uniform call shape.
func TestMigrate_FromCurrent_NoOp(t *testing.T) {
	ctx := newCtx()
	in := chainV2{Name: "luke"}

	got, migrated, err := versionedjson.Migrate(newChainSchema(), in, ctx)
	require.NoError(t, err)
	assert.False(t, migrated, "current-version value must not flag migrated")
	assert.Equal(t, "luke", got.Name, "current value passes through")
	assert.Empty(t, *ctx.steps, "no Up steps should run")
}

// TestMigrate_FromOlder_WalksUp: passing an older-version value walks the Up chain just like
// Load does, sets migrated=true, and yields the current value populated by every step.
func TestMigrate_FromOlder_WalksUp(t *testing.T) {
	ctx := newCtx()
	in := chainV0{Name: "luke"}

	got, migrated, err := versionedjson.Migrate(newChainSchema(), in, ctx)
	require.NoError(t, err)
	assert.True(t, migrated, "older-version input must flag migrated")
	assert.Equal(t, []string{"v0->v1", "v1->v2"}, *ctx.steps)
	assert.Equal(t, "hi luke", got.Greeting)
	assert.Equal(t, 2026, got.MigratedAt)
}

// TestMigrate_AutoStamps: when Current implements VersionStamper, Migrate stamps the version
// just like Load does. Ensures the two entry points stay symmetric.
func TestMigrate_AutoStamps(t *testing.T) {
	schema := versionedjson.New[stampedV2, chainCtx](2).
		Register(0, versionedjson.DecoderFor[stampedV0]()).
		Register(2, versionedjson.DecoderFor[stampedV2]())

	got, _, err := versionedjson.Migrate(schema, stampedV0{}, newCtx())
	require.NoError(t, err)
	require.NotNil(t, got.Version)
	assert.EqualValues(t, 2, *got.Version)
}

// TestMigrate_NoUp_Errors: an intermediate-only value with no Up method must surface a
// programming-error message rather than infinite-loop. Same path Load exercises. Reuses the
// chainFamily marker so the Migrate call type-checks against a chainV2 schema, leaving the
// runtime check (no Up method) as the only thing that can fail.
func TestMigrate_NoUp_Errors(t *testing.T) {
	schema := versionedjson.New[chainV2, chainCtx](2)

	_, _, err := versionedjson.Migrate(schema, orphan{}, newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not the current version")
}

// orphan is in chainFamily but has no Up method. Used by TestMigrate_NoUp_Errors to verify
// the runtime check when a value is neither Current nor an Upgrader.
type orphan struct{}

func (orphan) Chain() chainFamily { return chainFamily{} }

// TestMigrate_RejectsCrossFamily documents the compile-time safety the Chain[C] bound buys.
// The body has no runtime assertions because the check happens in the compiler - the
// commented-out call below would fail to type-check ("childV1 does not satisfy Chain[C]"
// because childV1 is in childFamily, not chainFamily). Kept around as executable
// documentation: uncomment to verify the compiler rejects it.
func TestMigrate_RejectsCrossFamily(t *testing.T) {
	chainSchema := newChainSchema()
	_ = chainSchema

	// Uncommenting the next line should produce a compile error of the form
	// "in call to versionedjson.Migrate, From (type childV1) does not satisfy
	//  versionedjson.Chain[C] (have method Chain() childFamily, want Chain() chainFamily)".
	//
	// _, _, _ = versionedjson.Migrate(chainSchema, childV1{Note: "wrong family"}, newCtx())
}

// TestMigrate_UpErrorPropagates: an error from an Up call must surface unchanged.
func TestMigrate_UpErrorPropagates(t *testing.T) {
	schema := versionedjson.New[chainV2, chainCtx](2)

	_, _, err := versionedjson.Migrate(schema, failingV0{}, newCtx())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated migration failure")
}

// TestLoad_NoStamperIsHarmless: a Current that does NOT implement VersionStamper still loads
// fine; auto-stamping just doesn't happen. Verifies the interface is opt-in.
func TestLoad_NoStamperIsHarmless(t *testing.T) {
	type noStampV1 struct {
		Name string `json:"name"`
	}
	schema := versionedjson.New[noStampV1, chainCtx](1).
		Register(1, versionedjson.DecoderFor[noStampV1]())

	got, err := schema.Load([]byte(`{"version":1,"name":"hi"}`), newCtx())
	require.NoError(t, err)
	assert.Equal(t, "hi", got.Value.Name)
}

// TestDecoderFor_UnmarshalsIntoTargetType covers the small convenience helper. Independent of
// the schema/Load machinery so it's clear which side of the abstraction broke if it fails.
func TestDecoderFor_UnmarshalsIntoTargetType(t *testing.T) {
	decode := versionedjson.DecoderFor[chainV2]()
	got, err := decode([]byte(`{"name":"luke","migrated_at":42}`))
	require.NoError(t, err)
	v, ok := got.(chainV2)
	require.True(t, ok)
	assert.Equal(t, "luke", v.Name)
	assert.Equal(t, 42, v.MigratedAt)
}
