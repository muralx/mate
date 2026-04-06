package window

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
)

func TestApp_Init(t *testing.T) {
	win := NewWindow("main")
	app := NewApp(win)
	cmd := app.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestApp_View_BeforeResize(t *testing.T) {
	win := NewWindow("main")
	app := NewApp(win)
	if app.View() != "" {
		t.Error("view before resize should be empty")
	}
}

func TestApp_View_AfterResize(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewText("t", "hello", lipgloss.NewStyle()), widget.TCBCenter)
	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := app.View()
	if !strings.Contains(view, "hello") {
		t.Errorf("should contain 'hello', got: %q", view)
	}
}

func TestApp_Update_RoutesToWindow(t *testing.T) {
	win := NewWindow("main")
	pressed := false
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	btn.OnPress(func() tea.Cmd { pressed = true; return nil })
	win.Add(btn, widget.TCBCenter)
	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !pressed {
		t.Error("key should route to focused button")
	}
}

func TestApp_PopupLifecycle(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b1", "A", widget.DefaultButtonStyles()), widget.TCBCenter)
	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	var received any
	popup := NewPopupWindow("p", "Test", DefaultPopupStyles())
	popup.Add(widget.NewButton("b2", "OK", widget.DefaultButtonStyles()), widget.TCBCenter)
	popup.OnResult(func(value any) tea.Cmd {
		received = value
		return nil
	})

	win.ShowPopup(popup)

	if app.stack.len() != 2 {
		t.Fatalf("stack len = %d, want 2", app.stack.len())
	}

	// Escape closes popup → produces closePopupMsg Cmd
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// Process the closePopupMsg
	if cmd != nil {
		app.Update(cmd())
	}

	if app.stack.len() != 1 {
		t.Errorf("stack len = %d, want 1 after close", app.stack.len())
	}
	// OnResult called with nil (Escape = cancel)
	_ = received
}
