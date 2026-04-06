package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// --- BaseComponent tests ---

func TestBaseComponent_ID(t *testing.T) {
	bc := NewBaseComponent("test-id")
	if bc.ID() != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", bc.ID())
	}
}

func TestBaseComponent_Size(t *testing.T) {
	bc := NewBaseComponent("c")
	bc.SetSize(80, 24)
	w, h := bc.Size()
	if w != 80 || h != 24 {
		t.Errorf("expected size (80, 24), got (%d, %d)", w, h)
	}
}

func TestBaseComponent_Position(t *testing.T) {
	bc := NewBaseComponent("c")
	bc.SetPosition(10, 20)
	x, y := bc.Position()
	if x != 10 || y != 20 {
		t.Errorf("expected position (10, 20), got (%d, %d)", x, y)
	}
}

func TestBaseComponent_Visible(t *testing.T) {
	bc := NewBaseComponent("c")
	if !bc.Visible() {
		t.Error("expected visible to default to true")
	}
	bc.SetVisible(false)
	if bc.Visible() {
		t.Error("expected visible to be false after SetVisible(false)")
	}
}

func TestBaseComponent_Enabled(t *testing.T) {
	bc := NewBaseComponent("c")
	if !bc.Enabled() {
		t.Error("expected enabled to default to true")
	}
	bc.SetEnabled(false)
	if bc.Enabled() {
		t.Error("expected enabled to be false after SetEnabled(false)")
	}
}

func TestBaseComponent_Active_NoParent(t *testing.T) {
	bc := NewBaseComponent("c")
	if !bc.Active() {
		t.Error("expected Active() true when enabled and no parent")
	}
}

func TestBaseComponent_Active_Disabled(t *testing.T) {
	bc := NewBaseComponent("c")
	bc.SetEnabled(false)
	if bc.Active() {
		t.Error("expected Active() false when disabled")
	}
}

func TestBaseComponent_Focused(t *testing.T) {
	bc := NewBaseComponent("c")
	if bc.Focused() {
		t.Error("expected focused to default to false")
	}
	bc.SetFocused(true)
	if !bc.Focused() {
		t.Error("expected focused to be true after SetFocused(true)")
	}
}

func TestBaseComponent_Focusable_DefaultFalse(t *testing.T) {
	bc := NewBaseComponent("c")
	if bc.Focusable() {
		t.Error("expected Focusable() to default to false")
	}
}

func TestBaseComponent_KeyBindings_DefaultNil(t *testing.T) {
	bc := NewBaseComponent("c")
	if bc.KeyBindings() != nil {
		t.Error("expected KeyBindings() to return nil by default")
	}
}

func TestBaseComponent_View_DefaultEmpty(t *testing.T) {
	bc := NewBaseComponent("c")
	if bc.View() != "" {
		t.Errorf("expected View() to return empty string, got %q", bc.View())
	}
}

func TestBaseComponent_Active_InactiveParent(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	parent.SetEnabled(false)
	if child.Active() {
		t.Error("expected child Active() false when parent is disabled")
	}

	parent.SetEnabled(true)
	if !child.Active() {
		t.Error("expected child Active() true when parent is re-enabled")
	}
}

// --- BaseContainer tests ---

func TestBaseContainer_NewBaseContainer(t *testing.T) {
	bc := NewBaseContainer("container", nil)
	if bc.ID() != "container" {
		t.Errorf("expected ID 'container', got %q", bc.ID())
	}
	if !bc.Visible() {
		t.Error("expected visible to default to true")
	}
	if !bc.Enabled() {
		t.Error("expected enabled to default to true")
	}
	if len(bc.Children()) != 0 {
		t.Error("expected no children initially")
	}
}

func TestBaseContainer_AddChild_SetsParent(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	if child.Parent() != parent {
		t.Error("expected child's parent to be set to the container")
	}
}

