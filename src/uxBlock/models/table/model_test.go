package table

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Render -------------------------------------------------------------

// All row cells must appear verbatim in the rendered string. The lipgloss
// table is the only renderer; this test guards against regressions in the
// cell-text pipeline (Cell.String() → row builder → table.Rows).
func TestRender_RowCellsAppearInOutput(t *testing.T) {
	body := NewBodyFromStrings(
		[]string{"id-1", "alpha", "ACTIVE"},
		[]string{"id-2", "beta", "PENDING"},
	)

	out := Render(body)

	for _, want := range []string{"id-1", "alpha", "ACTIVE", "id-2", "beta", "PENDING"} {
		assert.Containsf(t, out, want, "rendered table should contain %q", want)
	}
}

// Header cells are uppercased before rendering (model.go:42-44).
func TestRender_HeaderIsUppercased(t *testing.T) {
	header := NewRowFromStrings("id", "Name", "status")
	body := NewBodyFromStrings([]string{"x", "y", "z"})

	out := Render(body, WithHeader(header))

	for _, want := range []string{"ID", "NAME", "STATUS"} {
		assert.Containsf(t, out, want, "header %q should be uppercased", want)
	}
	// And the lowercase form of the mixed-case input must not leak through.
	assert.NotContains(t, out, "Name", "mixed-case header should be uppercased, not preserved")
}

// An empty body renders without crashing. lipgloss's table returns an empty
// string when there are no rows AND no header; this test just guards against
// future panics in that branch (e.g. project list with zero projects).
func TestRender_EmptyBodyDoesNotCrash(t *testing.T) {
	body := NewBody()
	require.NotPanics(t, func() { _ = Render(body) })
}

// --- Body / Row indexing -----------------------------------------------

// AddRow assigns sequential indices via the row's index field. The selector
// model relies on Row.Index() to map filtered rows back to original rows.
func TestBody_AddRowAssignsSequentialIndices(t *testing.T) {
	body := NewBody()
	body.AddStringsRow("a")
	body.AddStringsRow("b")
	body.AddStringsRow("c")

	rows := body.Rows()
	require.Len(t, rows, 3)
	assert.Equal(t, 0, rows[0].Index())
	assert.Equal(t, 1, rows[1].Index())
	assert.Equal(t, 2, rows[2].Index())
}

// A row constructed with an explicit index (via NewRow + manual index setup)
// keeps that index when added to a body; this is the path used by filter
// re-builds in the selector model.
func TestBody_AddRow_PreservesPresetIndex(t *testing.T) {
	r := NewRowFromStrings("x")
	r.index = 7 // simulate selector's filtered-body construction
	body := NewBody()
	body.AddRow(r)
	assert.Equal(t, 7, body.Rows()[0].Index(), "preset row index must survive AddRow")
}

// Clone returns a deep copy: mutating a row in the clone must not affect
// the original. Important for the selector filter, which clones the body
// to filter rows without disturbing the source.
func TestBody_CloneIsDeep(t *testing.T) {
	orig := NewBodyFromStrings([]string{"alpha"}, []string{"beta"})
	cl := orig.Clone()

	cl.Rows()[0].AddStringCell("extra")

	assert.Len(t, orig.Rows()[0].Cells(), 1, "original row should be untouched")
	assert.Len(t, cl.Rows()[0].Cells(), 2, "cloned row was extended")
}

// --- Row / Cell helpers ------------------------------------------------

// SetDisabled / IsDisabled is the toggle used by the selector to skip rows
// (e.g. "Create new project" prompt entries).
func TestRow_DisabledRoundTrip(t *testing.T) {
	r := NewRowFromStrings("alpha")
	assert.False(t, r.IsDisabled())
	r.SetDisabled(true)
	assert.True(t, r.IsDisabled())
	r.SetDisabled(false)
	assert.False(t, r.IsDisabled())
}

// AddStringCells appends multiple cells in one call (fluent builder API).
func TestRow_AddStringCellsAppendsAll(t *testing.T) {
	r := NewRowFromStrings("a")
	r.AddStringCells("b", "c", "d")
	assert.Len(t, r.Cells(), 4)
	assert.Equal(t, "d", r.Cells()[3].String())
}

// By default, Cell.String() returns the raw text; SetPretty(true) flips it
// to delegate to Styled(). We can't assert on ANSI escapes here because
// lipgloss strips them when the test process has no color profile, so the
// check is structural: String() == text when not pretty, String() ==
// Styled() when pretty.
func TestCell_StringRespectsPrettyFlag(t *testing.T) {
	plain := NewCell("hello")
	assert.Equal(t, "hello", plain.String(), "non-pretty cell returns raw text")

	bold := NewCell("hi").SetStyle(lipgloss.NewStyle().Bold(true)).SetPretty(true)
	assert.Equal(t, bold.Styled(), bold.String(), "pretty cell should delegate to Styled()")
	assert.Contains(t, bold.String(), "hi", "underlying text still present in pretty output")
}

// Styled() renders the styled string regardless of the pretty flag — used
// directly by callers that want a styled label even on a non-pretty cell.
func TestCell_StyledIgnoresPrettyFlag(t *testing.T) {
	c := NewCell("hi").SetStyle(lipgloss.NewStyle().Bold(true))
	out := c.Styled()
	assert.Contains(t, out, "hi")
}

// Render must wrap cell content in the lipgloss border — a tiny smoke test
// that the table package keeps using bordered tables.
func TestRender_HasBorderGlyphs(t *testing.T) {
	body := NewBodyFromStrings([]string{"alpha"})
	out := Render(body)
	// The default normal lipgloss border uses these characters; if any
	// future change drops the border, this test catches it.
	for _, glyph := range []string{"┌", "─", "┐", "│", "└", "┘"} {
		assert.Containsf(t, out, glyph, "rendered table should contain border glyph %q\n%s", glyph, strings.TrimSpace(out))
	}
}
