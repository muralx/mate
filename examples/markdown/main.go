// Minimal Mate example: a scrollable Markdown viewer.
//
// Demonstrates: MarkdownTextArea, ScrollableText scroll keys (inherited),
// OSC 8 hyperlinks for terminals that support them (kitty, WezTerm, iTerm2).
//
// Run with: go run ./examples/markdown
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

const content = `# Mate

A component framework for building terminal UIs in Go, built on top of
**Bubble Tea**. Compose widgets and set callbacks — no custom Update or
View methods.

## What you're looking at

This text is rendered by ` + "`widget.MarkdownTextArea`" + ` which extends
` + "`ScrollableText`" + ` with a small Markdown subset. You can scroll with
the arrow keys, ` + "`j`" + `/` + "`k`" + `, Page Up/Down, Home, and End.

## Supported markdown

- H1, H2, H3 headings
- **Bold** with double asterisks
- Inline ` + "`code`" + ` with backticks
- Fenced code blocks
- Horizontal rules
- Links — see the Mate repo at [muralx/mate](https://github.com/muralx/mate)

## Code blocks

` + "```" + `
mt := widget.NewMarkdownTextArea("docs", widget.DefaultMarkdownTextAreaStyles())
mt.SetMarkdown("# Hello\n\n**World**")
` + "```" + `

---

Press Ctrl+Q to quit.
`

func main() {
	win := window.NewWindow("markdown")

	panel := widget.NewPanel("panel")
	panel.SetBorder(widget.DefaultBorder())
	panel.SetTitle(" Markdown Viewer ")

	mt := widget.NewMarkdownTextArea("md", widget.DefaultMarkdownTextAreaStyles())
	mt.SetMarkdown(content)
	panel.Add(mt, widget.Next)

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
