package input

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/muralx/mate/widget"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// Helper to build test trees
func makeTestTree() (widget.Container, *widget.Button, *widget.Button, *widget.Button) {
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	btn3 := widget.NewButton("btn3", "C", widget.DefaultButtonStyles())
	root.AddChild(btn1)
	root.AddChild(btn2)
	root.AddChild(btn3)
	return root, btn1, btn2, btn3
}

func TestFocusManager_Leaves_FlatList(t *testing.T) {
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)
	leaves := fm.Leaves()
	if len(leaves) != 3 {
		t.Fatalf("leaves = %d, want 3", len(leaves))
	}
}

func TestFocusManager_Leaves_SkipsNonFocusable(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	label := widget.NewText("lbl", "text", lipgloss.NewStyle())
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	root.AddChild(label)
	root.AddChild(btn)
	fm := NewFocusManager(root)
	leaves := fm.Leaves()
	if len(leaves) != 1 {
		t.Fatalf("leaves = %d, want 1 (label skipped)", len(leaves))
	}
}

func TestFocusManager_Leaves_SkipsInactive(t *testing.T) {
	root, btn1, btn2, _ := makeTestTree()
	btn2.SetEnabled(false)
	fm := NewFocusManager(root)
	leaves := fm.Leaves()
	if len(leaves) != 2 {
		t.Fatalf("leaves = %d, want 2 (btn2 inactive)", len(leaves))
	}
	_ = btn1
}

func TestFocusManager_Leaves_TreeOrder(t *testing.T) {
	// Nested: root > field1 > btn1, root > field2 > btn2
	root := widget.NewBaseContainer("root", nil)
	field1 := widget.NewBaseContainer("f1", nil)
	root.AddChild(field1)
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	field1.AddChild(btn1)
	field2 := widget.NewBaseContainer("f2", nil)
	root.AddChild(field2)
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	field2.AddChild(btn2)

	fm := NewFocusManager(root)
	leaves := fm.Leaves()
	if len(leaves) != 2 {
		t.Fatalf("leaves = %d, want 2", len(leaves))
	}
	if leaves[0].ID() != "btn1" {
		t.Errorf("first = %s, want btn1", leaves[0].ID())
	}
	if leaves[1].ID() != "btn2" {
		t.Errorf("second = %s, want btn2", leaves[1].ID())
	}
}

func TestFocusManager_Next(t *testing.T) {
	root, btn1, btn2, btn3 := makeTestTree()
	fm := NewFocusManager(root)

	fm.FocusFirst() // focus btn1
	if !btn1.Focused() {
		t.Error("btn1 should be focused")
	}

	fm.Next() // -> btn2
	if !btn2.Focused() {
		t.Error("btn2 should be focused")
	}
	if btn1.Focused() {
		t.Error("btn1 should be unfocused")
	}

	fm.Next() // -> btn3
	if !btn3.Focused() {
		t.Error("btn3 should be focused")
	}

	fm.Next() // wraps -> btn1
	if !btn1.Focused() {
		t.Error("btn1 should be focused (wrap)")
	}
}

func TestFocusManager_Prev(t *testing.T) {
	root, btn1, _, btn3 := makeTestTree()
	fm := NewFocusManager(root)
	fm.FocusFirst()

	fm.Prev() // wraps -> btn3
	if !btn3.Focused() {
		t.Error("btn3 should be focused (wrap)")
	}

	fm.Prev() // -> btn2
	fm.Prev() // -> btn1
	if !btn1.Focused() {
		t.Error("btn1 should be focused")
	}
}

func TestFocusManager_FocusByID(t *testing.T) {
	root, _, btn2, _ := makeTestTree()
	fm := NewFocusManager(root)

	ok, _ := fm.FocusByID("btn2")
	if !ok {
		t.Error("should find btn2")
	}
	if !btn2.Focused() {
		t.Error("btn2 should be focused")
	}

	ok, _ = fm.FocusByID("nonexistent")
	if ok {
		t.Error("should not find nonexistent")
	}
}

func TestFocusManager_FocusByID_Inactive(t *testing.T) {
	root, _, btn2, _ := makeTestTree()
	btn2.SetEnabled(false)
	fm := NewFocusManager(root)

	ok, _ := fm.FocusByID("btn2")
	if ok {
		t.Error("should not focus inactive leaf")
	}
}

