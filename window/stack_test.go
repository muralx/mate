package window

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
)

func TestStack_NewStack(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
	if s.hasPopup() {
		t.Error("should not have popup")
	}
}

func TestStack_Push(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "B", widget.DefaultButtonStyles()), widget.TCBCenter)
	s.push(popup)

	if s.len() != 2 {
		t.Errorf("len = %d, want 2", s.len())
	}
	if !s.hasPopup() {
		t.Error("should have popup")
	}
	if win.Enabled() {
		t.Error("base window should be disabled")
	}
	if !popup.Enabled() {
		t.Error("popup should be enabled")
	}
}

func TestStack_Pop_RestoresFocus(t *testing.T) {
	win := NewWindow("main")
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	win.Add(btn1, widget.TCBTop)
	win.Add(btn2, widget.TCBCenter)
	s := newStack(win)

	// Focus btn2 before opening popup
	s.baseFM.Next()
	if !btn2.Focused() {
		t.Fatal("btn2 should be focused")
	}

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("pb", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	s.push(popup)

	s.pop()

	if !btn2.Focused() {
		t.Error("focus should be restored to btn2 after popup close")
	}
}

func TestStack_ClosePopup_CallsOnResult(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	var received any
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		received = value
		return nil
	})
	s.push(popup)

	s.closePopup(popup, "confirmed")

	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
	if received != "confirmed" {
		t.Errorf("OnResult received = %v, want 'confirmed'", received)
	}
}

func TestStack_ClosePopup_NilResult_StillCallsOnResult(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	var called bool
	var received any = "sentinel"
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		called = true
		received = value
		return nil
	})
	s.push(popup)
	s.closePopup(popup, nil)

	if !called {
		t.Error("OnResult should be called even with nil result (cancellation)")
	}
	if received != nil {
		t.Errorf("received = %v, want nil", received)
	}
}

func TestStack_ClosePopup_OnResult_CanShowNewPopup(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	popup1 := NewPopupWindow("p1", "First", DefaultPopupStyles())
	popup1.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup1.OnResult(func(value any) tea.Cmd {
		popup2 := NewPopupWindow("p2", "Second", DefaultPopupStyles())
		popup2.Add(widget.NewButton("b3", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
		return s.push(popup2)
	})
	s.push(popup1)
	s.closePopup(popup1, "done")

	if s.len() != 2 {
		t.Errorf("len = %d, want 2 (base + second popup)", s.len())
	}
}

func TestStack_Update_ProcessesClosePopupMsg(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	var received any
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		received = value
		return nil
	})
	s.push(popup)

	// Simulate: popup.Close -> Cmd -> closePopupMsg -> Stack processes it
	cmd := popup.Close("yes")
	msg := cmd()
	s.update(msg)

	if received != "yes" {
		t.Errorf("OnResult received = %v, want 'yes'", received)
	}
	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
}

func TestStack_NeverPopsBase(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	s.pop()
	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
}

func TestStack_Update_RoutesToPopup(t *testing.T) {
	win := NewWindow("main")
	winBtn := widget.NewButton("wb", "Win", widget.DefaultButtonStyles())
	winPressed := false
	winBtn.OnPress(func() tea.Cmd { winPressed = true; return nil })
	win.Add(winBtn, widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popBtn := widget.NewButton("pb", "Pop", widget.DefaultButtonStyles())
	popPressed := false
	popBtn.OnPress(func() tea.Cmd { popPressed = true; return nil })
	popup.Add(popBtn, widget.TCBCenter)
	s.push(popup)

	// Space should go to popup's focused button, not window's
	s.update(tea.KeyMsg{Type: tea.KeySpace})

	if winPressed {
		t.Error("window button should not receive events when popup is open")
	}
	if !popPressed {
		t.Error("popup button should receive events")
	}
}

func TestStack_View_BaseOnly(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewText("t", "base content", lipgloss.NewStyle()), widget.TCBCenter)
	s := newStack(win)

	view := s.view(80, 24)
	if !strings.Contains(view, "base content") {
		t.Error("should render base window content")
	}
}

