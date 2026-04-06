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
var _ Leaf = (*TabBar)(nil)

func TestTabBar_Defaults(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	if tb.ID() != "tabs" {
		t.Errorf("ID = %q", tb.ID())
	}
	if !tb.Visible() {
		t.Error("should be visible")
	}
	if !tb.Enabled() {
		t.Error("should be enabled")
	}
	if tb.ActiveTab() != 0 {
		t.Errorf("active = %d, want 0", tb.ActiveTab())
	}
	if tb.CursorTab() != 0 {
		t.Errorf("cursor = %d, want 0", tb.CursorTab())
	}
}

func TestTabBar_Interface(t *testing.T) {
	var _ Leaf = (*TabBar)(nil)
}

func TestTabBar_CursorMove_Right(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetFocused(true)
	_, consumed := tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if !consumed {
		t.Error("right should be consumed")
	}
	if tb.CursorTab() != 1 {
		t.Errorf("cursor = %d, want 1", tb.CursorTab())
	}
	// Active should NOT change — just cursor moves
	if tb.ActiveTab() != 0 {
		t.Errorf("active = %d, want 0 (unchanged)", tb.ActiveTab())
	}
}

func TestTabBar_CursorMove_Left(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetActiveTab(2)
	tb.SetFocused(true)
	_, consumed := tb.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if !consumed {
		t.Error("left should be consumed")
	}
	if tb.CursorTab() != 1 {
		t.Errorf("cursor = %d, want 1", tb.CursorTab())
	}
	// Active should NOT change
	if tb.ActiveTab() != 2 {
		t.Errorf("active = %d, want 2 (unchanged)", tb.ActiveTab())
	}
}

func TestTabBar_CursorBounds(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetFocused(true)

	// Already at 0, left should clamp
	tb.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if tb.CursorTab() != 0 {
		t.Errorf("cursor = %d, want 0 (clamped)", tb.CursorTab())
	}

	// Go to last, then try right
	tb.SetActiveTab(2)
	tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tb.CursorTab() != 2 {
		t.Errorf("cursor = %d, want 2 (clamped)", tb.CursorTab())
	}
}

func TestTabBar_SelectWithSpace(t *testing.T) {
	called := -1
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = i; return nil })
	tb.SetFocused(true)

	// Move cursor to B
	tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tb.CursorTab() != 1 {
		t.Fatalf("cursor = %d, want 1", tb.CursorTab())
	}
	if tb.ActiveTab() != 0 {
		t.Fatalf("active = %d, want 0 (not yet selected)", tb.ActiveTab())
	}

	// Press space to select
	tb.Update(tea.KeyMsg{Type: tea.KeySpace})
	if tb.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1 (after space)", tb.ActiveTab())
	}
	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabBar_SelectWithEnter(t *testing.T) {
	called := -1
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = i; return nil })
	tb.SetFocused(true)

	tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	tb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if tb.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1 (after enter)", tb.ActiveTab())
	}
	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabBar_SelectSameTab_NoCallback(t *testing.T) {
	called := false
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = true; return nil })
	tb.SetFocused(true)

	// Cursor is on active tab (0), pressing space should not fire callback
	tb.Update(tea.KeyMsg{Type: tea.KeySpace})
	if called {
		t.Error("selecting already-active tab should not fire callback")
	}
}

func TestTabBar_View_AllTabsPresent(t *testing.T) {
	tb := NewTabBar("tabs", []string{"Alpha", "Beta", "Gamma"}, DefaultTabBarStyles())
	output := stripansi.Strip(tb.View())
	for _, label := range []string{"Alpha", "Beta", "Gamma"} {
		if !strings.Contains(output, label) {
			t.Errorf("view should contain %q, got %q", label, output)
		}
	}
}

func TestTabBar_View_FocusedCursorHighlight(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())

	// Unfocused view
	unfocused := tb.View()

	// Focus and move cursor to B (inactive tab)
	tb.SetFocused(true)
	tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	focused := tb.View()

	// The views should differ (cursor on B should be highlighted)
	if unfocused == focused {
		t.Error("focused view with cursor on B should differ from unfocused view")
	}
}

func TestTabBar_View_CursorOnActiveTab_Underline(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())

	// Unfocused: active tab without underline
	unfocused := tb.View()

	// Focused: cursor on active tab should get underline
	tb.SetFocused(true)
	focused := tb.View()

	if unfocused == focused {
		t.Error("active tab should look different when focused (underline)")
	}
}

func TestTabBar_Inactive_NoCallback(t *testing.T) {
	called := false
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = true; return nil })
	tb.SetEnabled(false)

	_, consumed := tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if consumed {
		t.Error("inactive tab bar should not consume keys")
	}
	if called {
		t.Error("inactive tab bar should not fire callback")
	}
}

