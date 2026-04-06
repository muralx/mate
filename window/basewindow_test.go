package window

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/input"
	"github.com/muralx/mate/widget"
)

// testWindow wraps BaseWindow since BaseWindow is not instantiated directly.
type testWindow struct {
	BaseWindow
}

func newTestWindow(id string) *testWindow {
	tw := &testWindow{}
	tw.BaseWindow = newBaseWindow(id, tw)
	return tw
}

// --- Container behavior ---

func TestBaseWindow_AddChild_ParentBackReference(t *testing.T) {
	tw := newTestWindow("win")
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	tw.Add(btn, widget.TCBCenter)

	// The button's parent is the content panel, not the window directly
	if btn.Parent() == nil {
		t.Error("child should have a parent set")
	}
}

func TestBaseWindow_NotFocusable(t *testing.T) {
	tw := newTestWindow("win")
	if tw.Focusable() {
		t.Error("BaseWindow (container) should not be focusable")
	}
}

// --- View ---

func TestBaseWindow_View_RendersChildren(t *testing.T) {
	tw := newTestWindow("win")
	tw.Add(widget.NewText("t1", "Hello", lipgloss.NewStyle()), widget.TCBTop)
	tw.Add(widget.NewText("t2", "World", lipgloss.NewStyle()), widget.TCBCenter)

	view := stripansi.Strip(tw.View())
	if !strings.Contains(view, "Hello") {
		t.Error("view should contain 'Hello'")
	}
	if !strings.Contains(view, "World") {
		t.Error("view should contain 'World'")
	}
}

func TestBaseWindow_View_SkipsInvisibleChildren(t *testing.T) {
	tw := newTestWindow("win")
	t1 := widget.NewText("t1", "Visible", lipgloss.NewStyle())
	t2 := widget.NewText("t2", "Hidden", lipgloss.NewStyle())
	t2.SetVisible(false)
	tw.Add(t1, widget.TCBTop)
	tw.Add(t2, widget.TCBCenter)

	view := stripansi.Strip(tw.View())
	if !strings.Contains(view, "Visible") {
		t.Error("view should contain visible child")
	}
	if strings.Contains(view, "Hidden") {
		t.Error("view should not contain hidden child")
	}
}

// --- Size propagation ---

func TestBaseWindow_View_PropagatesWidth(t *testing.T) {
	tw := newTestWindow("w")
	tw.SetSize(80, 24)

	txt := widget.NewText("t1", "Hello", lipgloss.NewStyle())
	tw.Add(txt, widget.TCBCenter)

	tw.View()

	w, _ := txt.Size()
	if w != 80 {
		t.Errorf("child width = %d, want 80", w)
	}
}

func TestBaseWindow_View_FlexHeight(t *testing.T) {
	tw := newTestWindow("w")
	tw.SetSize(80, 24)

	fixed := widget.NewText("t1", "Header", lipgloss.NewStyle())
	fixed.SetPreferredHeight(4)
	flex := widget.NewText("t2", "Body", lipgloss.NewStyle())
	tw.Add(fixed, widget.TCBTop)
	tw.Add(flex, widget.TCBCenter)

	tw.View()

	_, fh := fixed.Size()
	if fh != 4 {
		t.Errorf("fixed child height = %d, want 4", fh)
	}

	_, flexH := flex.Size()
	expected := 24 - 4 // remaining height
	if flexH != expected {
		t.Errorf("flex child height = %d, want %d", flexH, expected)
	}
}

// --- Tab focus cycling ---

func TestBaseWindow_TabCyclesFocus(t *testing.T) {
	tw := newTestWindow("win")
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	tw.Add(btn1, widget.TCBTop)
	tw.Add(btn2, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused initially")
	}

	tw.update(tea.KeyMsg{Type: tea.KeyTab}, fm)
	if !btn2.Focused() {
		t.Error("Tab should move focus to btn2")
	}
	if btn1.Focused() {
		t.Error("btn1 should lose focus after Tab")
	}

	// Tab again wraps back to btn1
	tw.update(tea.KeyMsg{Type: tea.KeyTab}, fm)
	if !btn1.Focused() {
		t.Error("Tab should wrap focus back to btn1")
	}
}