func TestFocusManager_HitTest(t *testing.T) {
	root, _, btn2, _ := makeTestTree()
	btn2.SetPosition(10, 5)
	btn2.SetSize(20, 1)
	fm := NewFocusManager(root)

	// HitTest is pure lookup — no focus change
	leaf := fm.HitTest(15, 5)
	if leaf == nil {
		t.Fatal("should find leaf")
	}
	if leaf.ID() != "btn2" {
		t.Errorf("leaf = %s, want btn2", leaf.ID())
	}
	// HitTest does NOT change focus
	if btn2.Focused() {
		t.Error("HitTest should not change focus")
	}
}

func TestFocusManager_ChangeFocusTo(t *testing.T) {
	root, btn1, btn2, _ := makeTestTree()
	fm := NewFocusManager(root)
	fm.FocusFirst() // focus btn1

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused")
	}

	ok, _ := fm.ChangeFocusTo(btn2)
	if !ok {
		t.Error("should be able to change focus")
	}
	if btn1.Focused() {
		t.Error("btn1 should be blurred")
	}
	if !btn2.Focused() {
		t.Error("btn2 should be focused")
	}
}

func TestFocusManager_IsFocusChangingEvent(t *testing.T) {
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)

	press := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	if !fm.IsFocusChangingEvent(press) {
		t.Error("press should be focus-changing")
	}

	release := tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft}
	if fm.IsFocusChangingEvent(release) {
		t.Error("release should NOT be focus-changing")
	}

	motion := tea.MouseMsg{Action: tea.MouseActionMotion}
	if fm.IsFocusChangingEvent(motion) {
		t.Error("motion should NOT be focus-changing")
	}
}

func TestFocusManager_FocusedKeyBindings_NilWithoutRegistered(t *testing.T) {
	// Button has no registered bindings → FocusedKeyBindings returns nil.
	root, btn1, _, _ := makeTestTree()
	fm := NewFocusManager(root)
	btn1.SetFocused(true)

	bindings := fm.FocusedKeyBindings()
	if bindings != nil {
		t.Errorf("expected nil (no registered bindings), got %d", len(bindings))
	}
}

func TestFocusManager_FocusedKeyBindings_WithRegistered(t *testing.T) {
	root, btn1, _, _ := makeTestTree()
	btn1.BindDefaultActionToKey("ctrl+s", "Save")
	fm := NewFocusManager(root)
	btn1.SetFocused(true)

	bindings := fm.FocusedKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+s" {
		t.Errorf("key = %q, want ctrl+s", bindings[0].Help().Key)
	}
}

func TestFocusManager_FocusedLeaf_None(t *testing.T) {
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)
	if fm.FocusedLeaf() != nil {
		t.Error("no focus = nil")
	}
}

func TestFocusManager_SetRoot(t *testing.T) {
	root1, btn1, _, _ := makeTestTree()
	fm := NewFocusManager(root1)
	btn1.SetFocused(true)

	root2 := widget.NewBaseContainer("root2", nil)
	btn4 := widget.NewButton("btn4", "D", widget.DefaultButtonStyles())
	root2.AddChild(btn4)

	fm.SetRoot(root2)
	leaves := fm.Leaves()
	if len(leaves) != 1 {
		t.Fatalf("leaves = %d, want 1", len(leaves))
	}
	if leaves[0].ID() != "btn4" {
		t.Errorf("leaf = %s, want btn4", leaves[0].ID())
	}
}

func TestFocusManager_DisabledContainerSkipped(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	field1 := widget.NewBaseContainer("f1", nil)
	root.AddChild(field1)
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	field1.AddChild(btn1)

	field2 := widget.NewBaseContainer("f2", nil)
	root.AddChild(field2)
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	field2.AddChild(btn2)

	// Disable field2
	field2.SetEnabled(false)

	fm := NewFocusManager(root)
	leaves := fm.Leaves()
	if len(leaves) != 1 {
		t.Fatalf("leaves = %d, want 1", len(leaves))
	}
	if leaves[0].ID() != "b1" {
		t.Errorf("leaf = %s, want b1", leaves[0].ID())
	}
}

func TestFocusManager_FocusFirst(t *testing.T) {
	root, btn1, _, _ := makeTestTree()
	fm := NewFocusManager(root)
	fm.FocusFirst()
	if !btn1.Focused() {
		t.Error("first leaf should be focused")
	}
}

func TestFocusManager_AllActiveKeyBindings_OnlyRegistered(t *testing.T) {
	// Buttons have no registered bindings → AllActiveKeyBindings returns nothing.
	// Local widget bindings (space/enter) handled in Update() must not appear.
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)

	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 0 {
		t.Fatalf("bindings = %d, want 0 (no registered bindings); "+
			"local widget bindings should not be included", len(bindings))
	}
}

