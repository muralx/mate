package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Compile-time interface check.
var _ Leaf = (*Toggle)(nil)

func TestToggle_Defaults(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	if tg.ID() != "toggle1" {
		t.Errorf("ID = %q", tg.ID())
	}
	if !tg.Visible() {
		t.Error("should be visible")
	}
	if !tg.Enabled() {
		t.Error("should be enabled")
	}
	if !tg.On() {
		t.Error("should be on")
	}
}

func TestToggle_Interface(t *testing.T) {
	var _ Leaf = (*Toggle)(nil)
}

func TestToggle_Toggle_Space(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	_, consumed := tg.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !consumed {
		t.Error("space should be consumed")
	}
	if tg.On() {
		t.Error("should be off after toggle")
	}
}

func TestToggle_Toggle_Enter(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	_, consumed := tg.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed")
	}
	if !tg.On() {
		t.Error("should be on after toggle")
	}
}

func TestToggle_Callback(t *testing.T) {
	var got bool
	called := false
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.OnChange(func(v bool) tea.Cmd { called = true; got = v; return nil })
	tg.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !called {
		t.Error("onChange should be called")
	}
	if got {
		t.Error("onChange should receive false")
	}
}

func TestToggle_View_OnOff_On_Unfocused(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "Feature:") {
		t.Errorf("should contain label, got %q", output)
	}
	if !strings.Contains(output, "[on]") {
		t.Errorf("should contain [on], got %q", output)
	}
}

func TestToggle_View_OnOff_Off_Unfocused(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "Feature:") {
		t.Errorf("should contain label, got %q", output)
	}
	if !strings.Contains(output, "[off]") {
		t.Errorf("should contain [off], got %q", output)
	}
}

func TestToggle_View_OnOff_Focused(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetFocused(true)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "Feature:") {
		t.Errorf("should contain label, got %q", output)
	}
	if !strings.Contains(output, "[on]") {
		t.Errorf("should contain [on], got %q", output)
	}
}

func TestToggle_View_Radio_On_Unfocused(t *testing.T) {
	tg := NewToggle("src", "", true, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[Live]") {
		t.Errorf("should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("should contain [Cache], got %q", output)
	}
}

func TestToggle_View_Radio_Off_Unfocused(t *testing.T) {
	tg := NewToggle("src", "", false, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[Live]") {
		t.Errorf("should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("should contain [Cache], got %q", output)
	}
}

func TestToggle_View_Radio_Focused(t *testing.T) {
	tg := NewToggle("src", "", true, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	tg.SetFocused(true)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[Live]") {
		t.Errorf("should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("should contain [Cache], got %q", output)
	}
}

func TestToggle_View_WidthStable_OnOff(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())

	onW := lipgloss.Width(tg.View())
	tg.Update(tea.KeyMsg{Type: tea.KeySpace}) // toggle off
	offW := lipgloss.Width(tg.View())

	if onW != offW {
		t.Errorf("width changed: on=%d off=%d", onW, offW)
	}
}

func TestToggle_View_WidthStable_Radio(t *testing.T) {
	tg := NewToggle("src", "", true, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")

	onW := lipgloss.Width(tg.View())
	tg.Update(tea.KeyMsg{Type: tea.KeySpace}) // toggle off
	offW := lipgloss.Width(tg.View())

	if onW != offW {
		t.Errorf("width changed: on=%d off=%d", onW, offW)
	}
}

func TestToggle_View_Inactive(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetEnabled(false)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "Feature:") {
		t.Errorf("inactive should still show label, got %q", output)
	}
}

func TestToggle_Inactive_NoCallback(t *testing.T) {
	called := false
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.OnChange(func(bool) tea.Cmd { called = true; return nil })
	tg.SetEnabled(false)
	_, consumed := tg.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("inactive toggle should not consume keys")
	}
	if called {
		t.Error("inactive toggle should not fire callback")
	}
}

func TestToggle_KeyBindings_NilByDefault(t *testing.T) {
	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	if tg.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(tg.KeyBindings()))
	}
}

