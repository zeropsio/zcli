package selector

import (
	"bytes"
	"context"
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
	return NewRoot(context.Background(), body, WithLabel("Pick one"))
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

	final := tm.FinalModel(t).(*RootModel)
	require.NoError(t, final.Err())
	assert.Equal(t, 2, final.Selected()[0], "cursor should clamp at the last row (index 2)")
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
	root := NewRoot(context.Background(), body, WithEnableFiltering())

	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}) // enter filter mode
	tm.Type("be")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // exit filter mode
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // confirm selection

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	final := tm.FinalModel(t).(*RootModel)
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
	root := NewRoot(context.Background(), body,
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