func TestTabBar_KeyBindings_NilByDefault(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	if tb.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(tb.KeyBindings()))
	}
}

func TestTabBar_SetActive_MovesCursor(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetActiveTab(2)
	if tb.ActiveTab() != 2 {
		t.Errorf("active = %d, want 2", tb.ActiveTab())
	}
	if tb.CursorTab() != 2 {
		t.Errorf("cursor = %d, want 2 (should follow active)", tb.CursorTab())
	}
}

func TestTabBar_FocusResetsCursor(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetActiveTab(1)

	// Focus — cursor should reset to active tab
	tb.SetFocused(true)
	if tb.CursorTab() != 1 {
		t.Errorf("cursor = %d, want 1 (reset to active on focus)", tb.CursorTab())
	}

	// Move cursor
	tb.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tb.CursorTab() != 2 {
		t.Errorf("cursor = %d, want 2", tb.CursorTab())
	}

	// Blur and refocus — cursor should reset
	tb.SetFocused(false)
	tb.SetFocused(true)
	if tb.CursorTab() != 1 {
		t.Errorf("cursor = %d, want 1 (reset on refocus)", tb.CursorTab())
	}
}

func TestTabBar_SetTabKeyBinding_ActivatesTab(t *testing.T) {
	called := -1
	tb := NewTabBar("tabs", []string{"Overview", "Details", "Settings"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = i; return nil })

	tb.SetTabKeyBinding(1, "ctrl+e")

	// Resolve the binding
	action, found := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !found {
		t.Fatal("ctrl+e binding should be registered")
	}
	action()

	if tb.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1", tb.ActiveTab())
	}
	if tb.CursorTab() != 1 {
		t.Errorf("cursor = %d, want 1 (should follow active)", tb.CursorTab())
	}
	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabBar_SetTabKeyBinding_DefaultDescription(t *testing.T) {
	tb := NewTabBar("tabs", []string{"Overview", "Details"}, DefaultTabBarStyles())
	tb.SetTabKeyBinding(0, "ctrl+d")

	bindings := tb.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+d" {
		t.Errorf("key = %q, want ctrl+d", bindings[0].Help().Key)
	}
	if bindings[0].Help().Desc != "Overview" {
		t.Errorf("desc = %q, want Overview (derived from tab label)", bindings[0].Help().Desc)
	}
}

func TestTabBar_SetTabKeyBinding_CustomDescription(t *testing.T) {
	tb := NewTabBar("tabs", []string{"Overview", "Details"}, DefaultTabBarStyles())
	tb.SetTabKeyBinding(0, "ctrl+d", "Dash")

	bindings := tb.KeyBindings()
	if bindings[0].Help().Desc != "Dash" {
		t.Errorf("desc = %q, want Dash (custom override)", bindings[0].Help().Desc)
	}
}

func TestTabBar_SetTabKeyBinding_AlreadyActiveTab_NoCallback(t *testing.T) {
	called := false
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = true; return nil })

	tb.SetTabKeyBinding(0, "ctrl+a")

	// Tab 0 is already active — firing should be a no-op
	action, found := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !found {
		t.Fatal("binding should be registered")
	}
	action()

	if called {
		t.Error("should not fire onChange when tab is already active")
	}
}

func TestTabBar_SetTabKeyBinding_PanicsOnInvalidIndex(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-range index")
		}
	}()
	tb.SetTabKeyBinding(5, "ctrl+x")
}

func TestTabBar_SetTabKeyBinding_PanicsOnNegativeIndex(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative index")
		}
	}()
	tb.SetTabKeyBinding(-1, "ctrl+x")
}

func TestTabBar_SetTabKeyBinding_ReplaceExisting(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())

	tb.SetTabKeyBinding(0, "ctrl+a")
	tb.SetTabKeyBinding(0, "ctrl+x")

	// Old binding should be gone
	_, found := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Error("old binding ctrl+a should be removed")
	}

	// New binding should work
	_, found = tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlX})
	if !found {
		t.Error("new binding ctrl+x should be registered")
	}

	// Should have exactly 1 binding (not 2)
	if len(tb.KeyBindings()) != 1 {
		t.Errorf("bindings = %d, want 1 (replaced, not accumulated)", len(tb.KeyBindings()))
	}
}

func TestTabBar_SetTabKeyBinding_MultipleTabsRegistered(t *testing.T) {
	tb := NewTabBar("tabs", []string{"Overview", "Details", "Settings"}, DefaultTabBarStyles())

	tb.SetTabKeyBinding(0, "ctrl+d")
	tb.SetTabKeyBinding(1, "ctrl+e")
	tb.SetTabKeyBinding(2, "ctrl+g")

	if len(tb.KeyBindings()) != 3 {
		t.Fatalf("bindings = %d, want 3", len(tb.KeyBindings()))
	}

	// Activate tab 2 via binding
	action, _ := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlG})
	action()
	if tb.ActiveTab() != 2 {
		t.Errorf("active = %d, want 2", tb.ActiveTab())
	}
}