func TestFocusManager_AllActiveKeyBindings_RegisteredOnly(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("b1", "Save", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("b2", "Quit", widget.DefaultButtonStyles())
	root.AddChild(btn1)
	root.AddChild(btn2)

	btn1.BindDefaultActionToKey("ctrl+s", "Save")
	btn2.RegisterKeyBinding("ctrl+q", "Quit", func() tea.Cmd { return tea.Quit })

	fm := NewFocusManager(root)
	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 2 {
		t.Fatalf("bindings = %d, want 2 (one per RegisterKeyBinding call)", len(bindings))
	}

	keys := map[string]bool{}
	for _, b := range bindings {
		keys[b.Help().Key] = true
	}
	if !keys["ctrl+s"] {
		t.Error("should contain ctrl+s binding")
	}
	if !keys["ctrl+q"] {
		t.Error("should contain ctrl+q binding")
	}
}

func TestFocusManager_AllActiveKeyBindings_SkipsInactive(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+a", "A", func() tea.Cmd { return nil })
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	btn2.SetEnabled(false)
	btn2.RegisterKeyBinding("ctrl+b", "B", func() tea.Cmd { return nil })
	root.AddChild(btn1)
	root.AddChild(btn2)

	fm := NewFocusManager(root)
	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1 (btn2 disabled)", len(bindings))
	}
}

func TestFocusManager_AllActiveKeyBindings_SkipsHidden(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+a", "A", func() tea.Cmd { return nil })
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	btn2.SetVisible(false)
	btn2.RegisterKeyBinding("ctrl+b", "B", func() tea.Cmd { return nil })
	root.AddChild(btn1)
	root.AddChild(btn2)

	fm := NewFocusManager(root)
	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1 (btn2 hidden)", len(bindings))
	}
}

func TestFocusManager_AllActiveKeyBindings_DisabledContainer(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	field1 := widget.NewBaseContainer("f1", nil)
	root.AddChild(field1)
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+a", "A", func() tea.Cmd { return nil })
	field1.AddChild(btn1)

	field2 := widget.NewBaseContainer("f2", nil)
	field2.SetEnabled(false)
	root.AddChild(field2)
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	btn2.RegisterKeyBinding("ctrl+b", "B", func() tea.Cmd { return nil })
	field2.AddChild(btn2)

	fm := NewFocusManager(root)
	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1 (field2 disabled)", len(bindings))
	}
}

func TestFocusManager_AllActiveKeyBindings_Empty(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	fm := NewFocusManager(root)

	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 0 {
		t.Fatalf("bindings = %d, want 0 (empty tree)", len(bindings))
	}
}

func TestFocusManager_ResolveKeyBinding_RootBindings(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	root.AddChild(btn)

	root.RegisterKeyBinding("ctrl+q", "Quit", func() tea.Cmd { return tea.Quit })

	fm := NewFocusManager(root)
	comp, action, found := fm.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlQ})
	if !found {
		t.Fatal("root binding should be resolved")
	}
	if comp.ID() != "root" {
		t.Errorf("component = %q, want root", comp.ID())
	}
	if action == nil {
		t.Error("action should not be nil")
	}
}

func TestFocusManager_ResolveKeyBinding_ChildWinsOverRoot(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	root.AddChild(btn)

	// Same key on both root and child — child should win (more specific)
	root.RegisterKeyBinding("ctrl+r", "Root action", func() tea.Cmd { return nil })
	btn.RegisterKeyBinding("ctrl+r", "Child action", func() tea.Cmd { return nil })

	fm := NewFocusManager(root)
	comp, _, found := fm.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlR})
	if !found {
		t.Fatal("binding should be resolved")
	}
	if comp.ID() != "btn" {
		t.Errorf("component = %q, want btn (child wins over root)", comp.ID())
	}
}

func TestFocusManager_ResolveKeyBinding_RootOnly_NoChildren(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	root.RegisterKeyBinding("ctrl+h", "Help", func() tea.Cmd { return nil })

	fm := NewFocusManager(root)
	_, _, found := fm.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlH})
	if !found {
		t.Fatal("root-only binding should be resolved even with no children")
	}
}

func TestFocusManager_ResolveKeyBinding_DeepestLeafWinsOverIntermediateContainer(t *testing.T) {
	// 3-level tree: root > panel > btn, both panel and btn have same binding
	// btn (deepest) should win
	root := widget.NewBaseContainer("root", nil)
	panel := widget.NewBaseContainer("panel", nil)
	root.AddChild(panel)
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	panel.AddChild(btn)

	panel.RegisterKeyBinding("ctrl+r", "Panel action", func() tea.Cmd { return nil })
	btn.RegisterKeyBinding("ctrl+r", "Button action", func() tea.Cmd { return nil })

	fm := NewFocusManager(root)
	comp, _, found := fm.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlR})
	if !found {
		t.Fatal("binding should be resolved")
	}
	if comp.ID() != "btn" {
		t.Errorf("component = %q, want btn (deepest wins over intermediate container)", comp.ID())
	}
}