func TestBaseWindow_ShiftTabCyclesFocusBackwards(t *testing.T) {
	tw := newTestWindow("win")
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	tw.Add(btn1, widget.TCBTop)
	tw.Add(btn2, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	// Shift-Tab from first should wrap to last
	tw.update(tea.KeyMsg{Type: tea.KeyShiftTab}, fm)
	if !btn2.Focused() {
		t.Error("Shift-Tab from first should wrap to btn2")
	}

	tw.update(tea.KeyMsg{Type: tea.KeyShiftTab}, fm)
	if !btn1.Focused() {
		t.Error("Shift-Tab should move focus back to btn1")
	}
}

// --- Key routing to focused leaf ---

func TestBaseWindow_KeyRoutingToFocusedLeaf(t *testing.T) {
	tw := newTestWindow("win")
	pressed := false
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	tw.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	// Space triggers button press
	tw.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, fm)
	if !pressed {
		t.Error("space should trigger focused button press")
	}
}

// --- OnKeyPress fallthrough ---

func TestBaseWindow_OnKeyPressFallthrough(t *testing.T) {
	tw := newTestWindow("win")
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	tw.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	fellThrough := false
	tw.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "x" {
			fellThrough = true
		}
		return nil
	})

	// "x" is not consumed by button, should fallthrough to OnKeyPress
	tw.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, fm)
	if !fellThrough {
		t.Error("unconsumed key should fallthrough to OnKeyPress")
	}
}

func TestBaseWindow_OnKeyPressFallthrough_NotCalledWhenConsumed(t *testing.T) {
	tw := newTestWindow("win")
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { return nil })
	tw.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	called := false
	tw.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		called = true
		return nil
	})

	// Space is consumed by button, OnKeyPress should NOT fire
	tw.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, fm)
	if called {
		t.Error("OnKeyPress should not be called when key is consumed by focused leaf")
	}
}

// --- Mouse event routing ---

func TestBaseWindow_MouseClickFocusesAndTriggers(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	pressed := false
	btn2.OnPress(func() tea.Cmd { pressed = true; return nil })
	tw.Add(btn1, widget.TCBTop)
	tw.Add(btn2, widget.TCBCenter)

	// Render to set positions/sizes
	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused initially")
	}

	// Click on btn2 (positioned below btn1)
	_, y2 := btn2.Position()
	tw.update(tea.MouseMsg{X: 0, Y: y2, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, fm)

	if !btn2.Focused() {
		t.Error("click should focus btn2")
	}
	if !pressed {
		t.Error("click should trigger btn2 OnPress")
	}
}

func TestBaseWindow_MouseMotionDoesNotTrigger(t *testing.T) {
	tw := newTestWindow("win")
	pressed := false
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	tw.Add(btn, widget.TCBCenter)

	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	tw.update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionMotion, Button: tea.MouseButtonNone}, fm)
	if pressed {
		t.Error("mouse motion should not trigger button press")
	}
}

func TestBaseWindow_MouseReleaseDoesNotTrigger(t *testing.T) {
	tw := newTestWindow("win")
	pressed := false
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	tw.Add(btn, widget.TCBCenter)

	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	tw.update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft}, fm)
	if pressed {
		t.Error("mouse release should not trigger button press")
	}
}

// --- Key binding resolution ---

func TestBaseWindow_KeyBindingResolution(t *testing.T) {
	tw := newTestWindow("win")
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	fired := false
	btn.RegisterKeyBinding("ctrl+s", "", func() tea.Cmd {
		fired = true
		return nil
	})
	tw.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	tw.update(tea.KeyMsg{Type: tea.KeyCtrlS}, fm)
	if !fired {
		t.Error("registered key binding should fire on matching key")
	}
}

func TestBaseWindow_KeyBindingResolution_FocusesTarget(t *testing.T) {
	tw := newTestWindow("win")
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())
	fired := false
	btn2.RegisterKeyBinding("ctrl+b", "", func() tea.Cmd {
		fired = true
		return nil
	})
	tw.Add(btn1, widget.TCBTop)
	tw.Add(btn2, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused initially")
	}

	tw.update(tea.KeyMsg{Type: tea.KeyCtrlB}, fm)
	if !fired {
		t.Error("binding should fire")
	}
	if !btn2.Focused() {
		t.Error("key binding resolution should focus the target component")
	}
}

// --- update dispatching ---

