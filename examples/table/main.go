// Minimal Mate example: a scrollable table with cursor selection.
//
// Demonstrates: Table, ColumnDef (fixed + flex widths), SliceDataSource,
// per-column CellRenderer, OnRowClick.
//
// Run with: go run ./examples/table
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
	win := window.NewWindow("table")

	// Panel uses TCB so the Table in the center slot flexes to fill
	// all available height. Default Vertical layout would collapse
	// it to its preferred (≈1 line) height.
	panel := widget.NewPanel("panel", widget.TCB)
	panel.SetBorder(widget.DefaultBorder())
	panel.SetTitle(" Log viewer ")

	columns := []widget.ColumnDef{
		{Title: "TIME", Width: 10},
		{Title: "LEVEL", Width: 6, Renderer: levelRenderer},
		{Title: "MESSAGE", Width: 0}, // 0 = flex, takes remaining
	}

	ds := widget.NewSliceDataSource([][]string{
		{"12:00:01", "INFO", "Task completed successfully"},
		{"12:00:02", "WARN", "Slow response detected"},
		{"12:00:03", "ERROR", "Connection refused"},
		{"12:00:04", "INFO", "Retry scheduled"},
		{"12:00:05", "INFO", "Reconnected"},
	})

	status := widget.NewText("status", " Click or press Enter on a row.", lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")))

	table := widget.NewTable("logs", columns, ds, widget.DefaultTableStyles())
	table.OnRowClick(func(row int) tea.Cmd {
		status.SetText(fmt.Sprintf(" Selected row %d — %s %s: %s",
			row,
			ds.CellData(row, 0),
			ds.CellData(row, 1),
			ds.CellData(row, 2)))
		return nil
	})
	panel.Add(table, widget.TCBCenter)

	win.Add(panel, widget.TCBCenter)
	win.Add(status, widget.TCBBottom)

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

// levelRenderer colors the LEVEL column based on severity.
func levelRenderer(ds widget.TableDataSource, row, col int, selected bool, width int, styles widget.TableStyles) string {
	value := widget.PrepareCell(ds, row, col, width)
	style := styles.Cell
	if selected {
		style = styles.Selected.Width(width)
	}
	switch ds.CellData(row, col) {
	case "ERROR":
		style = style.Foreground(lipgloss.Color("#ef5350")).Bold(true)
	case "WARN":
		style = style.Foreground(lipgloss.Color("#ffb74d"))
	case "INFO":
		style = style.Foreground(lipgloss.Color("#81c784"))
	}
	return style.Render(value)
}