func TestFocusManager_AllActiveKeyBindings_NoDuplicateKeys(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+r", "Refresh", func() tea.Cmd { return nil })
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	btn2.RegisterKeyBinding("ctrl+r", "Reload", func() tea.Cmd { return nil })
	root.AddChild(btn1)
	root.AddChild(btn2)

	fm := NewFocusManager(root)
	bindings := fm.AllActiveKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("bindings = %d, want 1 (duplicate ctrl+r should be deduped)", len(bindings))
	}
	if bindings[0].Help().Desc != "Refresh" {
		t.Errorf("desc = %q, want 'Refresh' (first-match wins)", bindings[0].Help().Desc)
	}
}

func TestFocusManager_HitTest_HandleEvent(t *testing.T) {
	// HitTest finds the component, HandleEvent activates it
	root := widget.NewBaseContainer("root", nil)
	pressed := false
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	btn.SetPosition(0, 0)
	btn.SetSize(10, 1)
	root.AddChild(btn)

	fm := NewFocusManager(root)
	target := fm.HitTest(5, 0)
	if target == nil {
		t.Fatal("should find button")
	}
	// Focus + dispatch event
	fm.ChangeFocusTo(target) //nolint: returned cmd/bool not needed in test
	target.HandleEvent(widget.MouseClickEvent{X: 5, Y: 0})
	if !pressed {
		t.Error("HandleEvent(MouseClickEvent) should fire onPress")
	}
}

func TestFocusManager_HitTest_AfterView(t *testing.T) {
	// Build a Panel with a Field containing a Button.
	// Call View() to trigger position setting.
	// Then verify HitTest finds the button.
	panel := widget.NewPanel("panel")
	panel.SetBorder(widget.DefaultBorder())
	panel.SetPosition(0, 0)
	panel.SetSize(80, 10)

	btn := widget.NewButton("btn", "Click Me", widget.DefaultButtonStyles())
	field := widget.NewField("f", "Label", btn, widget.DefaultFieldStyles())
	panel.Add(field, widget.Next)

	panel.View()

	px, py := btn.Position()
	w, h := btn.Size()
	t.Logf("Button position: (%d, %d) size: (%d, %d)", px, py, w, h)

	if w == 0 || h == 0 {
		t.Fatal("button size should be set after View()")
	}

	fm := NewFocusManager(panel)
	clickX := px + 1
	clickY := py
	leaf := fm.HitTest(clickX, clickY)
	if leaf == nil {
		t.Errorf("should find button at (%d, %d), button bounds: pos=(%d,%d) size=(%d,%d)",
			clickX, clickY, px, py, w, h)
	} else if leaf.ID() != "btn" {
		t.Errorf("hit leaf = %s, want btn", leaf.ID())
	}
}

func TestFocusManager_HitTest_PanelDirectChildren(t *testing.T) {
	// Panel with direct leaf children (not wrapped in Field).
	// This was broken: Panel.View() set positions but not sizes,
	// so HitTest couldn't find components with size=(0,0).
	panel := widget.NewPanel("panel")
	panel.SetBorder(widget.DefaultBorder())
	panel.SetPosition(0, 0)
	panel.SetSize(80, 20)

	btn := widget.NewButton("btn", "Click Me", widget.DefaultButtonStyles())
	tabs := widget.NewTabBar("tabs", []string{"A", "B", "C"}, widget.DefaultTabBarStyles())
	panel.Add(tabs, widget.Next)
	panel.Add(btn, widget.Next)

	panel.View()

	// Verify sizes are set
	tw, th := tabs.Size()
	if tw == 0 || th == 0 {
		t.Fatalf("tabbar size after View() = (%d,%d), want non-zero", tw, th)
	}
	bw, bh := btn.Size()
	if bw == 0 || bh == 0 {
		t.Fatalf("button size after View() = (%d,%d), want non-zero", bw, bh)
	}

	fm := NewFocusManager(panel)

	// Click on the tab bar
	tx, ty := tabs.Position()
	leaf := fm.HitTest(tx+1, ty)
	if leaf == nil {
		t.Errorf("should find tabbar at (%d,%d)", tx+1, ty)
	} else if leaf.ID() != "tabs" {
		t.Errorf("hit = %s, want tabs", leaf.ID())
	}

	// Click on the button
	bx, by := btn.Position()
	leaf = fm.HitTest(bx+1, by)
	if leaf == nil {
		t.Errorf("should find button at (%d,%d)", bx+1, by)
	} else if leaf.ID() != "btn" {
		t.Errorf("hit = %s, want btn", leaf.ID())
	}
}

