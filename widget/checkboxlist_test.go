package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
)

// Compile-time interface check
var _ Leaf = (*CheckboxList)(nil)

func testItems() []CheckboxItem {
	return []CheckboxItem{
		{Label: "Alpha", Value: "a"},
		{Label: "Beta", Value: "b"},
		{Label: "Gamma", Value: "c"},
	}
}

func TestCheckboxList_Defaults(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	if cl.ID() != "cl" {
		t.Errorf("ID = %q", cl.ID())
	}
	if !cl.Visible() {
		t.Error("should be visible")
	}
	if !cl.Enabled() {
		t.Error("should be enabled")
	}
}

func TestCheckboxList_Interface(t *testing.T) {
	var _ Leaf = (*CheckboxList)(nil)
}

func TestCheckboxList_Navigate_Down(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !consumed {
		t.Error("down should be consumed")
	}
	if cl.cursor != 1 {
		t.Errorf("cursor = %d, want 1", cl.cursor)
	}
}

func TestCheckboxList_Navigate_Up(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	cl.cursor = 2
	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeyUp})
	if !consumed {
		t.Error("up should be consumed")
	}
	if cl.cursor != 1 {
		t.Errorf("cursor = %d, want 1", cl.cursor)
	}
}

func TestCheckboxList_Navigate_Bounds(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)

	// Already at 0, up should not go negative
	cl.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cl.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", cl.cursor)
	}

	// Move to last item, down should not exceed
	cl.cursor = len(cl.items) - 1
	cl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cl.cursor != len(cl.items)-1 {
		t.Errorf("cursor should stay at %d, got %d", len(cl.items)-1, cl.cursor)
	}
}

func TestCheckboxList_Toggle_Space(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !consumed {
		t.Error("space should be consumed")
	}
	if !cl.items[0].Checked {
		t.Error("item 0 should be checked after toggle")
	}

	// Toggle again to uncheck
	cl.Update(tea.KeyMsg{Type: tea.KeySpace})
	if cl.items[0].Checked {
		t.Error("item 0 should be unchecked after second toggle")
	}
}

func TestCheckboxList_Enter_NotConsumed(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("enter should NOT be consumed")
	}
}

func TestCheckboxList_Selected(t *testing.T) {
	items := testItems()
	items[0].Checked = true
	items[2].Checked = true
	cl := NewCheckboxList("cl", items, DefaultCheckboxListStyles())
	sel := cl.Selected()
	if len(sel) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(sel))
	}
	if sel[0] != "a" || sel[1] != "c" {
		t.Errorf("selected = %v, want [a c]", sel)
	}
}

func TestCheckboxList_View_Cursor(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	cl.cursor = 0
	output := stripansi.Strip(cl.View())
	lines := strings.Split(output, "\n")
	if !strings.HasPrefix(lines[0], "> ") {
		t.Errorf("cursor line should start with '> ', got %q", lines[0])
	}
	if strings.HasPrefix(lines[1], "> ") {
		t.Errorf("non-cursor line should not start with '> ', got %q", lines[1])
	}
}

func TestCheckboxList_View_Checked(t *testing.T) {
	items := testItems()
	items[0].Checked = true
	cl := NewCheckboxList("cl", items, DefaultCheckboxListStyles())
	cl.SetFocused(true)
	output := stripansi.Strip(cl.View())
	lines := strings.Split(output, "\n")
	if !strings.Contains(lines[0], "[x]") {
		t.Errorf("checked item should contain '[x]', got %q", lines[0])
	}
}

func TestCheckboxList_View_Unchecked(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	output := stripansi.Strip(cl.View())
	lines := strings.Split(output, "\n")
	if !strings.Contains(lines[0], "[ ]") {
		t.Errorf("unchecked item should contain '[ ]', got %q", lines[0])
	}
}

func TestCheckboxList_View_Group(t *testing.T) {
	items := []CheckboxItem{
		{Label: "Group1", Value: "g1", IsGroup: true},
		{Label: "Item1", Value: "i1"},
	}
	cl := NewCheckboxList("cl", items, DefaultCheckboxListStyles())
	cl.SetFocused(true)
	// Just verify it renders without panic and contains the group label
	output := stripansi.Strip(cl.View())
	if !strings.Contains(output, "Group1") {
		t.Errorf("should contain group label, got %q", output)
	}
}

func TestCheckboxList_Callback(t *testing.T) {
	var called bool
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	cl.OnChange(func(items []CheckboxItem) tea.Cmd { called = true; return nil })
	cl.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !called {
		t.Error("onChange should be called on toggle")
	}
}

func TestCheckboxList_Inactive_NoCallback(t *testing.T) {
	var called bool
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.OnChange(func(items []CheckboxItem) tea.Cmd { called = true; return nil })
	cl.SetEnabled(false)
	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeySpace})
	if consumed {
		t.Error("inactive checkbox should not consume keys")
	}
	if called {
		t.Error("inactive checkbox should not fire callback")
	}
}