func TestBaseContainer_Children(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	c1 := NewBaseComponent("c1")
	c2 := NewBaseComponent("c2")
	parent.AddChild(c1)
	parent.AddChild(c2)

	children := parent.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0].ID() != "c1" || children[1].ID() != "c2" {
		t.Error("children not in expected order")
	}
}

func TestBaseContainer_Focusable_False(t *testing.T) {
	bc := NewBaseContainer("c", nil)
	if bc.Focusable() {
		t.Error("expected container Focusable() to be false")
	}
}

func TestBaseContainer_InnerFocused_NoFocus(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	if parent.InnerFocused() {
		t.Error("expected InnerFocused() false when no child is focused")
	}
}

func TestBaseContainer_InnerFocused_ChildFocused(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)
	child.SetFocused(true)

	if !parent.InnerFocused() {
		t.Error("expected InnerFocused() true when a child is focused")
	}
}

func TestBaseContainer_InnerFocused_NestedChild(t *testing.T) {
	grandparent := NewBaseContainer("grandparent", nil)
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")

	grandparent.AddChild(parent)
	parent.AddChild(child)
	child.SetFocused(true)

	if !parent.InnerFocused() {
		t.Error("expected parent InnerFocused() true when nested child is focused")
	}
	if !grandparent.InnerFocused() {
		t.Error("expected grandparent InnerFocused() true when nested child is focused")
	}
}

func TestBaseContainer_Active_Propagation(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	parent.SetEnabled(false)
	if child.Active() {
		t.Error("expected child Active() false when parent is disabled")
	}
}

func TestBaseContainer_Active_ChildDisabled(t *testing.T) {
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	child.SetEnabled(false)
	if child.Active() {
		t.Error("expected child Active() false when child is disabled")
	}
	if !parent.Active() {
		t.Error("expected parent Active() true when only child is disabled")
	}
}

// --- Alignment and RenderInSize tests ---

func TestBaseComponent_Alignment_Default(t *testing.T) {
	bc := NewBaseComponent("test")
	if bc.Alignment() != AlignLeft {
		t.Error("default should be AlignLeft")
	}
}

func TestBaseComponent_RenderInSize_NoWidth(t *testing.T) {
	bc := NewBaseComponent("test")
	// Width 0 — returns content as-is
	result := bc.RenderInSize("Hello")
	if result != "Hello" {
		t.Errorf("no width should return as-is, got %q", result)
	}
}

func TestBaseComponent_RenderInSize_PadsLeft(t *testing.T) {
	bc := NewBaseComponent("test")
	bc.SetSize(20, 1)
	result := bc.RenderInSize("Hi")
	w := lipgloss.Width(result)
	if w != 20 {
		t.Errorf("width = %d, want 20", w)
	}
}

func TestBaseComponent_RenderInSize_Center(t *testing.T) {
	bc := NewBaseComponent("test")
	bc.SetSize(20, 1)
	bc.SetAlignment(AlignCenter)
	result := bc.RenderInSize("Hi")
	w := lipgloss.Width(result)
	if w != 20 {
		t.Errorf("width = %d, want 20", w)
	}
	// Content should be centered — spaces on both sides
	stripped := stripansi.Strip(result)
	trimmed := strings.TrimSpace(stripped)
	if trimmed != "Hi" {
		t.Errorf("content = %q, want 'Hi'", trimmed)
	}
	// Should have leading spaces (centered)
	if stripped[0] != ' ' {
		t.Error("centered text should have leading space")
	}
}

func TestBaseComponent_RenderInSize_Right(t *testing.T) {
	bc := NewBaseComponent("test")
	bc.SetSize(20, 1)
	bc.SetAlignment(AlignRight)
	result := bc.RenderInSize("Hi")
	stripped := stripansi.Strip(result)
	if !strings.HasSuffix(strings.TrimRight(stripped, " "), "Hi") {
		t.Errorf("right-aligned: %q should end with 'Hi'", stripped)
	}
}

