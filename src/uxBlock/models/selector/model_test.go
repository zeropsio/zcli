package selector

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/uxBlock/models/table"
)

const testTermWidth, testTermHeight = 80, 24

// newSelectorWithRows builds a selector root model populated with the given
// rows. Each row is a single-column "name". Used as the fixture for all
// teatest-driven UI tests in this package.
func newSelectorWithRows(t *testing.T, rows ...string) *RootModel {
	t.Helper()
	body := table.NewBody()
	for _, r := range rows {
		body.AddStringsRow(r)
	}
	return NewRoot(t.Context(), body, WithLabel("Pick one"))
}

// Pressing the down arrow then enter selects the row at index 1.
func TestSelector_DownThenEnterPicksSecondRow(t *testing.T) {
	root := newSelectorWithRows(t, "alpha", "beta", "gamma")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final, ok := tm.FinalModel(t).(*RootModel)
	require.True(t, ok, "final model should be *RootModel")
	require.NoError(t, final.Err(), "selector should exit cleanly")
	selected := final.Selected()
	require.Len(t, selected, 1, "expected exactly one selection")
	assert.Equal(t, 1, selected[0], "down+enter should pick the second row (index 1)")
}

// Esc (or Ctrl+C) aborts the selector with models.ErrCtrlC.
func TestSelector_EscReturnsErrCtrlC(t *testing.T) {
	root := newSelectorWithRows(t, "alpha", "beta")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final, ok := tm.FinalModel(t).(*RootModel)
	require.True(t, ok)
	require.Error(t, final.Err(), "esc should produce a cancellation error")
}

// Wrapping behavior: pressing down past the last row stays on the last row.
// Locks the "no wrap" contract of the selector — useful to catch accidental
// `cursor = (cursor + 1) % len` regressions.
func TestSelector_DownPastEndStaysOnLastRow(t *testing.T) {
	root := newSelectorWithRows(t, "alpha", "beta", "gamma")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	for i := 0; i < 10; i++ {
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final, ok := tm.FinalModel(t).(*RootModel)
	require.True(t, ok)
	require.NoError(t, final.Err())
	assert.Equal(t, 2, final.Selected()[0], "cursor should clamp at the last row (index 2)")
}

// --- multi-select mode ---------------------------------------------------

// newMultiSelectorWithRows builds a multi-select selector. Spacebar toggles
// the row at the cursor, Enter exits and returns all toggled indices.
func newMultiSelectorWithRows(t *testing.T, rows ...string) *RootModel {
	t.Helper()
	body := table.NewBody()
	for _, r := range rows {
		body.AddStringsRow(r)
	}
	return NewRoot(t.Context(), body,
		WithLabel("Pick many"),
		WithEnableMultiSelect(),
	)
}

// Spacebar toggles individual rows; Enter then returns all toggled indices
// via GetMultipleSelectedFunc.
func TestSelector_MultiSelect_SpaceTogglesRowsThenEnterReturnsAll(t *testing.T) {
	root := newMultiSelectorWithRows(t, "alpha", "beta", "gamma")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	// Cursor starts at 0: toggle alpha.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	// Down twice to gamma, toggle it.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	// Enter to confirm.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	final := tm.FinalModel(t)
	selected, err := GetMultipleSelectedFunc(final)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 2}, selected, "alpha (0) and gamma (2) should be selected")
}

// Toggling the same row twice with space deselects it.
func TestSelector_MultiSelect_SpaceTwiceDeselects(t *testing.T) {
	root := newMultiSelectorWithRows(t, "alpha", "beta")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}) // toggle alpha on
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}) // toggle alpha off
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	selected, err := GetMultipleSelectedFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Empty(t, selected, "two toggles should leave the row unselected")
}

// Ctrl+A selects every row at once; Enter returns all indices in order.
func TestSelector_MultiSelect_CtrlASelectsAll(t *testing.T) {
	root := newMultiSelectorWithRows(t, "alpha", "beta", "gamma")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	selected, err := GetMultipleSelectedFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2}, selected)
}

// Ctrl+D clears any current selection (regardless of prior state).
func TestSelector_MultiSelect_CtrlDClearsSelection(t *testing.T) {
	root := newMultiSelectorWithRows(t, "alpha", "beta", "gamma")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlA}) // select all
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlD}) // then clear
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	selected, err := GetMultipleSelectedFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Empty(t, selected, "ctrl+d should clear all selections")
}

// Calling GetOneSelectedFunc on a multi-select model is a programmer error
// and must surface as a non-nil error rather than panic.
func TestSelector_MultiSelect_GetOneSelectedRejectsMultiMode(t *testing.T) {
	root := newMultiSelectorWithRows(t, "alpha", "beta")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	_, err := GetOneSelectedFunc(tm.FinalModel(t))
	require.Error(t, err, "GetOneSelectedFunc must reject multi-select models")
}

// Filter-mode round trip: "/" enters filter mode, typed chars narrow the
// rows, first Enter confirms the filter and resets the cursor, second Enter
// selects the cursor-row. The returned Selected() index is the original
// (unfiltered) row index.
func TestSelector_FilterModeNarrowsAndSelects(t *testing.T) {
	body := table.NewBody()
	body.AddStringsRow("alpha")
	body.AddStringsRow("beta")
	body.AddStringsRow("gamma")
	root := NewRoot(t.Context(), body, WithEnableFiltering())

	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}) // enter filter mode
	tm.Type("be")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // exit filter mode
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // confirm selection

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final, ok := tm.FinalModel(t).(*RootModel)
	require.True(t, ok)
	require.NoError(t, final.Err())
	selected := final.Selected()
	require.Len(t, selected, 1)
	assert.Equal(t, 1, selected[0], "filter 'be' should isolate 'beta' (original index 1)")
}

// The label passed via WithLabel renders in the initial frame when
// filtering is enabled (model.go's View only emits the label alongside the
// filter input — otherwise the filter placeholder overwrites it).
func TestSelector_LabelAppearsInOutput(t *testing.T) {
	body := table.NewBody()
	body.AddStringsRow("alpha")
	root := NewRoot(t.Context(), body,
		WithLabel("Pick one"),
		WithEnableFiltering(),
	)

	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	// Wait for the initial frame before quitting; without this, Esc can be
	// processed before bubbletea flushes anything.
	teatest.WaitFor(t, tm.Output(),
		func(b []byte) bool { return bytes.Contains(b, []byte("Pick one")) },
		teatest.WithDuration(2*time.Second),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}
