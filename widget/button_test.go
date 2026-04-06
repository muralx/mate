package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Compile-time interface checks
var _ Leaf = (*Button)(nil)
var _ Component = (*Button)(nil)

func TestButton_Defaults(t *testing.T) {
	b := NewButton("btn", "Search", DefaultButtonStyles())
	if b.ID() != "btn" {
		t.Errorf("ID = %q", b.ID())
	}
	if !b.Visible() {
		t.Error("should be visible")
	}
	if !b.Enabled() {
		t.Error("should be enabled")
	}
}

func TestButton_Focusable(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	if !b.Focusable() {
		t.Error("button should be focusable")
	}
}

func TestButton_Press_Space(t *testing.T) {
	pressed := false
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnPress(func() tea.Cmd { pressed = true; return nil })
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !consumed {
		t.Error("space should be consumed")
	}
	if !pressed {
		t.Error("onPress should be called")
	}
}

func TestButton_Press_Enter(t *testing.T) {
	pressed := false
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnPress(func() tea.Cmd { pressed = true; return nil })
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed")
	}
	if !pressed {
		t.Error("onPress should be called")
	}
}

func TestButton_UnconsumedKeys(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if consumed {
		t.Error("'a' should not be consumed")
	}
	_, consumed = b.Update(tea.KeyMsg{Type: tea.KeyTab})
	if consumed {
		t.Error("tab should not be consumed")
	}
}

func TestButton_View_Unfocused(t *testing.T) {
	b := NewButton("btn", "Search", DefaultButtonStyles())
	output := stripansi.Strip(b.View())
	if !strings.Contains(output, "[ Search ]") {
		t.Errorf("unfocused should render '[ Search ]', got %q", output)
	}
}

func TestButton_View_Focused(t *testing.T) {
	b := NewButton("btn", "Search", DefaultButtonStyles())
	b.SetFocused(true)
	output := stripansi.Strip(b.View())
	if !strings.Contains(output, "[ Search ]") {
		t.Errorf("focused should render '[ Search ]', got %q", output)
	}
}

func TestButton_View_WidthStable(t *testing.T) {
	b := NewButton("btn", "Search", DefaultButtonStyles())

	b.SetFocused(false)
	unfocusedW := lipgloss.Width(b.View())

	b.SetFocused(true)
	focusedW := lipgloss.Width(b.View())

	if unfocusedW != focusedW {
		t.Errorf("width changed: unfocused=%d focused=%d", unfocusedW, focusedW)
	}
}

func TestButton_View_Inactive(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.SetEnabled(false)
	output := stripansi.Strip(b.View())
	if !strings.Contains(output, "[ OK ]") {
		t.Errorf("inactive should still show label, got %q", output)
	}
}

func TestButton_Inactive_NoCallback(t *testing.T) {
	pressed := false
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnPress(func() tea.Cmd { pressed = true; return nil })
	b.SetEnabled(false)
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("inactive button should not consume keys")
	}
	if pressed {
		t.Error("inactive button should not fire callback")
	}
}

func TestButton_KeyBindings_NilByDefault(t *testing.T) {
	// Button should not expose local keys (space/enter) via KeyBindings().
	// Those are handled internally in Update() and are not global shortcuts.
	b := NewButton("btn", "Search", DefaultButtonStyles())
	if b.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(b.KeyBindings()))
	}
}

func TestButton_KeyBindings_WithRegistered(t *testing.T) {
	// Only bindings from RegisterKeyBinding/BindDefaultActionToKey should appear.
	b := NewButton("btn", "Save", DefaultButtonStyles())
	b.BindDefaultActionToKey("ctrl+s", "Save")

	bindings := b.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+s" {
		t.Errorf("key = %q, want ctrl+s", bindings[0].Help().Key)
	}
}

func TestButton_View_StyleApplied(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	b := NewButton("btn", "OK", DefaultButtonStyles())
	normalOutput := b.View()

	b.SetFocused(true)
	focusedOutput := b.View()

	if normalOutput == focusedOutput {
		t.Error("Normal and Focused views should differ with TrueColor")
	}
}

func TestButton_NoCallback_NoPanic(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	// No OnPress set — pressing should not panic
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should still be consumed even without callback")
	}
}

func TestButton_View_RespectsWidth(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.SetSize(30, 1)
	w := lipgloss.Width(b.View())
	if w != 30 {
		t.Errorf("width = %d, want 30", w)
	}
}

func TestButton_View_CenteredInWidth(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.SetSize(30, 1)
	b.SetAlignment(AlignCenter)
	output := stripansi.Strip(b.View())
	// "[ OK ]" is 6 chars, centered in 30 should have leading spaces
	if output[0] != ' ' {
		t.Error("centered button should have leading space")
	}
	w := lipgloss.Width(b.View())
	if w != 30 {
		t.Errorf("width = %d, want 30", w)
	}
}

func TestButton_HandleEvent_Click_Activates(t *testing.T) {
	pressed := false
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnPress(func() tea.Cmd { pressed = true; return nil })
	_, activated := b.HandleEvent(MouseClickEvent{})
	if !activated {
		t.Error("click should activate button")
	}
	if !pressed {
		t.Error("click should fire onPress")
	}
}

func TestButton_HandleEvent_Click_Inactive(t *testing.T) {
	pressed := false
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnPress(func() tea.Cmd { pressed = true; return nil })
	b.SetEnabled(false)
	_, activated := b.HandleEvent(MouseClickEvent{})
	if activated {
		t.Error("inactive button click should not activate")
	}
	if pressed {
		t.Error("inactive button click should not fire onPress")
	}
}

func TestButton_HandleEvent_Click_NoOnPress(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	// No OnPress set
	_, consumed := b.HandleEvent(MouseClickEvent{})
	if !consumed {
		t.Error("click should still be consumed even without onPress")
	}
}

func TestButton_Update_OnKeyPress_Fallthrough(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	called := false
	b.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "x" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := b.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if !consumed {
		t.Error("custom key should be consumed via onKeyPress")
	}
	if !called {
		t.Error("onKeyPress should be called")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func TestButton_Update_OnKeyPress_ReturnsNil(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd { return nil })

	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if consumed {
		t.Error("nil-returning onKeyPress should not consume")
	}
}

func TestButton_BindDefaultActionToKey_NoOnPress(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.BindDefaultActionToKey("ctrl+s")

	action, found := b.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !found {
		t.Fatal("ctrl+s binding should be registered")
	}
	// Should return nil cmd when no onPress set (no panic)
	cmd := action()
	if cmd != nil {
		t.Error("action with no onPress should return nil")
	}
}

func TestButton_BindDefaultActionToKey_WithDescription(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	b.BindDefaultActionToKey("ctrl+s", "Save")

	bindings := b.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1", len(bindings))
	}
	if bindings[0].Help().Desc != "Save" {
		t.Errorf("desc = %q, want %q", bindings[0].Help().Desc, "Save")
	}
}

func TestButton_Update_SpaceNoCallback(t *testing.T) {
	b := NewButton("btn", "OK", DefaultButtonStyles())
	// No OnPress set — space should still be consumed but no panic
	_, consumed := b.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !consumed {
		t.Error("space should be consumed even without callback")
	}
}
