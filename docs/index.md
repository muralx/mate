# Mate Framework Documentation

Mate is a Go component framework for building terminal user interfaces on top of [Bubble Tea](https://github.com/charmbracelet/bubbletea). It provides a composable component tree with focus management, keyboard and mouse event routing, global key bindings, and popup window support.

## Documentation

1. **[Getting Started](getting-started.md)** — Installation, first application, core concepts
2. **[Components](components.md)** — All available widgets: Button, TextInput, Toggle, CheckboxList, TabComponent, Table, ScrollableText, Card, Text, Field
3. **[Layout](layout.md)** — Panel layouts (Vertical, Horizontal, TCB), preferred sizes, fields, custom containers
4. **[Focus and Keyboard](focus-and-keyboard.md)** — Focus cycling, key bindings, global shortcuts, key event flow
5. **[Mouse](mouse.md)** — Hit testing, click-to-focus, mouse event handling
6. **[Windows and Popups](windows-and-popups.md)** — MainWindow, PopupWindow, App, popup lifecycle
7. **[Styling](styling.md)** — lipgloss integration, style structs, BorderConfig, themes
8. **[API Reference](api-reference.md)** — Complete type and method reference

## Architecture at a Glance

```
┌─────────────────────────────────────────────┐
│               App (tea.Model)                │  window/
│  ┌────────────────────────────────────────┐  │
│  │      MainWindow (TCB layout)           │  │
│  │  ┌──────────────────────────────────┐  │  │
│  │  │  Panel (TCBTop: tab bar)         │  │  │  widget/
│  │  ├──────────────────────────────────┤  │  │
│  │  │  Panel (TCBCenter: content)      │  │  │
│  │  │  ┌────────────────────────────┐  │  │  │
│  │  │  │  Field: Label + TextInput  │  │  │  │
│  │  │  ├────────────────────────────┤  │  │  │
│  │  │  │  Button "Submit"           │  │  │  │
│  │  │  └────────────────────────────┘  │  │  │
│  │  ├──────────────────────────────────┤  │  │
│  │  │  Text (TCBBottom: status bar)    │  │  │
│  │  └──────────────────────────────────┘  │  │
│  └────────────────────────────────────────┘  │
│                FocusManager                   │  input/
└─────────────────────────────────────────────┘
```

Three packages with clear responsibilities:

- **`widget/`** — Components, containers, layouts, and the `Component` interface
- **`input/`** — Focus management and key binding resolution
- **`window/`** — Screen management, popup stack, overlay rendering

## Quick Example

```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/muralx/mate/widget"
    "github.com/muralx/mate/window"
)

func main() {
    win := window.NewWindow("main") // defaults to TCB layout

    panel := widget.NewPanel("panel")
    panel.SetBorder(widget.DefaultBorder())

    nameInput := widget.NewTextInput("name", 30)
    nameInput.WithPlaceholder("Enter your name")
    field := widget.NewField("name_field", "Name", nameInput, widget.DefaultFieldStyles())
    panel.Add(field, widget.Next)

    submitBtn := widget.NewButton("submit", "Submit", widget.DefaultButtonStyles())
    submitBtn.OnPress(func() tea.Cmd {
        fmt.Println("Submitted:", nameInput.Value())
        return tea.Quit
    })
    // Global shortcut: Ctrl+S triggers submit from anywhere
    submitBtn.BindDefaultActionToKey("ctrl+s", "Submit")
    panel.Add(submitBtn, widget.Next)

    win.Add(panel, widget.TCBCenter)

    win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
        if msg.String() == "ctrl+q" {
            return tea.Quit
        }
        return nil
    })

    app := window.NewApp(win)
    p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion())
    p.Run()
}
```
