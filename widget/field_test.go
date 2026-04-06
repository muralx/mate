package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var _ Container = (*Field)(nil)

func TestField_Defaults(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	if f.ID() != "f" {
		t.Errorf("ID = %q", f.ID())
	}
	if !f.Visible() {
		t.Error("should be visible")
	}
	if !f.Enabled() {
		t.Error("should be enabled")
	}
}

func TestField_HasThreeChildren(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	if len(f.Children()) != 3 {
		t.Errorf("children = %d, want 3 (label + separator + input)", len(f.Children()))
	}
}

func TestField_ChildrenParentSet(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	for _, child := range f.Children() {
		if child.Parent() != f {
			t.Errorf("child %s parent should be the field", child.ID())
		}
	}
}

func TestField_InnerFocused_NoFocus(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	if f.InnerFocused() {
		t.Error("no inner focus expected")
	}
}

func TestField_InnerFocused_InputFocused(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	btn.SetFocused(true)
	if !f.InnerFocused() {
		t.Error("field should have inner focus when input focused")
	}
}

func TestField_View_ContainsAll(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	output := stripansi.Strip(f.View())
	if !strings.Contains(output, "Name") {
		t.Errorf("missing label in %q", output)
	}
	if !strings.Contains(output, ": ") {
		t.Errorf("missing separator in %q", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("missing input in %q", output)
	}
}

func TestField_View_LabelHighlightsOnFocus(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())

	normalView := f.View()

	btn.SetFocused(true)
	hotView := f.View()

	if normalView == hotView {
		t.Error("label should look different when input is focused")
	}
}

func TestField_View_SeparatorNeverHighlights(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())

	f.View()
	sepNormal := f.Separator().View()

	btn.SetFocused(true)
	f.View()
	sepFocused := f.Separator().View()

	if sepNormal != sepFocused {
		t.Error("separator should not change when focus changes")
	}
}

func TestField_Active_Propagation(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	f.SetEnabled(false)
	if btn.Active() {
		t.Error("input should be inactive when field disabled")
	}
}

func TestField_Label(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	label := f.Label()
	if label == nil {
		t.Fatal("Label() should not return nil")
	}
	if label.GetText() != "Name" {
		t.Errorf("Label text = %q, want %q", label.GetText(), "Name")
	}
}

func TestField_Input(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	input := f.Input()
	if input == nil {
		t.Fatal("Input() should not return nil")
	}
	if input.ID() != "btn" {
		t.Errorf("Input ID = %q, want %q", input.ID(), "btn")
	}
}

func TestField_View_Invisible(t *testing.T) {
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	f := NewField("f", "Name", btn, DefaultFieldStyles())
	f.SetVisible(false)
	if f.View() != "" {
		t.Error("invisible field should return empty")
	}
}
