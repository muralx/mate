package widget

import "github.com/charmbracelet/lipgloss"

// Text is a non-focusable component that renders a styled string.
// It knows nothing about focus, parents, or context — it just displays
// text with whatever style has been set on it. The parent controls
// the style based on application state.
type Text struct {
	BaseComponent
	text  string
	style lipgloss.Style
}

// NewText creates a new Text component with the given ID, text, and style.
func NewText(id, text string, style lipgloss.Style) *Text {
	t := &Text{
		text:  text,
		style: style,
	}
	t.BaseComponent = *NewBaseComponent(id)
	return t
}

// SetText updates the displayed text.
func (t *Text) SetText(text string) { t.text = text }

// GetText returns the current text.
func (t *Text) GetText() string { return t.text }

// SetStyle sets the rendering style.
func (t *Text) SetStyle(s lipgloss.Style) { t.style = s }

// Style returns the current style.
func (t *Text) Style() lipgloss.Style { return t.style }

// View renders the text with its current style.
func (t *Text) View() string {
	if !t.Active() {
		return t.RenderInSize(lipgloss.NewStyle().Faint(true).Render(t.text))
	}
	return t.RenderInSize(t.style.Render(t.text))
}