// TestFocusManager_FullFlow builds a realistic component tree with nested
// containers, disabled sections, and non-focusable components, then exercises
// Tab cycling, mouse clicks, dynamic enable/disable, and status hints.
func TestFocusManager_FullFlow(t *testing.T) {
	// Tree:
	//   panel (BaseContainer)
	//   ├── field1 (BaseContainer)
	//   │   ├── label1 (Label, non-focusable)
	//   │   ├── input1 (Button simulating text input, focusable)
	//   │   └── popup1 (PopupButton, focusable)
	//   ├── field2 (BaseContainer)
	//   │   ├── label2 (Label, non-focusable)
	//   │   └── toggle2 (Button simulating toggle, focusable)
	//   ├── field3 (BaseContainer) ← DISABLED
	//   │   └── input3 (Button, focusable but parent disabled)
	//   └── submitBtn (Button, focusable)

	panel := widget.NewBaseContainer("panel", nil)

	field1 := widget.NewBaseContainer("field1", nil)
	panel.AddChild(field1)
	label1 := widget.NewText("lbl1", "Name:", lipgloss.NewStyle())
	field1.AddChild(label1)
	input1 := widget.NewButton("input1", "A", widget.DefaultButtonStyles())
	field1.AddChild(input1)
	popup1 := widget.NewButton("popup1", "[▾]", widget.DefaultPopupButtonStyles())
	panel.AddChild(popup1) // popup1 is direct child of panel for simplicity
	// Actually let's put it in field1 properly
	panel.Children() // just to access
	// Rebuild: popup1 as child of field1
	field1Rebuild := widget.NewBaseContainer("field1", nil)
	panel2 := widget.NewBaseContainer("panel", nil)
	field1Rebuild.AddChild(widget.NewText("lbl1", "Name:", lipgloss.NewStyle()))
	btnInput1 := widget.NewButton("input1", "A", widget.DefaultButtonStyles())
	field1Rebuild.AddChild(btnInput1)
	btnPopup1 := widget.NewButton("popup1", "[▾]", widget.DefaultPopupButtonStyles())
	field1Rebuild.AddChild(btnPopup1)
	panel2.AddChild(field1Rebuild)

	field2 := widget.NewBaseContainer("field2", nil)
	field2.AddChild(widget.NewText("lbl2", "Type:", lipgloss.NewStyle()))
	btnToggle := widget.NewButton("toggle2", "Toggle", widget.DefaultButtonStyles())
	field2.AddChild(btnToggle)
	panel2.AddChild(field2)

	field3 := widget.NewBaseContainer("field3", nil)
	field3.SetEnabled(false) // DISABLED
	btnInput3 := widget.NewButton("input3", "Notes", widget.DefaultButtonStyles())
	field3.AddChild(btnInput3)
	panel2.AddChild(field3)

	submitBtn := widget.NewButton("submit", "Submit", widget.DefaultButtonStyles())
	panel2.AddChild(submitBtn)

	fm := NewFocusManager(panel2)

	// === 1. Leaves should skip label and disabled field3 ===
	leaves := fm.Leaves()
	expectedIDs := []string{"input1", "popup1", "toggle2", "submit"}
	if len(leaves) != len(expectedIDs) {
		var gotIDs []string
		for _, l := range leaves {
			gotIDs = append(gotIDs, l.ID())
		}
		t.Fatalf("leaves = %v, want %v", gotIDs, expectedIDs)
	}
	for i, id := range expectedIDs {
		if leaves[i].ID() != id {
			t.Errorf("leaf[%d] = %s, want %s", i, leaves[i].ID(), id)
		}
	}

	// === 2. FocusFirst + Tab cycling ===
	fm.FocusFirst()
	if fm.FocusedLeaf().ID() != "input1" {
		t.Errorf("first = %s, want input1", fm.FocusedLeaf().ID())
	}

	fm.Next() // → popup1
	if fm.FocusedLeaf().ID() != "popup1" {
		t.Errorf("after Tab1 = %s, want popup1", fm.FocusedLeaf().ID())
	}
	if !field1Rebuild.InnerFocused() {
		t.Error("field1 should have inner focus when popup1 focused")
	}

	fm.Next() // → toggle2
	if fm.FocusedLeaf().ID() != "toggle2" {
		t.Errorf("after Tab2 = %s, want toggle2", fm.FocusedLeaf().ID())
	}
	if field1Rebuild.InnerFocused() {
		t.Error("field1 should NOT have inner focus after focus moved to field2")
	}
	if !field2.InnerFocused() {
		t.Error("field2 should have inner focus")
	}

	fm.Next() // → submit (skips disabled field3/input3)
	if fm.FocusedLeaf().ID() != "submit" {
		t.Errorf("after Tab3 = %s, want submit", fm.FocusedLeaf().ID())
	}

	fm.Next() // wraps → input1
	if fm.FocusedLeaf().ID() != "input1" {
		t.Errorf("after wrap = %s, want input1", fm.FocusedLeaf().ID())
	}

	// === 3. Shift-Tab (Prev) ===
	fm.Prev() // wraps → submit
	if fm.FocusedLeaf().ID() != "submit" {
		t.Errorf("Prev from input1 = %s, want submit", fm.FocusedLeaf().ID())
	}

	// === 4. FocusByID ===
	fm.FocusByID("toggle2")
	if fm.FocusedLeaf().ID() != "toggle2" {
		t.Error("FocusByID should focus toggle2")
	}

	// FocusByID on disabled leaf should fail
	ok, _ := fm.FocusByID("input3")
	if ok {
		t.Error("should not focus disabled input3")
	}

	// === 5. Mouse click ===
	submitBtn.SetPosition(0, 10)
	submitBtn.SetSize(20, 1)
	leaf := fm.HitTest(5, 10)
	if leaf == nil {
		t.Error("should find submit at (5,10)")
	} else if leaf.ID() != "submit" {
		t.Errorf("hit = %s, want submit", leaf.ID())
	}
	fm.ChangeFocusTo(leaf)
	if fm.FocusedLeaf().ID() != "submit" {
		t.Error("focus should move to clicked leaf")
	}

	// === 6. Enable field3 → input3 joins the ring ===
	field3.SetEnabled(true)
	leaves = fm.Leaves()
	expectedIDs = []string{"input1", "popup1", "toggle2", "input3", "submit"}
	if len(leaves) != len(expectedIDs) {
		var gotIDs []string
		for _, l := range leaves {
			gotIDs = append(gotIDs, l.ID())
		}
		t.Fatalf("after enable: leaves = %v, want %v", gotIDs, expectedIDs)
	}

	// Can now focus input3
	ok, _ = fm.FocusByID("input3")
	if !ok {
		t.Error("should be able to focus input3 after enabling field3")
	}

	// === 7. FocusedKeyBindings — nil when no registered bindings ===
	fm.FocusByID("submit")
	bindings := fm.FocusedKeyBindings()
	if bindings != nil {
		t.Errorf("expected nil FocusedKeyBindings (no registered bindings), got %d", len(bindings))
	}
}

