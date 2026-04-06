package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Compile-time interface check
var _ Leaf = (*TextInput)(nil)

func TestTextInput_Defaults(t *testing.T) {
	ti := NewTextInput("ti", 20)
	if ti.ID() != "ti" {
		t.Errorf("ID = %q", ti.ID())
	}
	if !ti.Visible() {
		t.Error("should be visible")
	}
	if !ti.Enabled() {
		t.Error("should be enabled")
	}
	if !ti.Focusable() {
		t.Error("should be focusable")
	}
	if ti.Value() != "" {
		t.Errorf("initial value = %q, want empty", ti.Value())
	}
}

func TestTextInput_Value_SetGet(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetValue("hello")
	if ti.Value() != "hello" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "hello")
	}
}

func TestTextInput_Placeholder(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.WithPlaceholder("type here...")
	output := stripansi.Strip(ti.View())
	if !strings.Contains(output, "type here...") {
		t.Errorf("placeholder not visible in view: %q", output)
	}
}

func TestTextInput_View_JustInput(t *testing.T) {
	// TextInput renders only the input area — no label, no separator
	ti := NewTextInput("ti", 20)
	ti.SetValue("hello")
	output := stripansi.Strip(ti.View())
	if !strings.Contains(output, "hello") {
		t.Errorf("view should contain value, got %q", output)
	}
	// Should NOT contain any label/colon
	if strings.Contains(output, ": ") {
		t.Errorf("view should not contain label separator, got %q", output)
	}
}

func TestTextInput_View_WidthStable(t *testing.T) {
	ti := NewTextInput("ti", 20)

	ti.SetFocused(false)
	unfocusedW := lipgloss.Width(ti.View())

	ti.SetFocused(true)
	focusedW := lipgloss.Width(ti.View())

	if unfocusedW != focusedW {
		t.Errorf("width changed: unfocused=%d focused=%d", unfocusedW, focusedW)
	}
}

func TestTextInput_Update_Enter_Submits(t *testing.T) {
	var submitted string
	ti := NewTextInput("ti", 20)
	ti.SetFocused(true)
	ti.OnSubmit(func(v string) tea.Cmd { submitted = v; return nil })
	ti.SetValue("test")

	_, consumed := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed")
	}
	if submitted != "test" {
		t.Errorf("submitted = %q, want %q", submitted, "test")
	}
}

func TestTextInput_Update_Space_IsRegularChar(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetFocused(true)

	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	ti.Update(tea.KeyMsg{Type: tea.KeySpace})
	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if ti.Value() != "a b" {
		t.Errorf("value = %q, want 'a b'", ti.Value())
	}
}

func TestTextInput_Update_Typing(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetFocused(true)

	_, consumed := ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !consumed {
		t.Error("rune key should be consumed")
	}
	if ti.Value() != "a" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "a")
	}
}

func TestTextInput_Update_Inactive_Ignored(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetEnabled(false)

	_, consumed := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("inactive textinput should not consume keys")
	}
}

func TestTextInput_KeyBindings_NilByDefault(t *testing.T) {
	ti := NewTextInput("ti", 20)
	if ti.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(ti.KeyBindings()))
	}
}

func TestTextInput_FocusBlur(t *testing.T) {
	ti := NewTextInput("ti", 20)

	ti.SetFocused(true)
	if !ti.input.Focused() {
		t.Error("inner textinput should be focused")
	}

	ti.SetFocused(false)
	if ti.input.Focused() {
		t.Error("inner textinput should be blurred")
	}
}

func TestTextInput_View_GrowsWithWidth(t *testing.T) {
	ti := NewTextInput("ti", 10)
	ti.SetSize(40, 1)
	w := lipgloss.Width(ti.View())
	if w != 40 {
		t.Errorf("width = %d, want 40", w)
	}
}

func TestTextInput_View_FocusedDiffers(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	ti := NewTextInput("ti", 20)
	ti.WithPlaceholder("placeholder")
	normalOutput := ti.View()

	ti.SetFocused(true)
	focusedOutput := ti.View()

	// Focused shows cursor, unfocused shows placeholder differently
	if normalOutput == focusedOutput {
		t.Error("focused and unfocused views should differ")
	}
}

func TestTextInput_HandleEvent_Click_NoAction(t *testing.T) {
	ti := NewTextInput("ti", 20)
	_, activated := ti.HandleEvent(MouseClickEvent{})
	if activated {
		t.Error("clicking text input should not activate, just focus")
	}
}

func TestTextInput_WithCharLimit(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.WithCharLimit(5)
	ti.SetFocused(true)

	// Type more than the limit
	for _, r := range "abcdefghij" {
		ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if len(ti.Value()) > 5 {
		t.Errorf("Value() length = %d, want <= 5 (char limit)", len(ti.Value()))
	}
}

func TestTextInput_OnChange(t *testing.T) {
	var received string
	ti := NewTextInput("ti", 20)
	ti.SetFocused(true)
	ti.OnChange(func(v string) tea.Cmd { received = v; return nil })

	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if received != "a" {
		t.Errorf("onChange received = %q, want %q", received, "a")
	}

	ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if received != "ab" {
		t.Errorf("onChange received = %q, want %q", received, "ab")
	}
}

func TestTextInput_OnChange_NoChangeNoCallback(t *testing.T) {
	called := false
	ti := NewTextInput("ti", 20)
	ti.SetFocused(true)
	ti.OnChange(func(v string) tea.Cmd { called = true; return nil })

	// Navigation keys don't change value — onChange should not be called
	ti.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if called {
		t.Error("onChange should not be called when value doesn't change")
	}
}

func TestTextInput_Update_Inactive(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetEnabled(false)

	_, consumed := ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if consumed {
		t.Error("inactive textinput should not consume keys")
	}
}

func TestTextInput_View_Inactive(t *testing.T) {
	ti := NewTextInput("ti", 20)
	ti.SetValue("test")
	ti.SetEnabled(false)
	output := stripansi.Strip(ti.View())
	if !strings.Contains(output, "test") {
		t.Errorf("inactive view should show value, got %q", output)
	}
}
