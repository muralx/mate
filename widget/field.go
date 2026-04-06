package widget

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FieldStyles defines the styles used by a Field.
type FieldStyles struct {
	Label     lipgloss.Style // label when no inner focus
	LabelHot  lipgloss.Style // label when input has focus
	Separator lipgloss.Style // separator style (always the same)
}

// DefaultFieldStyles returns a FieldStyles with sensible defaults.
func DefaultFieldStyles() FieldStyles {
	return FieldStyles{
		Label:     lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		LabelHot:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")),
		Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	}
}

// Field composes a label (Text), separator (Text), and an input component.
// The Field controls the label's style based on focus state — when any
// descendant has focus, the label highlights. The separator never highlights.
//
// Construction: NewField("name", "Name", textBox, styles)
// Renders: <label><separator><input>
type Field struct {
	BaseContainer
	label     *Text     // managed internally
	separator *Text     // managed internally
	input     Component // the input passed in
	styles    FieldStyles
}

// NewField creates a Field with a label, ": " separator, and the given input component.
func NewField(id, labelText string, input Component, styles FieldStyles) *Field {
	f := &Field{
		styles: styles,
	}
	f.BaseContainer = *NewBaseContainer(id, f)

	f.label = NewText(id+"_label", labelText, styles.Label)
	f.separator = NewText(id+"_sep", ": ", styles.Separator)
	f.input = input

	f.AddChild(f.label)
	f.AddChild(f.separator)
	f.AddChild(f.input)

	return f
}

// Label returns the label Text component.
func (f *Field) Label() *Text { return f.label }

// Separator returns the separator Text component.
func (f *Field) Separator() *Text { return f.separator }

// Input returns the input component.
func (f *Field) Input() Component { return f.input }

// View renders label + separator + input horizontally.
// Field controls the label style based on focus state.
func (f *Field) View() string {
	if !f.Visible() {
		return ""
	}

	// Field is the smart one — it sets the label style based on state.
	if f.InnerFocused() {
		f.label.SetStyle(f.styles.LabelHot)
	} else {
		f.label.SetStyle(f.styles.Label)
	}

	// Lay out children left-to-right. Pass availH=1 so children without a
	// preferred height are sized to exactly one row (Field is always single-line).
	fx, fy := f.Position()
	fw, _ := f.Size()
	var visible []Component
	for _, child := range f.Children() {
		if child.Visible() {
			visible = append(visible, child)
		}
	}
	LayoutHorizontal(visible, fx, fy, fw, 1, 0)

	var parts []string
	for _, child := range visible {
		parts = append(parts, child.View())
	}
	return f.RenderInSize(strings.Join(parts, ""))
}