func TestBaseComponent_PreferredSize(t *testing.T) {
	bc := NewBaseComponent("c")
	if bc.PreferredWidth() != 0 || bc.PreferredHeight() != 0 {
		t.Error("preferred size should default to 0")
	}
	bc.SetPreferredWidth(40)
	bc.SetPreferredHeight(10)
	if bc.PreferredWidth() != 40 {
		t.Errorf("preferred width = %d, want 40", bc.PreferredWidth())
	}
	if bc.PreferredHeight() != 10 {
		t.Errorf("preferred height = %d, want 10", bc.PreferredHeight())
	}
}

func TestBaseComponent_RemoveKeyBinding(t *testing.T) {
	bc := NewBaseComponent("c")
	b1 := key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "A"))
	bc.RegisterKeyBinding("ctrl+a", "A", func() tea.Cmd { return nil })
	bc.RegisterKeyBinding("ctrl+b", "B", func() tea.Cmd { return nil })

	if len(bc.KeyBindings()) != 2 {
		t.Fatalf("bindings = %d, want 2", len(bc.KeyBindings()))
	}

	bc.RemoveKeyBinding(b1)
	if len(bc.KeyBindings()) != 1 {
		t.Fatalf("bindings = %d, want 1 after remove", len(bc.KeyBindings()))
	}
	if bc.KeyBindings()[0].Help().Key != "ctrl+b" {
		t.Errorf("remaining binding = %q, want ctrl+b", bc.KeyBindings()[0].Help().Key)
	}

	// Removing a binding that doesn't exist is a no-op
	bc.RemoveKeyBinding(b1)
	if len(bc.KeyBindings()) != 1 {
		t.Fatalf("bindings = %d, want 1 (no-op remove)", len(bc.KeyBindings()))
	}
}

func TestBaseContainer_View_ReturnsEmpty(t *testing.T) {
	bc := NewBaseContainer("c", nil)
	child := NewBaseComponent("child")
	bc.AddChild(child)
	if bc.View() != "" {
		t.Errorf("BaseContainer.View() should return empty string, got %q", bc.View())
	}
}

func TestBaseComponent_HandleEvent_WithParent(t *testing.T) {
	// When a component has a parent, HandleEvent should bubble to the parent
	parent := NewBaseContainer("parent", nil)
	child := NewBaseComponent("child")
	parent.AddChild(child)

	// The base HandleEvent on child should call parent.HandleEvent
	// which also returns (nil, false) since BaseContainer uses BaseComponent.HandleEvent
	cmd, consumed := child.HandleEvent(MouseClickEvent{})
	if consumed {
		t.Error("base HandleEvent should not consume")
	}
	if cmd != nil {
		t.Error("base HandleEvent should return nil cmd")
	}
}

func TestBaseComponent_HandleEvent_NoParent(t *testing.T) {
	bc := NewBaseComponent("c")
	cmd, consumed := bc.HandleEvent(MouseClickEvent{})
	if consumed {
		t.Error("HandleEvent with no parent should return false")
	}
	if cmd != nil {
		t.Error("HandleEvent with no parent should return nil cmd")
	}
}

func TestMouseClickEvent_IsEvent(t *testing.T) {
	var e Event = MouseClickEvent{X: 10, Y: 20, Button: tea.MouseButtonLeft}
	// Just verifying it satisfies the Event interface
	if e == nil {
		t.Error("MouseClickEvent should satisfy Event interface")
	}
}

func TestMouseScrollEvent_IsEvent(t *testing.T) {
	var e Event = MouseScrollEvent{X: 5, Y: 10, Direction: -1}
	if e == nil {
		t.Error("MouseScrollEvent should satisfy Event interface")
	}
}

func TestBaseComponent_RemoveKeyBinding_ResolveNoLongerMatches(t *testing.T) {
	bc := NewBaseComponent("c")
	b := key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "A"))
	bc.RegisterKeyBinding("ctrl+a", "A", func() tea.Cmd { return nil })

	_, found := bc.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !found {
		t.Fatal("should resolve before removal")
	}

	bc.RemoveKeyBinding(b)

	_, found = bc.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Fatal("should not resolve after removal")
	}
}