func TestStack_View_PopupOverlay(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewText("t", "base", lipgloss.NewStyle()), widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "TestPopup", DefaultPopupStyles())
	popup.Add(widget.NewText("pt", "popup text", lipgloss.NewStyle()), widget.TCBCenter)
	s.push(popup)

	view := s.view(80, 24)
	if !strings.Contains(view, "popup text") {
		t.Error("should render popup content")
	}
}

// --- Push: multiple popups and focus management ---

func TestStack_PushMultiplePopups_DisablesPrevious(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b1", "Main", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)
	s := newStack(win)

	popup1 := NewPopupWindow("p1", "First", DefaultPopupStyles())
	pb1 := widget.NewButton("pb1", "P1", widget.DefaultButtonStyles())
	popup1.Add(pb1, widget.TCBCenter)
	s.push(popup1)

	if win.Enabled() {
		t.Error("base should be disabled after first push")
	}
	if !popup1.Enabled() {
		t.Error("popup1 should be enabled")
	}

	popup2 := NewPopupWindow("p2", "Second", DefaultPopupStyles())
	pb2 := widget.NewButton("pb2", "P2", widget.DefaultButtonStyles())
	popup2.Add(pb2, widget.TCBCenter)
	s.push(popup2)

	if s.len() != 3 {
		t.Errorf("len = %d, want 3", s.len())
	}
	if popup1.Enabled() {
		t.Error("popup1 should be disabled after second push")
	}
	if !popup2.Enabled() {
		t.Error("popup2 should be enabled")
	}
	if !pb2.Focused() {
		t.Error("popup2's button should be focused")
	}
}

func TestStack_PopMultiplePopups_RestoresState(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b1", "Main", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)
	s := newStack(win)

	popup1 := NewPopupWindow("p1", "First", DefaultPopupStyles())
	pb1 := widget.NewButton("pb1", "P1", widget.DefaultButtonStyles())
	popup1.Add(pb1, widget.TCBCenter)
	s.push(popup1)

	popup2 := NewPopupWindow("p2", "Second", DefaultPopupStyles())
	pb2 := widget.NewButton("pb2", "P2", widget.DefaultButtonStyles())
	popup2.Add(pb2, widget.TCBCenter)
	s.push(popup2)

	// Pop popup2 — popup1 should be re-enabled
	s.pop()
	if s.len() != 2 {
		t.Errorf("len = %d, want 2", s.len())
	}
	if !popup1.Enabled() {
		t.Error("popup1 should be re-enabled after popping popup2")
	}

	// Pop popup1 — base should be re-enabled
	s.pop()
	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
	if !win.Enabled() {
		t.Error("base window should be re-enabled after all popups popped")
	}
}

// --- Pop: focus restoration with no previous focused ID ---

func TestStack_Pop_NoPrevFocusedID_FocusesFirst(t *testing.T) {
	win := NewWindow("main")
	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	win.Add(btn1, widget.TCBTop)
	win.Add(btn2, widget.TCBCenter)
	s := newStack(win)

	// Manually blur all leaves so there's no focused leaf when push happens
	btn1.SetFocused(false)
	btn2.SetFocused(false)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("pb", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	s.push(popup)

	// prevFocusedID is "" since nothing was focused
	s.pop()

	// FocusFirst should be called as fallback
	if !btn1.Focused() {
		t.Error("should focus first leaf when no prev focused ID")
	}
}

// --- ClosePopup: no onResult callback ---

func TestStack_ClosePopup_NoOnResult(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	// Don't set OnResult
	s.push(popup)

	cmd := s.closePopup(popup, "result")
	// Should not panic and should still pop
	if s.len() != 1 {
		t.Errorf("len = %d, want 1", s.len())
	}
	_ = cmd
}

func TestStack_ClosePopup_OnResult_ReturnsCmd(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		return func() tea.Msg { return "result-msg" }
	})
	s.push(popup)

	cmd := s.closePopup(popup, "yes")
	if cmd == nil {
		t.Fatal("closePopup should return a batched cmd when onResult returns cmd")
	}
}