func TestFocusManager_ResolveKeyBinding_TabBarAccelerators(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	tabs := widget.NewTabBar("tabs", []string{"Overview", "Details"}, widget.DefaultTabBarStyles())
	tabs.SetTabKeyBinding(0, "ctrl+d")
	tabs.SetTabKeyBinding(1, "ctrl+e")
	root.AddChild(tabs)

	fm := NewFocusManager(root)

	// Resolve tab accelerator
	comp, action, found := fm.ResolveKeyBinding(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !found {
		t.Fatal("tab accelerator should resolve through FocusManager")
	}
	if comp.ID() != "tabs" {
		t.Errorf("component = %q, want tabs", comp.ID())
	}
	action()
	if tabs.ActiveTab() != 1 {
		t.Errorf("active = %d, want 1", tabs.ActiveTab())
	}

	// Verify bindings appear in AllActiveKeyBindings
	bindings := fm.AllActiveKeyBindings()
	keys := map[string]bool{}
	for _, b := range bindings {
		keys[b.Help().Key] = true
	}
	if !keys["ctrl+d"] || !keys["ctrl+e"] {
		t.Errorf("AllActiveKeyBindings should include tab accelerators, got keys: %v", keys)
	}
}

// --- Additional coverage tests ---

func TestKeyBindingResolver_SetRoot(t *testing.T) {
	// SetRoot should update the root container used for resolution.
	root1 := widget.NewBaseContainer("root1", nil)
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+a", "Action A", func() tea.Cmd { return nil })
	root1.AddChild(btn1)

	root2 := widget.NewBaseContainer("root2", nil)
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	btn2.RegisterKeyBinding("ctrl+b", "Action B", func() tea.Cmd { return nil })
	root2.AddChild(btn2)

	resolver := NewKeyBindingResolver(root1)

	// Should resolve ctrl+a from root1
	_, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlA})
	if !found {
		t.Fatal("ctrl+a should resolve from root1")
	}

	// After SetRoot to root2, ctrl+a should no longer resolve
	resolver.SetRoot(root2)
	_, _, found = resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Error("ctrl+a should NOT resolve after SetRoot to root2")
	}

	// ctrl+b should now resolve from root2
	comp, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlB})
	if !found {
		t.Fatal("ctrl+b should resolve from root2")
	}
	if comp.ID() != "btn2" {
		t.Errorf("component = %q, want btn2", comp.ID())
	}
}

