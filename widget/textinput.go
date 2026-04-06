package widget

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInput is a focusable text input box. It wraps bubbles/textinput
// and handles cursor, placeholder, and character input.
// It does NOT render a label — use Field for label + separator + input.
type TextInput struct {
	FocusableComponent
	inputWidth int
	onSubmit   func(string) tea.Cmd
	onChange   func(string) tea.Cmd
	input      textinput.Model
	err        string // validation error, set externally
}

// NewTextInput creates a new TextInput with the given ID and input width.
func NewTextInput(id string, inputWidth int) *TextInput {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Width = inputWidth
	ti.CharLimit = 200

	t := &TextInput{
		inputWidth: inputWidth,
		input:      ti,
	}
	t.FocusableComponent = NewFocusableComponent(id)
	t.SetPreferredWidth(inputWidth)
	return t
}

// WithPlaceholder sets placeholder text displayed when the field is empty.
func (t *TextInput) WithPlaceholder(p string) *TextInput {
	t.input.Placeholder = p
	return t
}

// WithCharLimit sets the character limit for the input.
func (t *TextInput) WithCharLimit(n int) *TextInput {
	t.input.CharLimit = n
	return t
}

// OnSubmit sets the callback invoked when Enter is pressed.
func (t *TextInput) OnSubmit(fn func(string) tea.Cmd) { t.onSubmit = fn }

// OnChange sets the callback invoked when the input value changes.
func (t *TextInput) OnChange(fn func(string) tea.Cmd) { t.onChange = fn }

// SetValue sets the input value.
func (t *TextInput) SetValue(v string) { t.input.SetValue(v) }

// Value returns the current input value.
func (t *TextInput) Value() string { return t.input.Value() }

// SetError sets a validation error message.
func (t *TextInput) SetError(e string) { t.err = e }

// Error returns the current validation error.
func (t *TextInput) Error() string { return t.err }

// SetFocused sets the focused state and focuses/blurs the inner textinput.
func (t *TextInput) SetFocused(v bool) tea.Cmd {
	t.BaseComponent.SetFocused(v)
	if v {
		return t.input.Focus()
	}
	t.input.Blur()
	return nil
}

// Update handles key input. Enter triggers onSubmit; all other keys are
// delegated to the inner textinput. Returns (cmd, true) consuming ALL input
// when active. If not active, returns (nil, false).
func (t *TextInput) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !t.Active() {
		return nil, false
	}
	switch msg.Type {
	case tea.KeyEnter:
		if t.onSubmit != nil {
			return t.onSubmit(t.input.Value()), true
		}
		return nil, true
	case tea.KeySpace:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	}
	oldVal := t.input.Value()
	var cmd tea.Cmd
	t.input, cmd = t.input.Update(msg)
	if t.onChange != nil && t.input.Value() != oldVal {
		t.onChange(t.input.Value())
	}
	return cmd, true
}

// SetSize overrides BaseComponent to adjust the inner input width.
func (t *TextInput) SetSize(w, h int) {
	t.BaseComponent.SetSize(w, h)
	if w > 0 {
		t.input.Width = w
	}
}

// View renders the input area only (no label).
func (t *TextInput) View() string {
	if !t.Active() {
		rendered := lipgloss.NewStyle().Faint(true).Render(t.input.View())
		return t.RenderInSize(rendered)
	}
	return t.RenderInSize(t.input.View())
}

// HandleEvent — text inputs do not react to clicks (just focus).
func (t *TextInput) HandleEvent(event Event) (tea.Cmd, bool) {
	return t.BaseComponent.HandleEvent(event)
}