func TestTabBar_HandleEvent_ClickActivatesTab(t *testing.T) {
	called := -1
	tb := NewTabBar("tabs", []string{"AA", "BB", "CC"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = i; return nil })
	tb.SetPosition(0, 0)

	// Render to get tab widths
	tb.View()
	w, _ := tb.Size()
	if w == 0 {
		// View doesn't set size on self, but we need rendered widths.
		// The tab positions are determined by rendered widths of each label+padding.
		// With Padding(0,2), "AA" renders as "  AA  " = 6 chars. Same for BB, CC.
		// Tab 0: columns 0-5, Tab 1: columns 6-11, Tab 2: columns 12-17
	}

	// Click in the middle of the second tab (BB)
	// Each tab with padding(0,2): 2+2+2 = 6 chars wide
	_, consumed := tb.HandleEvent(MouseClickEvent{X: 7, Y: 0})
	if !consumed {
		t.Error("click should be consumed")
	}
	if tb.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1 (clicked second tab)", tb.ActiveTab())
	}
	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabBar_HandleEvent_ClickAlreadyActive_NoCallback(t *testing.T) {
	called := false
	tb := NewTabBar("tabs", []string{"AA", "BB"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = true; return nil })
	tb.SetPosition(0, 0)

	// Click on first tab (already active)
	_, consumed := tb.HandleEvent(MouseClickEvent{X: 1, Y: 0})
	if !consumed {
		t.Error("click should be consumed even on active tab")
	}
	if called {
		t.Error("should not fire onChange when clicking already-active tab")
	}
}

func TestTabBar_HandleEvent_ClickInactive_Ignored(t *testing.T) {
	tb := NewTabBar("tabs", []string{"AA", "BB"}, DefaultTabBarStyles())
	tb.SetPosition(0, 0)
	tb.SetEnabled(false)

	_, consumed := tb.HandleEvent(MouseClickEvent{X: 1, Y: 0})
	if consumed {
		t.Error("click on inactive tabbar should not be consumed")
	}
}

func TestTabBar_HandleEvent_ClickSetsActiveCursor(t *testing.T) {
	tb := NewTabBar("tabs", []string{"AA", "BB", "CC"}, DefaultTabBarStyles())
	tb.SetPosition(0, 0)

	// Click on third tab
	_, _ = tb.HandleEvent(MouseClickEvent{X: 13, Y: 0})
	if tb.ActiveTab() != 2 {
		t.Errorf("active = %d, want 2", tb.ActiveTab())
	}
	if tb.CursorTab() != 2 {
		t.Errorf("cursor = %d, want 2 (should follow active)", tb.CursorTab())
	}
}

func TestTabBar_BindDefaultActionToKey(t *testing.T) {
	called := -1
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = i; return nil })
	tb.SetFocused(true)
	tb.BindDefaultActionToKey("ctrl+a", "Activate tab")

	bindings := tb.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1", len(bindings))
	}

	// Move cursor to B first
	tb.Update(tea.KeyMsg{Type: tea.KeyRight})

	action, found := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !found {
		t.Fatal("ctrl+a binding should be registered")
	}
	action()
	if tb.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1", tb.ActiveTab())
	}
	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabBar_BindDefaultActionToKey_AlreadyActive(t *testing.T) {
	called := false
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	tb.OnChange(func(i int) tea.Cmd { called = true; return nil })
	tb.BindDefaultActionToKey("ctrl+a")

	// Cursor and active are both at 0 — should be no-op
	action, _ := tb.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	action()
	if called {
		t.Error("should not fire onChange when cursor == active")
	}
}

func TestTabBar_Update_OnKeyPress_Fallthrough(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B"}, DefaultTabBarStyles())
	tb.SetFocused(true)
	called := false
	tb.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "x" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := tb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
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

func TestTabBar_HandleEvent_ClickOutsideTabs(t *testing.T) {
	tb := NewTabBar("tabs", []string{"AA"}, DefaultTabBarStyles())
	tb.SetPosition(0, 0)

	// Click far beyond any tab
	_, consumed := tb.HandleEvent(MouseClickEvent{X: 100, Y: 0})
	if consumed {
		t.Error("click outside tabs should not be consumed")
	}
}

func TestTabBar_VimKeys(t *testing.T) {
	tb := NewTabBar("tabs", []string{"A", "B", "C"}, DefaultTabBarStyles())
	tb.SetFocused(true)

	tb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if tb.CursorTab() != 1 {
		t.Errorf("'l' should move cursor right, cursor = %d", tb.CursorTab())
	}

	tb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if tb.CursorTab() != 0 {
		t.Errorf("'h' should move cursor left, cursor = %d", tb.CursorTab())
	}
}