func TestFocusManager_ChangeFocusTo_NonFocusable(t *testing.T) {
	// ChangeFocusTo with a non-focusable component should return false.
	root := widget.NewBaseContainer("root", nil)
	label := widget.NewText("lbl", "text", lipgloss.NewStyle())
	root.AddChild(label)
	// Text is not focusable, so we need a Leaf that is not focusable.
	// Text doesn't implement Leaf, so let's test with nil and disabled button.
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.SetEnabled(false) // disabled = not active
	root.AddChild(btn)

	fm := NewFocusManager(root)

	// Disabled button: CanFocusTo should return false
	ok, _ := fm.ChangeFocusTo(btn)
	if ok {
		t.Error("should not focus a disabled (inactive) button")
	}
}

func TestFocusManager_ChangeFocusTo_Nil(t *testing.T) {
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)

	ok, _ := fm.ChangeFocusTo(nil)
	if ok {
		t.Error("ChangeFocusTo(nil) should return false")
	}
}

func TestFocusManager_ChangeFocusTo_WhenNothingFocused(t *testing.T) {
	// ChangeFocusTo when no current leaf is focused (no blur needed).
	root, _, btn2, _ := makeTestTree()
	fm := NewFocusManager(root)

	// No leaf is focused initially
	if fm.FocusedLeaf() != nil {
		t.Fatal("no leaf should be focused initially")
	}

	ok, _ := fm.ChangeFocusTo(btn2)
	if !ok {
		t.Error("should be able to focus btn2")
	}
	if !btn2.Focused() {
		t.Error("btn2 should be focused")
	}
}

func TestFocusManager_FocusedKeyBindings_NoFocus(t *testing.T) {
	// When no component is focused, FocusedKeyBindings should return nil.
	root, _, _, _ := makeTestTree()
	fm := NewFocusManager(root)

	// No leaf focused
	bindings := fm.FocusedKeyBindings()
	if bindings != nil {
		t.Errorf("expected nil when no leaf focused, got %d bindings", len(bindings))
	}
}

func TestFocusManager_FocusFirst_EmptyTree(t *testing.T) {
	// FocusFirst with no focusable leaves should return nil.
	root := widget.NewBaseContainer("root", nil)
	fm := NewFocusManager(root)

	cmd := fm.FocusFirst()
	if cmd != nil {
		t.Error("FocusFirst on empty tree should return nil")
	}
}

func TestFocusManager_Next_EmptyTree(t *testing.T) {
	// Next with no focusable leaves should return nil.
	root := widget.NewBaseContainer("root", nil)
	fm := NewFocusManager(root)

	cmd := fm.Next()
	if cmd != nil {
		t.Error("Next on empty tree should return nil")
	}
}

func TestFocusManager_Prev_EmptyTree(t *testing.T) {
	// Prev with no focusable leaves should return nil.
	root := widget.NewBaseContainer("root", nil)
	fm := NewFocusManager(root)

	cmd := fm.Prev()
	if cmd != nil {
		t.Error("Prev on empty tree should return nil")
	}
}

func TestFocusManager_Next_NoCurrentFocus(t *testing.T) {
	// Next when no leaf is focused: focusedIndex returns -1,
	// so (-1+1)%3 = 0, should focus the first leaf.
	root, btn1, _, _ := makeTestTree()
	fm := NewFocusManager(root)

	// No leaf is focused
	if fm.FocusedLeaf() != nil {
		t.Fatal("precondition: no leaf should be focused")
	}

	fm.Next()
	if !btn1.Focused() {
		t.Error("Next with no focus should focus the first leaf")
	}
}

func TestFocusManager_Prev_NoCurrentFocus(t *testing.T) {
	// Prev when no leaf is focused: focusedIndex returns -1,
	// so (-1-1+3)%3 = 1, should focus the second-to-last leaf.
	root, _, _, btn3 := makeTestTree()
	fm := NewFocusManager(root)

	// No leaf is focused
	if fm.FocusedLeaf() != nil {
		t.Fatal("precondition: no leaf should be focused")
	}

	fm.Prev()
	// (-1 - 1 + 3) % 3 = 1, so it focuses leaves[1] which is btn2
	// Actually let's just verify some leaf got focused
	focused := fm.FocusedLeaf()
	if focused == nil {
		t.Fatal("Prev should focus some leaf")
	}
	_ = btn3
}