// --- Update: mouse event with popup adjusts coordinates ---

func TestStack_Update_MouseWithPopup_AdjustsCoordinates(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "Main", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popBtn := widget.NewButton("pb", "OK", widget.DefaultButtonStyles())
	pressed := false
	popBtn.OnPress(func() tea.Cmd { pressed = true; return nil })
	popup.Add(popBtn, widget.TCBCenter)
	s.push(popup)

	// Render to set popup offset
	s.view(80, 24)

	// Send a mouse event — coordinates should be adjusted by popupOffset
	// The popup button's position is relative to the popup content area
	bx, by := popBtn.Position()
	ox := s.popupOffset.X
	oy := s.popupOffset.Y
	s.update(tea.MouseMsg{
		X: bx + ox, Y: by + oy,
		Action: tea.MouseActionPress, Button: tea.MouseButtonLeft,
	})

	if !pressed {
		t.Error("mouse click with offset adjustment should reach popup button")
	}
}

// --- Update: closePopupMsg processed by stack ---

func TestStack_Update_ClosePopupMsg(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	received := false
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		received = true
		return nil
	})
	s.push(popup)

	// Send closePopupMsg directly
	s.update(closePopupMsg{popup: popup, result: "done"})

	if !received {
		t.Error("closePopupMsg should trigger OnResult")
	}
	if s.len() != 1 {
		t.Errorf("len = %d, want 1 after close", s.len())
	}
}

// --- WithOnUpdate: onUpdate returns cmd, both non-nil ---

func TestStack_WithOnUpdate_OnUpdateReturnsCmd(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	updateFired := false
	win.OnUpdate(func() tea.Cmd {
		updateFired = true
		return func() tea.Msg { return "update-msg" }
	})

	// Call withOnUpdate with nil cmd — should return just updateCmd
	cmd := s.withOnUpdate(nil)
	if cmd == nil {
		t.Fatal("withOnUpdate should return updateCmd when cmd is nil")
	}
	if !updateFired {
		t.Error("onUpdate should have been called")
	}
}

func TestStack_WithOnUpdate_BothCmdAndUpdateCmd(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	win.OnUpdate(func() tea.Cmd {
		return func() tea.Msg { return "update-msg" }
	})

	existingCmd := func() tea.Msg { return "existing-msg" }

	// Both cmd and onUpdate return non-nil — should batch
	cmd := s.withOnUpdate(existingCmd)
	if cmd == nil {
		t.Fatal("withOnUpdate should return batched cmd when both are non-nil")
	}
}

func TestStack_WithOnUpdate_UpdateCmdNil(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	win.OnUpdate(func() tea.Cmd {
		return nil // onUpdate returns nil cmd
	})

	existingCmd := func() tea.Msg { return "existing-msg" }

	// cmd is non-nil but updateCmd is nil — should return just cmd
	cmd := s.withOnUpdate(existingCmd)
	if cmd == nil {
		t.Fatal("withOnUpdate should return cmd when updateCmd is nil")
	}
}

func TestStack_WithOnUpdate_PopupOnUpdate(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "Main", widget.DefaultButtonStyles()), widget.TCBCenter)
	s := newStack(win)

	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)

	popupUpdateFired := false
	popup.OnUpdate(func() tea.Cmd {
		popupUpdateFired = true
		return nil
	})
	s.push(popup)

	// withOnUpdate should use popup's onUpdate when popup is on top
	s.withOnUpdate(nil)
	if !popupUpdateFired {
		t.Error("popup's onUpdate should fire when popup is on top of stack")
	}
}
