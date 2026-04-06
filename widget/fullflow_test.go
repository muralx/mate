package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestFullFlow builds a realistic component tree and exercises all framework
// features: focus cycling, mouse clicks, active/inactive propagation,
// visibility, container focus reactions, and key dispatch.
//
// Tree structure:
//
//	panel (Panel, bordered)
//	├── field1 (Field: "Name:" + TextInput + PopupButton)
//	├── field2 (Field: "Type:" + Toggle)
//	├── field3 (Field: "Notes:" + TextInput) ← starts disabled
//	└── submitBtn (Button "Submit")
func TestFullFlow(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)

	// Build the tree
	panel := NewPanel("panel")
	panel.SetBorder(DefaultBorder())
	panel.SetSize(80, 20)

	// Field 1: Name text input with popup button
	nameInput := NewTextInput("name", 20)
	nameInput.WithPlaceholder("enter name")
	nameBtn := NewButton("name_btn", "[▾]", DefaultPopupButtonStyles())
	field1 := NewField("field1", "Name", nameInput, DefaultFieldStyles())
	field1.AddChild(nameBtn)
	panel.Add(field1, Next)

	// Field 2: Type toggle
	typeToggle := NewToggle("type", "Mode", false, ToggleModeOnOff, DefaultToggleStyles())
	typeToggle.SetLabels("[A]", "[B]")
	field2 := NewField("field2", "Type", typeToggle, DefaultFieldStyles())
	panel.Add(field2, Next)

	// Field 3: Notes text input — starts DISABLED
	notesInput := NewTextInput("notes", 20)
	field3 := NewField("field3", "Notes", notesInput, DefaultFieldStyles())
	field3.SetEnabled(false) // disabled field
	panel.Add(field3, Next)

	// Submit button (direct child of panel, not in a field)
	submitBtn := NewButton("submit", "Submit", DefaultButtonStyles())
	panel.Add(submitBtn, Next)

	// --- Use FocusManager from the input package ---
	// We can't import input from widget (circular), so we test focus
	// mechanics manually here. The FocusManager tests cover the tree walking.

	// === 1. Verify tree structure ===
	if len(panel.Children()) != 4 {
		t.Fatalf("panel children = %d, want 4", len(panel.Children()))
	}
	if nameInput.Parent() != field1 {
		t.Error("nameInput parent should be field1")
	}
	if field1.Parent() != panel {
		t.Error("field1 parent should be panel")
	}

	// === 2. Collect focusable+active leaves manually (simulating FocusManager) ===
	var focusableLeaves []Leaf
	var walkTree func(c Container)
	walkTree = func(c Container) {
		for _, child := range c.Children() {
			if !child.Visible() || !child.Active() {
				continue
			}
			if leaf, ok := child.(Leaf); ok && child.Focusable() {
				focusableLeaves = append(focusableLeaves, leaf)
			}
			if container, ok := child.(Container); ok {
				walkTree(container)
			}
		}
	}
	walkTree(panel)

	// Expected focusable leaves: nameInput, nameBtn, typeToggle, submitBtn
	// (notesInput is in disabled field3, so excluded)
	expectedIDs := []string{"name", "name_btn", "type", "submit"}
	if len(focusableLeaves) != len(expectedIDs) {
		var gotIDs []string
		for _, l := range focusableLeaves {
			gotIDs = append(gotIDs, l.ID())
		}
		t.Fatalf("focusable leaves = %v, want %v", gotIDs, expectedIDs)
	}
	for i, id := range expectedIDs {
		if focusableLeaves[i].ID() != id {
			t.Errorf("leaf[%d] = %s, want %s", i, focusableLeaves[i].ID(), id)
		}
	}

	// === 3. Focus cycling (simulating Tab) ===
	// Focus first leaf
	focusableLeaves[0].SetFocused(true) // nameInput

	if !nameInput.Focused() {
		t.Error("nameInput should be focused")
	}
	if !field1.InnerFocused() {
		t.Error("field1 should have inner focus")
	}
	if !panel.InnerFocused() {
		t.Error("panel should have inner focus")
	}

	// Tab: nameInput → nameBtn
	focusableLeaves[0].SetFocused(false)
	focusableLeaves[1].SetFocused(true)
	if !nameBtn.Focused() {
		t.Error("nameBtn should be focused")
	}
	if !field1.InnerFocused() {
		t.Error("field1 should still have inner focus (nameBtn is child)")
	}

	// Tab: nameBtn → typeToggle
	focusableLeaves[1].SetFocused(false)
	focusableLeaves[2].SetFocused(true)
	if !typeToggle.Focused() {
		t.Error("typeToggle should be focused")
	}
	if field1.InnerFocused() {
		t.Error("field1 should NOT have inner focus (focus moved to field2)")
	}
	if !field2.InnerFocused() {
		t.Error("field2 should have inner focus")
	}

	// Tab: typeToggle → submitBtn (skips disabled field3/notesInput)
	focusableLeaves[2].SetFocused(false)
	focusableLeaves[3].SetFocused(true)
	if !submitBtn.Focused() {
		t.Error("submitBtn should be focused")
	}
	if field2.InnerFocused() {
		t.Error("field2 should NOT have inner focus")
	}

	// === 4. Key dispatch ===
	// Type in nameInput
	nameInput.SetFocused(true)
	submitBtn.SetFocused(false)
	nameInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	nameInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	nameInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if nameInput.Value() != "Joe" {
		t.Errorf("nameInput value = %q, want 'Joe'", nameInput.Value())
	}

	// Toggle the type toggle
	nameInput.SetFocused(false)
	typeToggle.SetFocused(true)
	typeToggle.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !typeToggle.On() {
		t.Error("toggle should be on after space")
	}

	// Press submit button
	submitted := false
	submitBtn.OnPress(func() tea.Cmd { submitted = true; return nil })
	typeToggle.SetFocused(false)
	submitBtn.SetFocused(true)
	submitBtn.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !submitted {
		t.Error("submit callback should have fired")
	}

	// === 5. Inactive component ignores input ===
	notesInput.SetFocused(true) // force focus on disabled field's child
	_, consumed := notesInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if consumed {
		t.Error("inactive input should not consume keys")
	}
	notesInput.SetFocused(false)

	// === 6. Enable field3, verify notesInput becomes focusable ===
	field3.SetEnabled(true)
	if !notesInput.Active() {
		t.Error("notesInput should be active after field3 enabled")
	}

	// Re-collect leaves
	focusableLeaves = nil
	walkTree(panel)
	expectedIDs = []string{"name", "name_btn", "type", "notes", "submit"}
	if len(focusableLeaves) != len(expectedIDs) {
		var gotIDs []string
		for _, l := range focusableLeaves {
			gotIDs = append(gotIDs, l.ID())
		}
		t.Fatalf("after enable: focusable leaves = %v, want %v", gotIDs, expectedIDs)
	}

	// === 7. Mouse click simulation ===
	// Set bounds on nameBtn
	nameBtn.SetPosition(30, 2)
	nameBtn.SetSize(4, 1)
	// Check hit testing
	px, py := nameBtn.Position()
	w, h := nameBtn.Size()
	if !(32 >= px && 32 < px+w && 2 >= py && 2 < py+h) {
		t.Error("click at (32,2) should be in nameBtn bounds")
	}

	// === 8. Rendering — verify all visible components appear ===
	output := stripansi.Strip(panel.View())
	checks := []string{"Name", "Notes", "Submit"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("panel view missing %q", check)
		}
	}

	// === 9. Hide field2, verify Type not in output ===
	field2.SetVisible(false)
	output = stripansi.Strip(panel.View())
	if strings.Contains(output, "Type") {
		t.Error("hidden field2 should not appear in output")
	}
	field2.SetVisible(true) // restore

	// === 10. Disable entire panel — everything inactive ===
	panel.SetEnabled(false)
	if nameInput.Active() {
		t.Error("nameInput should be inactive when panel disabled")
	}
	if submitBtn.Active() {
		t.Error("submitBtn should be inactive when panel disabled")
	}
	// Inactive panel still renders (visible) but in faint style
	output = stripansi.Strip(panel.View())
	if !strings.Contains(output, "Name") {
		t.Error("inactive panel should still render content")
	}
	panel.SetEnabled(true) // restore

	// === 11. Width and alignment ===
	submitBtn.SetSize(20, 1)
	submitBtn.SetAlignment(AlignCenter)
	output = stripansi.Strip(submitBtn.View())
	btnW := lipgloss.Width(submitBtn.View())
	if btnW != 20 {
		t.Errorf("button width = %d, want 20", btnW)
	}
}