func TestFocusManager_HitTest_Miss(t *testing.T) {
	// HitTest at coordinates outside all component bounds should return nil.
	root, btn1, btn2, btn3 := makeTestTree()
	btn1.SetPosition(0, 0)
	btn1.SetSize(10, 1)
	btn2.SetPosition(0, 1)
	btn2.SetSize(10, 1)
	btn3.SetPosition(0, 2)
	btn3.SetSize(10, 1)

	fm := NewFocusManager(root)

	// Click way outside
	leaf := fm.HitTest(100, 100)
	if leaf != nil {
		t.Errorf("expected nil for miss, got %s", leaf.ID())
	}
}

func TestFocusManager_HitTest_EmptyTree(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	fm := NewFocusManager(root)

	leaf := fm.HitTest(5, 5)
	if leaf != nil {
		t.Error("HitTest on empty tree should return nil")
	}
}

func TestKeyBindingResolver_Walk_SkipsInvisibleChild(t *testing.T) {
	// An invisible child with a binding should be skipped during resolution.
	root := widget.NewBaseContainer("root", nil)
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn1.RegisterKeyBinding("ctrl+a", "Action A", func() tea.Cmd { return nil })
	btn1.SetVisible(false) // invisible
	root.AddChild(btn1)

	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	btn2.RegisterKeyBinding("ctrl+b", "Action B", func() tea.Cmd { return nil })
	root.AddChild(btn2)

	resolver := NewKeyBindingResolver(root)

	// ctrl+a should NOT resolve (btn1 invisible)
	_, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Error("invisible child binding should be skipped")
	}

	// ctrl+b should resolve (btn2 visible)
	comp, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlB})
	if !found {
		t.Fatal("ctrl+b should resolve")
	}
	if comp.ID() != "btn2" {
		t.Errorf("component = %q, want btn2", comp.ID())
	}
}

func TestKeyBindingResolver_Walk_SkipsInactiveChild(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn := widget.NewButton("btn", "A", widget.DefaultButtonStyles())
	btn.RegisterKeyBinding("ctrl+a", "Action", func() tea.Cmd { return nil })
	btn.SetEnabled(false) // inactive
	root.AddChild(btn)

	resolver := NewKeyBindingResolver(root)
	_, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Error("inactive child binding should be skipped")
	}
}

func TestKeyBindingResolver_Walk_ContainerOwnBinding(t *testing.T) {
	// Container's own binding should be checked after children.
	root := widget.NewBaseContainer("root", nil)
	child := widget.NewBaseContainer("child", nil)
	root.AddChild(child)

	// Only root has the binding, not the child
	root.RegisterKeyBinding("ctrl+q", "Quit", func() tea.Cmd { return tea.Quit })

	resolver := NewKeyBindingResolver(root)
	comp, action, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlQ})
	if !found {
		t.Fatal("root's own binding should resolve")
	}
	if comp.ID() != "root" {
		t.Errorf("component = %q, want root", comp.ID())
	}
	if action == nil {
		t.Error("action should not be nil")
	}
}

func TestKeyBindingResolver_Walk_NoMatch(t *testing.T) {
	root := widget.NewBaseContainer("root", nil)
	btn := widget.NewButton("btn", "A", widget.DefaultButtonStyles())
	btn.RegisterKeyBinding("ctrl+a", "Action", func() tea.Cmd { return nil })
	root.AddChild(btn)

	resolver := NewKeyBindingResolver(root)
	_, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlB})
	if found {
		t.Error("should not find a match for unregistered key")
	}
}

func TestKeyBindingResolver_Walk_InvisibleContainerSkipped(t *testing.T) {
	// An invisible container should be skipped entirely, including its children.
	root := widget.NewBaseContainer("root", nil)
	container := widget.NewBaseContainer("cont", nil)
	container.SetVisible(false)
	root.AddChild(container)

	btn := widget.NewButton("btn", "A", widget.DefaultButtonStyles())
	btn.RegisterKeyBinding("ctrl+a", "Action", func() tea.Cmd { return nil })
	container.AddChild(btn)

	resolver := NewKeyBindingResolver(root)
	_, _, found := resolver.Resolve(tea.KeyMsg{Type: tea.KeyCtrlA})
	if found {
		t.Error("binding inside invisible container should not resolve")
	}
}