func TestCheckboxList_KeyBindings_NilByDefault(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	if cl.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(cl.KeyBindings()))
	}
}

func TestCheckboxList_Items(t *testing.T) {
	items := testItems()
	cl := NewCheckboxList("cl", items, DefaultCheckboxListStyles())
	got := cl.Items()
	if len(got) != 3 {
		t.Fatalf("Items() len = %d, want 3", len(got))
	}
	if got[0].Label != "Alpha" {
		t.Errorf("Items()[0].Label = %q, want %q", got[0].Label, "Alpha")
	}
	if got[2].Value != "c" {
		t.Errorf("Items()[2].Value = %q, want %q", got[2].Value, "c")
	}
}

func TestCheckboxList_Cursor(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	if cl.Cursor() != 0 {
		t.Errorf("Cursor() = %d, want 0", cl.Cursor())
	}
	cl.SetFocused(true)
	cl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cl.Cursor() != 1 {
		t.Errorf("Cursor() = %d after down, want 1", cl.Cursor())
	}
}

func TestCheckboxList_BindDefaultActionToKey(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	cl.BindDefaultActionToKey("ctrl+t", "Toggle")

	bindings := cl.KeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+t" {
		t.Errorf("key = %q, want %q", bindings[0].Help().Key, "ctrl+t")
	}

	// Resolve and fire the binding — should toggle item at cursor (0)
	action, found := cl.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlT})
	if !found {
		t.Fatal("ctrl+t binding should be registered")
	}
	action()
	if !cl.items[0].Checked {
		t.Error("binding action should toggle item at cursor")
	}
}

func TestCheckboxList_BindDefaultActionToKey_WithOnChange(t *testing.T) {
	called := false
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.OnChange(func(items []CheckboxItem) tea.Cmd { called = true; return nil })
	cl.BindDefaultActionToKey("ctrl+t")

	action, _ := cl.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlT})
	action()
	if !called {
		t.Error("binding should trigger onChange callback")
	}
}

func TestCheckboxList_HandleEvent_MouseClick(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetPosition(0, 0)

	// Click on row 1 (second item)
	_, consumed := cl.HandleEvent(MouseClickEvent{X: 5, Y: 1})
	if !consumed {
		t.Error("click should be consumed")
	}
	if cl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 after click", cl.Cursor())
	}
	if !cl.items[1].Checked {
		t.Error("clicked item should be toggled")
	}
}

func TestCheckboxList_HandleEvent_MouseClick_WithOnChange(t *testing.T) {
	called := false
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetPosition(0, 0)
	cl.OnChange(func(items []CheckboxItem) tea.Cmd { called = true; return nil })

	cl.HandleEvent(MouseClickEvent{X: 5, Y: 0})
	if !called {
		t.Error("click should trigger onChange")
	}
}

func TestCheckboxList_HandleEvent_MouseClick_OutOfRange(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetPosition(0, 0)

	_, consumed := cl.HandleEvent(MouseClickEvent{X: 5, Y: 10})
	if consumed {
		t.Error("click out of range should not be consumed")
	}
}

func TestCheckboxList_HandleEvent_MouseClick_Inactive(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetPosition(0, 0)
	cl.SetEnabled(false)

	_, consumed := cl.HandleEvent(MouseClickEvent{X: 5, Y: 0})
	if consumed {
		t.Error("click on inactive checkboxlist should not be consumed")
	}
}

func TestCheckboxList_Update_OnKeyPress_Fallthrough(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	called := false
	cl.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "enter" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := cl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed via onKeyPress")
	}
	if !called {
		t.Error("onKeyPress should be called")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func TestCheckboxList_Update_OnKeyPress_ReturnsNil_NotConsumed(t *testing.T) {
	cl := NewCheckboxList("cl", testItems(), DefaultCheckboxListStyles())
	cl.SetFocused(true)
	cl.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		return nil
	})

	_, consumed := cl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if consumed {
		t.Error("unhandled key with nil-returning onKeyPress should not be consumed")
	}
}

func TestCheckboxList_View_GroupItem(t *testing.T) {
	items := []CheckboxItem{
		{Label: "GroupHeader", Value: "g", IsGroup: true},
		{Label: "Child1", Value: "c1"},
	}
	cl := NewCheckboxList("cl", items, DefaultCheckboxListStyles())
	cl.SetFocused(true)
	output := stripansi.Strip(cl.View())
	if !strings.Contains(output, "GroupHeader") {
		t.Errorf("should contain group label, got %q", output)
	}
	if !strings.Contains(output, "Child1") {
		t.Errorf("should contain child label, got %q", output)
	}
}
