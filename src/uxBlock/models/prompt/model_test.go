package prompt

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTermWidth, testTermHeight = 80, 24

func newPrompt(t *testing.T, message string, choices ...string) *RootModel {
	t.Helper()
	return NewRoot(t.Context(), message, choices)
}

// Right then Enter advances the cursor and confirms; GetChoiceCursor reports
// the new index.
func TestPrompt_RightThenEnterAdvancesCursor(t *testing.T) {
	root := newPrompt(t, "Continue?", "Yes", "No")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyRight})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	final := tm.FinalModel(t)

	cur, err := GetChoiceCursor(final)
	require.NoError(t, err)
	assert.Equal(t, 1, cur)
	val, err := GetChoiceValue(final)
	require.NoError(t, err)
	assert.Equal(t, "No", val)
}

// Multiple Rights past the last choice clamp at len-1 (no wrap).
func TestPrompt_RightPastEndClampsAtLast(t *testing.T) {
	root := newPrompt(t, "Pick", "A", "B", "C")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	for i := 0; i < 10; i++ {
		tm.Send(tea.KeyMsg{Type: tea.KeyRight})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	cur, err := GetChoiceCursor(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Equal(t, 2, cur, "cursor should clamp at last index")
}

// Left at index 0 stays at 0 (no wrap to end).
func TestPrompt_LeftAtZeroStaysAtZero(t *testing.T) {
	root := newPrompt(t, "Pick", "A", "B")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyLeft})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	cur, err := GetChoiceCursor(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Equal(t, 0, cur)
}

// Esc aborts the prompt with the cancellation error.
func TestPrompt_EscReturnsError(t *testing.T) {
	root := newPrompt(t, "Pick", "A", "B")
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	_, err := GetChoiceCursor(tm.FinalModel(t))
	require.Error(t, err)
}

// WithCursorPosition pre-positions the cursor and clamps out-of-range values.
func TestPrompt_WithCursorPositionInitial(t *testing.T) {
	t.Run("in-range", func(t *testing.T) {
		root := NewRoot(t.Context(), "Pick", []string{"A", "B", "C"}, WithCursorPosition(2))
		tm := teatest.NewTestModel(t, root, teatest.WithInitialTermSize(testTermWidth, testTermHeight))
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
		cur, err := GetChoiceCursor(tm.FinalModel(t))
		require.NoError(t, err)
		assert.Equal(t, 2, cur)
	})
	t.Run("over-range-clamps-to-last", func(t *testing.T) {
		root := NewRoot(t.Context(), "Pick", []string{"A", "B", "C"}, WithCursorPosition(99))
		tm := teatest.NewTestModel(t, root, teatest.WithInitialTermSize(testTermWidth, testTermHeight))
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
		cur, err := GetChoiceCursor(tm.FinalModel(t))
		require.NoError(t, err)
		assert.Equal(t, 2, cur)
	})
	t.Run("under-range-clamps-to-zero", func(t *testing.T) {
		root := NewRoot(t.Context(), "Pick", []string{"A", "B", "C"}, WithCursorPosition(-5))
		tm := teatest.NewTestModel(t, root, teatest.WithInitialTermSize(testTermWidth, testTermHeight))
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
		cur, err := GetChoiceCursor(tm.FinalModel(t))
		require.NoError(t, err)
		assert.Equal(t, 0, cur)
	})
}
