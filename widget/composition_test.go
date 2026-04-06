package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestComposition_FieldWithTextInputAndPopupButton verifies that a Field's
// label highlights (Hot) when any of its focusable children receive focus.
func TestComposition_FieldWithTextInputAndPopupButton(t *testing.T) {
	input := NewTextInput("nodes", 20)
	btn := NewButton("nodes_btn", "[▾]", DefaultPopupButtonStyles())

	field := NewField("nodes_field", "Nodes", input, DefaultFieldStyles())
	field.AddChild(btn)

	// No focus - label Normal
	if field.InnerFocused() {
		t.Error("no inner focus expected")
	}

	// Focus TextInput - label should go Hot
	input.SetFocused(true)
	if !field.InnerFocused() {
		t.Error("field should report inner focus")
	}

	// Move focus to PopupButton - label should still be Hot
	input.SetFocused(false)
	btn.SetFocused(true)
	if !field.InnerFocused() {
		t.Error("field should still report inner focus")
	}

	// Remove all focus
	btn.SetFocused(false)
	if field.InnerFocused() {
		t.Error("no inner focus expected")
	}
}

// TestComposition_PanelWithMultipleFields verifies that a Panel's border
// highlights when ANY field's child is focused.
func TestComposition_PanelWithMultipleFields(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())

	btn1 := NewButton("b1", "One", DefaultButtonStyles())
	field1 := NewField("f1", "A", btn1, DefaultFieldStyles())

	btn2 := NewButton("b2", "Two", DefaultButtonStyles())
	field2 := NewField("f2", "B", btn2, DefaultFieldStyles())

	btn3 := NewButton("b3", "Three", DefaultButtonStyles())
	field3 := NewField("f3", "C", btn3, DefaultFieldStyles())

	panel.Add(field1, Next)
	panel.Add(field2, Next)
	panel.Add(field3, Next)

	normalOutput := panel.View()

	// Focus a button in the second field
	btn2.SetFocused(true)
	if !panel.InnerFocused() {
		t.Error("panel should have inner focus when a nested child is focused")
	}

	activeOutput := panel.View()
	if normalOutput == activeOutput {
		t.Error("panel border should change when a nested child is focused")
	}
}

// TestComposition_ThreeLevelNesting verifies InnerFocused propagates through
// Panel > Field > TextInput and that Active propagation works when disabling Panel.
func TestComposition_ThreeLevelNesting(t *testing.T) {
	panel := NewPanel("panel")
	panel.SetBorder(DefaultBorder())
	input := NewTextInput("name", 20)
	field := NewField("name_field", "Name", input, DefaultFieldStyles())
	panel.Add(field, Next)

	// Focus the deeply nested TextInput
	input.SetFocused(true)
	if !field.InnerFocused() {
		t.Error("field should have inner focus")
	}
	if !panel.InnerFocused() {
		t.Error("panel should have inner focus")
	}

	// Disable panel - TextInput should become inactive
	panel.SetEnabled(false)
	if input.Active() {
		t.Error("input should be inactive when panel disabled")
	}
	if field.Active() {
		t.Error("field should be inactive when panel disabled")
	}

	// Re-enable panel
	panel.SetEnabled(true)
	if !input.Active() {
		t.Error("input should be active again")
	}
}

// TestComposition_NestedActive verifies that disabling a Field within a Panel
// only affects that field's children, not siblings.
func TestComposition_NestedActive(t *testing.T) {
	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())

	btn1 := NewButton("b1", "One", DefaultButtonStyles())
	field1 := NewField("f1", "A", btn1, DefaultFieldStyles())

	btn2 := NewButton("b2", "Two", DefaultButtonStyles())
	field2 := NewField("f2", "B", btn2, DefaultFieldStyles())

	panel.Add(field1, Next)
	panel.Add(field2, Next)

	// Disable field2 only
	field2.SetEnabled(false)
	if !btn1.Active() {
		t.Error("btn1 should still be active")
	}
	if btn2.Active() {
		t.Error("btn2 should be inactive")
	}
}

// TestComposition_Visibility verifies that hiding a Field causes its children
// to not appear in the Panel's rendered output.
func TestComposition_Visibility(t *testing.T) {
	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())

	txt1 := NewText("t1", "Visible:", lipgloss.NewStyle())
	field1 := NewField("f1", "Visible", txt1, DefaultFieldStyles())

	txt2 := NewText("t2", "Hidden:", lipgloss.NewStyle())
	field2 := NewField("f2", "Hidden", txt2, DefaultFieldStyles())
	field2.SetVisible(false)

	panel.Add(field1, Next)
	panel.Add(field2, Next)

	output := stripansi.Strip(panel.View())
	if !strings.Contains(output, "Visible") {
		t.Error("visible field should be in output")
	}
	if strings.Contains(output, "Hidden") {
		t.Error("hidden field should NOT be in output")
	}
}

// TestComposition_InactiveLeaf_NoCallback verifies that a disabled button
// does not fire its callback.
func TestComposition_InactiveLeaf_NoCallback(t *testing.T) {
	pressed := false
	btn := NewButton("b", "OK", DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	btn.SetEnabled(false)

	_, consumed := btn.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("inactive button should not consume keys")
	}
	if pressed {
		t.Error("inactive button should not fire callback")
	}
}

// TestComposition_SingleLeafInContainer verifies that a single Button in a
// Field works correctly for focus, rendering, and key handling.
func TestComposition_SingleLeafInContainer(t *testing.T) {
	btn := NewButton("btn", "Go", DefaultButtonStyles())
	field := NewField("f", "Action", btn, DefaultFieldStyles())

	// Not focused
	if field.InnerFocused() {
		t.Error("field should not have inner focus initially")
	}

	// Focus the button
	btn.SetFocused(true)
	if !field.InnerFocused() {
		t.Error("field should have inner focus when button is focused")
	}

	// Verify rendering contains both label and button
	output := stripansi.Strip(field.View())
	if !strings.Contains(output, "Action") {
		t.Error("missing label in output")
	}
	if !strings.Contains(output, "[ Go ]") {
		t.Error("missing button in output")
	}

	// Verify key handling works
	pressed := false
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	btn.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !pressed {
		t.Error("button should fire callback when focused and active")
	}
}

// TestComposition_FieldView_ContainsAllChildren verifies that a Field with
// Label + TextInput + PopupButton renders all visible children.
func TestComposition_FieldView_ContainsAllChildren(t *testing.T) {
	input := NewTextInput("nodes", 20)
	input.SetValue("fix-pub-1")
	btn := NewButton("nodes_btn", "[▾]", DefaultPopupButtonStyles())

	field := NewField("f", "Nodes", input, DefaultFieldStyles())
	field.AddChild(btn)

	output := stripansi.Strip(field.View())
	if !strings.Contains(output, "Nodes") {
		t.Error("missing label")
	}
	if !strings.Contains(output, "fix-pub-1") {
		t.Error("missing input value")
	}
	if !strings.Contains(output, "[▾]") {
		t.Error("missing popup button")
	}
}
