package input

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTermWidth, testTermHeight = 80, 24

func newInput(t *testing.T, opts ...Option) *RootModel {
	t.Helper()
	return NewRoot(t.Context(), opts...)
}

// Typing characters then Enter submits the value; GetValueFunc reports it.
func TestInput_TypeThenEnterSubmitsValue(t *testing.T) {
	root := newInput(t)
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Type("demo-service")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	val, err := GetValueFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Equal(t, "demo-service", val)
}

// Enter without any prior input submits an empty value.
func TestInput_EnterWithEmptyValue(t *testing.T) {
	root := newInput(t)
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	val, err := GetValueFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Empty(t, val)
}

// Esc aborts the input and propagates the cancellation error.
func TestInput_EscReturnsError(t *testing.T) {
	root := newInput(t)
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Type("partial")
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	_, err := GetValueFunc(tm.FinalModel(t))
	require.Error(t, err)
}

// Backspace removes the last typed character.
func TestInput_BackspaceErasesPreviousChar(t *testing.T) {
	root := newInput(t)
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	tm.Type("demoX")
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	val, err := GetValueFunc(tm.FinalModel(t))
	require.NoError(t, err)
	assert.Equal(t, "demo", val)
}

// WithLabel renders the label before the input field in the initial frame.
func TestInput_LabelRendered(t *testing.T) {
	root := newInput(t, WithLabel("Type service name"))
	tm := teatest.NewTestModel(t, root,
		teatest.WithInitialTermSize(testTermWidth, testTermHeight),
	)

	teatest.WaitFor(t, tm.Output(),
		func(b []byte) bool { return bytes.Contains(b, []byte("Type service name")) },
		teatest.WithDuration(2*time.Second),
	)

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}