func TestBaseWindow_UpdateDispatchesKeyAndMouse(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	pressed := 0
	btn.OnPress(func() tea.Cmd { pressed++; return nil })
	tw.Add(btn, widget.TCBCenter)

	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	// Key dispatch
	tw.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, fm)
	if pressed != 1 {
		t.Errorf("key dispatch: pressed = %d, want 1", pressed)
	}

	// Mouse dispatch — click at button's position
	bx, by := btn.Position()
	tw.update(tea.MouseMsg{X: bx, Y: by, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, fm)
	if pressed != 2 {
		t.Errorf("mouse dispatch: pressed = %d, want 2", pressed)
	}
}

func TestBaseWindow_UpdateIgnoresUnknownMsg(t *testing.T) {
	tw := newTestWindow("win")
	fm := input.NewFocusManager(tw)

	// Should not panic on unknown message types
	cmd := tw.update(tea.WindowSizeMsg{Width: 80, Height: 24}, fm)
	if cmd != nil {
		t.Error("unknown message should return nil cmd")
	}
}

// --- OnUpdate and ActiveKeyBindings ---

func TestBaseWindow_OnUpdate_FiresAfterKeyEvent(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	updateCalled := false
	win.OnUpdate(func() tea.Cmd {
		updateCalled = true
		return nil
	})

	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !updateCalled {
		t.Error("OnUpdate should fire after key event")
	}
}

func TestBaseWindow_OnUpdate_FiresAfterMouseEvent(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	btn.SetPosition(0, 0)
	btn.SetSize(10, 1)
	win.Add(btn, widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	updateCalled := false
	win.OnUpdate(func() tea.Cmd {
		updateCalled = true
		return nil
	})

	app.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !updateCalled {
		t.Error("OnUpdate should fire after mouse event")
	}
}

func TestBaseWindow_OnUpdate_FiresAfterPopupClose(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	updateCount := 0
	win.OnUpdate(func() tea.Cmd {
		updateCount++
		return nil
	})

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	win.ShowPopup(popup)

	// Escape closes popup → produces closePopupMsg
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// Process the closePopupMsg
	if cmd != nil {
		app.Update(cmd())
	}

	if updateCount == 0 {
		t.Error("OnUpdate should fire after popup close restores base window")
	}
}

func TestBaseWindow_ActiveKeyBindings(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b", "Save", widget.DefaultButtonStyles())
	btn.RegisterKeyBinding("ctrl+s", "Save", func() tea.Cmd { return nil })
	win.Add(btn, widget.TCBCenter)

	app := NewApp(win)
	_ = app // stack is set on win

	bindings := win.ActiveKeyBindings()
	if len(bindings) != 1 {
		t.Fatalf("ActiveKeyBindings() = %d, want 1", len(bindings))
	}
	if bindings[0].Help().Key != "ctrl+s" {
		t.Errorf("binding key = %q, want 'ctrl+s'", bindings[0].Help().Key)
	}
}

func TestBaseWindow_ActiveKeyBindings_NoStack(t *testing.T) {
	tw := newTestWindow("win")
	bindings := tw.ActiveKeyBindings()
	if bindings != nil {
		t.Error("ActiveKeyBindings should return nil when no stack")
	}
}

// --- ShowPopup via stack ---

func TestBaseWindow_ShowPopup_WithStack(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b1", "Main", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)
	app := NewApp(win) // creates stack
	_ = app

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popBtn := widget.NewButton("pb", "OK", widget.DefaultButtonStyles())
	popup.Add(popBtn, widget.TCBCenter)

	cmd := win.ShowPopup(popup)

	// ShowPopup should return a cmd (from push, which focuses the popup's first child)
	// The popup button should be focused after push
	if cmd != nil {
		// Execute the batch cmd (focus commands)
		cmd()
	}
	if !popBtn.Focused() {
		t.Error("popup's button should be focused after ShowPopup")
	}
}

func TestBaseWindow_ShowPopup_NoStack_ReturnsNil(t *testing.T) {
	tw := newTestWindow("win")
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	cmd := tw.ShowPopup(popup)
	if cmd != nil {
		t.Error("ShowPopup without stack should return nil")
	}
}

// --- Mouse: HitTest returns nil ---

func TestBaseWindow_MouseClick_MissesAllComponents(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	tw.Add(btn, widget.TCBCenter)
	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	// Click far outside any component — HitTest returns nil
	cmd := tw.update(tea.MouseMsg{X: 79, Y: 23, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, fm)
	if cmd != nil {
		t.Error("click on empty area should return nil cmd")
	}
}

// --- Mouse: scroll wheel dispatches MouseScrollEvent ---

func TestBaseWindow_MouseWheelDown_DispatchesScrollEvent(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)

	// Use a ScrollableText which handles MouseScrollEvent
	st := widget.NewScrollableText("st", widget.DefaultScrollableTextStyles())
	st.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	tw.Add(st, widget.TCBCenter)
	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	// Wheel down on the scrollable text
	_, y := st.Position()
	cmd := tw.update(tea.MouseMsg{X: 0, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown}, fm)
	// Scroll event should be dispatched without error; cmd may or may not be nil
	_ = cmd
}

func TestBaseWindow_MouseWheelUp_DispatchesScrollEvent(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)

	st := widget.NewScrollableText("st", widget.DefaultScrollableTextStyles())
	st.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	tw.Add(st, widget.TCBCenter)
	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	_, y := st.Position()
	// First scroll down, then scroll up to test direction = -1
	tw.update(tea.MouseMsg{X: 0, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown}, fm)
	tw.update(tea.MouseMsg{X: 0, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown}, fm)
	cmd := tw.update(tea.MouseMsg{X: 0, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelUp}, fm)
	_ = cmd
}

