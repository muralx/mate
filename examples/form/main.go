// Minimal Mate example: a labelled text input + submit button.
//
// Demonstrates: Panel, Field, TextInput, Button, OnSubmit / OnPress,
// global key bindings via BindDefaultActionToKey.
//
// Run with: go run ./examples/form
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

func main() {
	win := window.NewWindow("form")

	panel := widget.NewPanel("panel")
	panel.SetBorder(widget.DefaultBorder())
	panel.SetTitle(" Submit a name ")

	nameInput := widget.NewTextInput("name", 30)
	nameInput.WithPlaceholder("Enter your name")
	nameInput.OnSubmit(func(value string) tea.Cmd {
		fmt.Println("Submitted:", value)
		return tea.Quit
	})
	panel.Add(widget.NewField("name_field", "Name", nameInput, widget.DefaultFieldStyles()), widget.Next)

	submit := widget.NewButton("submit", "Submit", widget.DefaultButtonStyles())
	submit.OnPress(func() tea.Cmd {
		fmt.Println("Submitted:", nameInput.Value())
		return tea.Quit
	})
	// Ctrl+S triggers submit from anywhere in the tree.
	submit.BindDefaultActionToKey("ctrl+s", "Submit")
	panel.Add(submit, widget.Next)

	win.Add(panel, widget.TCBCenter)

	win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "ctrl+q" {
			return tea.Quit
		}
		return nil
	})

	app := window.NewApp(win)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
