package window

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muralx/mate/input"
	"github.com/muralx/mate/widget"
)

func TestPopupWindow_Construction(t *testing.T) {
	popup := NewPopupWindow("p", "Title", DefaultPopupStyles())
	if popup.ID() != "p" {
		t.Errorf("ID = %q, want %q", popup.ID(), "p")
	}
	if popup.Title() != "Title" {
		t.Errorf("Title = %q, want %q", popup.Title(), "Title")
	}
}

func TestPopupWindow_IsContainer(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	popup.Add(btn, widget.TCBCenter)

	if btn.Parent() == nil {
		t.Error("child should have a parent set")
	}
}

func TestPopupWindow_CloseProducesMsg(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())

	cmd := popup.Close("confirmed")
	msg := cmd()
	closeMsg, ok := msg.(closePopupMsg)
	if !ok {
		t.Fatal("Close should produce closePopupMsg")
	}
	if closeMsg.result != "confirmed" {
		t.Errorf("result = %v, want 'confirmed'", closeMsg.result)
	}
	if closeMsg.popup != popup {
		t.Error("closePopupMsg should reference the popup")
	}
}

func TestPopupWindow_CloseNilResult(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())

	cmd := popup.Close(nil)
	msg := cmd()
	closeMsg := msg.(closePopupMsg)
	if closeMsg.result != nil {
		t.Error("nil close should have nil result")
	}
}

func TestPopupWindow_EscapeClosesWithNil(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	popup.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(popup)
	fm.FocusFirst()

	cmd := popup.update(tea.KeyMsg{Type: tea.KeyEscape}, fm)
	if cmd == nil {
		t.Fatal("Escape should produce a close command")
	}
	msg := cmd()
	closeMsg, ok := msg.(closePopupMsg)
	if !ok {
		t.Fatal("Escape should produce closePopupMsg")
	}
	if closeMsg.result != nil {
		t.Error("Escape should close with nil result")
	}
}

func TestPopupWindow_NonEscapeRoutesToBaseWindow(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	pressed := false
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	popup.Add(btn, widget.TCBCenter)

	fm := input.NewFocusManager(popup)
	fm.FocusFirst()

	popup.update(tea.KeyMsg{Type: tea.KeySpace}, fm)
	if !pressed {
		t.Error("non-Escape keys should route to BaseWindow (button press)")
	}
}

func TestPopupWindow_TabCyclesInPopup(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	btn1 := widget.NewButton("b1", "Yes", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("b2", "No", widget.DefaultButtonStyles())
	popup.Add(btn1, widget.TCBTop)
	popup.Add(btn2, widget.TCBCenter)

	fm := input.NewFocusManager(popup)
	fm.FocusFirst()

	if !btn1.Focused() {
		t.Fatal("btn1 should be focused first")
	}

	popup.update(tea.KeyMsg{Type: tea.KeyTab}, fm)
	if !btn2.Focused() {
		t.Error("Tab should cycle to btn2")
	}
}

func TestPopupWindow_OnResultStored(t *testing.T) {
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	var received any
	popup.OnResult(func(value any) tea.Cmd {
		received = value
		return nil
	})

	// OnResult callback is stored but not called yet
	if received != nil {
		t.Error("OnResult should not fire on set")
	}
	if popup.onResult == nil {
		t.Error("onResult callback should be stored")
	}
}