// --- Mouse: click on non-focusable component returns nil ---

func TestBaseWindow_MouseClick_NonFocusableTarget(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)

	// Text is not focusable
	txt := widget.NewText("txt", "Hello", lipgloss.NewStyle())
	tw.Add(txt, widget.TCBCenter)
	tw.View()

	fm := input.NewFocusManager(tw)

	_, y := txt.Position()
	// Left click on a non-focusable text — IsFocusChangingEvent=true but CanFocusTo=false → returns nil
	cmd := tw.update(tea.MouseMsg{X: 0, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, fm)
	if cmd != nil {
		t.Error("clicking on non-focusable component should return nil")
	}
}

// --- Mouse: press dispatches click event and returns cmd ---

func TestBaseWindow_MousePress_ReturnsCmd(t *testing.T) {
	tw := newTestWindow("win")
	tw.SetSize(80, 24)

	pressed := false
	btn := widget.NewButton("btn", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd {
		pressed = true
		return func() tea.Msg { return "test-msg" }
	})
	tw.Add(btn, widget.TCBCenter)
	tw.View()

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	bx, by := btn.Position()
	cmd := tw.update(tea.MouseMsg{X: bx, Y: by, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, fm)
	if !pressed {
		t.Error("button should be pressed via click")
	}
	if cmd == nil {
		t.Error("click that produces cmd should return non-nil cmd")
	}
}

// --- handleKey: no focused leaf and no OnKeyPress ---

func TestBaseWindow_HandleKey_NoFocusedLeaf_NoOnKeyPress(t *testing.T) {
	tw := newTestWindow("win")
	// No children, no focused leaf, no onKeyPress
	fm := input.NewFocusManager(tw)

	cmd := tw.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, fm)
	if cmd != nil {
		t.Error("key with no focused leaf and no onKeyPress should return nil")
	}
}

// --- handleKey: key binding on unfocused component focuses it then fires ---

func TestBaseWindow_HandleKey_BindingFocusesAndReturnsCmd(t *testing.T) {
	tw := newTestWindow("win")
	btn1 := widget.NewButton("btn1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("btn2", "B", widget.DefaultButtonStyles())

	actionFired := false
	btn2.RegisterKeyBinding("ctrl+b", "action B", func() tea.Cmd {
		actionFired = true
		return func() tea.Msg { return "action-result" }
	})

	tw.Add(btn1, widget.TCBTop)
	tw.Add(btn2, widget.TCBCenter)

	fm := input.NewFocusManager(tw)
	fm.FocusFirst()

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused initially")
	}

	cmd := tw.update(tea.KeyMsg{Type: tea.KeyCtrlB}, fm)
	if !actionFired {
		t.Error("key binding action should fire")
	}
	if !btn2.Focused() {
		t.Error("btn2 should be focused after key binding resolution")
	}
	if cmd == nil {
		t.Error("key binding that returns a cmd should produce non-nil result")
	}
}