func TestToggle_View_StyleApplied(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	tg := NewToggle("toggle1", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	normalOutput := tg.View()

	tg.SetFocused(true)
	focusedOutput := tg.View()

	if normalOutput == focusedOutput {
		t.Error("Normal and Focused views should differ with TrueColor")
	}
}

func TestToggle_HandleEvent_Click_Toggles(t *testing.T) {
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	_, activated := tg.HandleEvent(MouseClickEvent{})
	if !activated {
		t.Error("click should activate toggle")
	}
	if !tg.On() {
		t.Error("click should toggle to on")
	}
}

func TestToggle_HandleEvent_Click_Inactive(t *testing.T) {
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetEnabled(false)
	_, activated := tg.HandleEvent(MouseClickEvent{})
	if activated {
		t.Error("inactive toggle click should not activate")
	}
}

func TestToggle_SetOn(t *testing.T) {
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	if tg.On() {
		t.Error("initial state should be off")
	}
	tg.SetOn(true)
	if !tg.On() {
		t.Error("SetOn(true) should set state to on")
	}
	tg.SetOn(false)
	if tg.On() {
		t.Error("SetOn(false) should set state to off")
	}
}

func TestToggle_BindDefaultActionToKey(t *testing.T) {
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	tg.BindDefaultActionToKey("ctrl+t", "Toggle feature")

	bindings := tg.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+t" {
		t.Errorf("key = %q, want %q", bindings[0].Help().Key, "ctrl+t")
	}

	action, found := tg.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlT})
	if !found {
		t.Fatal("ctrl+t binding should be registered")
	}
	action()
	if !tg.On() {
		t.Error("binding action should toggle state to on")
	}
}

func TestToggle_BindDefaultActionToKey_WithOnChange(t *testing.T) {
	called := false
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	tg.OnChange(func(v bool) tea.Cmd { called = true; return nil })
	tg.BindDefaultActionToKey("ctrl+t")

	action, _ := tg.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlT})
	action()
	if !called {
		t.Error("binding should trigger onChange callback")
	}
}

func TestToggle_Update_OnKeyPress_Fallthrough(t *testing.T) {
	tg := NewToggle("t", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetFocused(true)
	called := false
	tg.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "x" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
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

func TestToggle_Update_OnKeyPress_ReturnsNil(t *testing.T) {
	tg := NewToggle("t", "Feature", true, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetFocused(true)
	tg.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd { return nil })

	_, consumed := tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if consumed {
		t.Error("nil-returning onKeyPress should not consume")
	}
}

func TestToggle_HandleEvent_Click_WithOnChange(t *testing.T) {
	var received bool
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	tg.OnChange(func(v bool) tea.Cmd { received = v; return nil })

	_, consumed := tg.HandleEvent(MouseClickEvent{})
	if !consumed {
		t.Error("click should be consumed")
	}
	if !received {
		t.Error("onChange should receive true (toggled from off to on)")
	}
}

func TestToggle_ViewRadio_OffFocused(t *testing.T) {
	tg := NewToggle("src", "Mode", false, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	tg.SetFocused(true)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[Live]") {
		t.Errorf("should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("should contain [Cache], got %q", output)
	}
}

func TestToggle_ViewRadio_OffUnfocused(t *testing.T) {
	tg := NewToggle("src", "Mode", false, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[Live]") {
		t.Errorf("should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("should contain [Cache], got %q", output)
	}
}

func TestToggle_ViewInactive_Radio(t *testing.T) {
	tg := NewToggle("src", "Mode", true, ToggleModeRadio, DefaultToggleStyles())
	tg.SetLabels("[Live]", "[Cache]")
	tg.SetEnabled(false)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "Mode:") {
		t.Errorf("inactive radio should show label, got %q", output)
	}
	if !strings.Contains(output, "[Live]") {
		t.Errorf("inactive radio should contain [Live], got %q", output)
	}
	if !strings.Contains(output, "[Cache]") {
		t.Errorf("inactive radio should contain [Cache], got %q", output)
	}
}

func TestToggle_View_OnOff_OffFocused(t *testing.T) {
	tg := NewToggle("t", "Feature", false, ToggleModeOnOff, DefaultToggleStyles())
	tg.SetFocused(true)
	output := stripansi.Strip(tg.View())
	if !strings.Contains(output, "[off]") {
		t.Errorf("should contain [off] when off+focused, got %q", output)
	}
}
