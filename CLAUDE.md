# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mate is a Go TUI component framework. It provides a composable component system for building terminal applications on top of Bubble Tea (charmbracelet/bubbletea).

## Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test -v ./widget
go test -v ./input
go test -v ./window

# Run a single test
go test -v ./widget -run TestButton_Press_Space

# Build check
go build ./...

# Format and vet
go fmt ./...
go vet ./...
```

## Architecture

Three packages with clear responsibilities:

**`widget/`** — UI components implementing the `Component` interface defined in `component.go`:
- `BaseComponent` → `FocusableComponent` → leaf widgets (Button, TextInput, Checkbox, Toggle, TabBar, Table, ScrollableText, FormattedTextInput)
- `BaseComponent` → `BaseContainer` → container widgets (Panel, TabComponent, Field, Card, Text)
- `Panel` is the universal container with configurable layout: `Vertical` (stack top-to-bottom), `Horizontal` (stack left-to-right), or `TCB` (Top-Center-Bottom, center flexes)
- Children are added with `Add(child, Position)` using `Next`, `TCBTop`, `TCBCenter`, `TCBBottom`
- Border is configured with `SetBorder(BorderConfig)` using `DefaultBorder()` or `SingleLineBorder()`
- Sizing uses `SetPreferredWidth`/`SetPreferredHeight` (user-facing); `SetSize` is used internally by the layout engine
- All components have ID, size, position, visibility, enabled state, and focus management
- Containers manage children with parent back-references

**`input/`** — Focus and key binding management:
- `FocusManager` walks the component tree to find focusable leaves, handles Tab/Shift-Tab cycling, click-to-focus via hit testing, and ID-based focus
- `KeyBindingResolver` resolves global key bindings by walking the component tree

**`window/`** — Screen-level management:
- `BaseWindow` provides shared container + event routing (keyboard, mouse, focus cycling)
- `BaseWindow` contains an internal `Panel` with configurable layout (default `TCB`)
- Children are added with `win.Add(child, Position)` which delegates to the content panel
- `MainWindow` (`NewWindow(id, layout...)`) is the main application window — entry point, defaults to TCB
- `PopupWindow` (`NewPopupWindow`) adds overlay rendering, Escape-to-close, and OnResult callback
- `App` (`NewApp`) bridges to Bubble Tea's Model interface
- Stack is internal — manages popup push/pop lifecycle, focus restoration

## Key Patterns

- **Component interface** (`widget/component.go`): Central contract — `View()`, `HandleEvent()`, `SetSize()`, `Focusable()`, `Focused()`, `PreferredWidth()`, `PreferredHeight()`, etc.
- **Event flow**: `BaseWindow.Update()` routes keyboard events to the focused leaf's `HandleEvent()`, mouse events go through `FocusManager.HitTest()`
- **Composition over inheritance**: Containers hold `[]Component` children; `FocusManager.Leaves()` recursively flattens the tree to find focusable components
- **Styling**: All components use lipgloss for terminal rendering; leaf styles passed via constructor config structs (e.g., `DefaultButtonStyles()`); Panel borders via `SetBorder(BorderConfig)`
- **Testing**: Standard Go `testing` package with `stripansi` for assertion on rendered output; see `widget/fullflow_test.go` for integration test patterns
