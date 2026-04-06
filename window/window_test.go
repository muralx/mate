package window

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
)

func TestMainWindow_Construction(t *testing.T) {
	win := NewWindow("main")
	if win.ID() != "main" {
		t.Errorf("ID = %q, want %q", win.ID(), "main")
	}
	if win.Focusable() {
		t.Error("should not be focusable")
	}
}

func TestMainWindow_Add(t *testing.T) {
	win := NewWindow("main")
	btn := widget.NewButton("b", "OK", widget.DefaultButtonStyles())
	win.Add(btn, widget.TCBCenter)

	// btn is added to the content panel, which is a child of win
	if btn.Parent() == nil {
		t.Error("child should have a parent set")
	}
}

func TestMainWindow_InheritsView(t *testing.T) {
	win := NewWindow("main")
	txt := widget.NewText("t", "hello", lipgloss.NewStyle())
	win.Add(txt, widget.TCBCenter)

	view := win.View()
	if !strings.Contains(view, "hello") {
		t.Error("should render children via inherited View")
	}
}
