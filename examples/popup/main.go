// Minimal Mate example: a popup window that returns a result.
//
// Demonstrates: NewPopupWindow, Close(result), OnResult callback,
// focus restoration after popup closes.
//
// Run with: go run ./examples/popup
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

func main() {
	win := window.NewWindow("main")

	status := widget.NewText("status", "No choice yet — press Enter on the button.", lipgloss.NewStyle())

	panel := widget.NewPanel("panel")
	panel.SetBorder(widget.DefaultBorder())
	panel.SetTitle(" Popup demo ")

	open := widget.NewButton("open", "Open dialog", widget.DefaultButtonStyles())
	open.OnPress(func() tea.Cmd {
		popup := buildConfirm()
		popup.OnResult(func(result any) tea.Cmd {
			if b, ok := result.(bool); ok && b {
				status.SetText("Confirmed!")
			} else if ok {
				status.SetText("Declined.")
			} else {
				status.SetText("Dialog cancelled.")
			}
			return nil
		})
		return win.ShowPopup(popup)
	})
	panel.Add(open, widget.Next)
	panel.Add(status, widget.Next)

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

func buildConfirm() *window.PopupWindow {
	popup := window.NewPopupWindow("confirm", "Confirm", window.DefaultPopupStyles())

	prompt := widget.NewText("prompt", "Proceed?", lipgloss.NewStyle().Width(30).Align(lipgloss.Center))
	popup.Add(prompt, widget.Next)

	row := widget.NewPanel("row", widget.Horizontal)
	row.SetSpacing(2)
	yes := widget.NewButton("yes", "Yes", widget.DefaultPopupButtonStyles())
	yes.OnPress(func() tea.Cmd { return popup.Close(true) })
	no := widget.NewButton("no", "No", widget.DefaultPopupButtonStyles())
	no.OnPress(func() tea.Cmd { return popup.Close(false) })
	row.Add(yes, widget.Next)
	row.Add(no, widget.Next)
	popup.Add(row, widget.Next)

	return popup
}
