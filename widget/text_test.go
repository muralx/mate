package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var _ Component = (*Text)(nil)

func TestText_Defaults(t *testing.T) {
	txt := NewText("t", "Hello", lipgloss.NewStyle())
	if txt.ID() != "t" {
		t.Errorf("ID = %q", txt.ID())
	}
	if !txt.Visible() {
		t.Error("should be visible")
	}
	if !txt.Enabled() {
		t.Error("should be enabled")
	}
}

func TestText_NotFocusable(t *testing.T) {
	txt := NewText("t", "Hello", lipgloss.NewStyle())
	if txt.Focusable() {
		t.Error("text should not be focusable")
	}
}

func TestText_View(t *testing.T) {
	txt := NewText("t", "Name:", lipgloss.NewStyle())
	output := stripansi.Strip(txt.View())
	if !strings.Contains(output, "Name:") {
		t.Errorf("view = %q, want 'Name:'", output)
	}
}

func TestText_View_Inactive(t *testing.T) {
	txt := NewText("t", "Name:", lipgloss.NewStyle())
	txt.SetEnabled(false)
	output := stripansi.Strip(txt.View())
	if !strings.Contains(output, "Name:") {
		t.Errorf("inactive should still show text, got %q", output)
	}
}

func TestText_View_Invisible(t *testing.T) {
	txt := NewText("t", "Hidden", lipgloss.NewStyle())
	txt.SetVisible(false)
	// Text itself doesn't check visibility in View — the parent skips it.
	// But the component is still renderable if called directly.
	output := stripansi.Strip(txt.View())
	if !strings.Contains(output, "Hidden") {
		t.Errorf("view = %q", output)
	}
}

func TestText_SetText(t *testing.T) {
	txt := NewText("t", "Before", lipgloss.NewStyle())
	txt.SetText("After")
	output := stripansi.Strip(txt.View())
	if !strings.Contains(output, "After") {
		t.Errorf("view = %q, want 'After'", output)
	}
}

func TestText_SetStyle(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	txt := NewText("t", "Test", lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")))
	output1 := txt.View()

	txt.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")))
	output2 := txt.View()

	if output1 == output2 {
		t.Error("different styles should produce different output")
	}
}

func TestText_KeyBindings_Nil(t *testing.T) {
	txt := NewText("t", "Test", lipgloss.NewStyle())
	if txt.KeyBindings() != nil {
		t.Error("text should have nil key bindings")
	}
}

func TestText_GetText(t *testing.T) {
	txt := NewText("t", "Hello", lipgloss.NewStyle())
	if txt.GetText() != "Hello" {
		t.Errorf("GetText() = %q, want %q", txt.GetText(), "Hello")
	}
	txt.SetText("World")
	if txt.GetText() != "World" {
		t.Errorf("GetText() = %q after SetText, want %q", txt.GetText(), "World")
	}
}

func TestText_Style(t *testing.T) {
	s := lipgloss.NewStyle().Bold(true)
	txt := NewText("t", "Test", s)
	got := txt.Style()
	if got.GetBold() != true {
		t.Error("Style() should return the style set on the component")
	}
}

func TestText_RespectsWidth(t *testing.T) {
	txt := NewText("t", "Hi", lipgloss.NewStyle())
	txt.SetSize(20, 1)
	w := lipgloss.Width(txt.View())
	if w != 20 {
		t.Errorf("width = %d, want 20", w)
	}
}
