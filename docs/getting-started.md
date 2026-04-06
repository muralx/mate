# Getting Started

## Installation

```bash
go get github.com/muralx/mate
```

Mate depends on:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — the TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — key binding types
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling

## How Mate Differs from Bubble Tea

Bubble Tea follows the Elm Architecture: you write a Model, an Update function, and a View function. As your UI grows, you route events manually, track focus yourself, and wire up component interactions inside Update.

Mate replaces that with a component tree. You create components, add them to containers, and set callbacks. The framework handles event routing, focus cycling, keyboard dispatch, and rendering. You still use `tea.Cmd` for side effects and `tea.Program` to run the app — Mate uses Bubble Tea as its runtime, not as a pattern to follow.

If you know Bubble Tea: think of Mate as what happens when you stop writing Update/View and start composing components instead.

## Core Concepts

### Component Tree

Every Mate application is a tree of components. There are three kinds:

| Kind | Can receive focus? | Has children? | Examples |
|------|-------------------|---------------|----------|
| **Leaf** | Yes | No | Button, TextInput, Toggle, CheckboxList, TabBar, Table |
| **Container** | No | Yes | Panel, Field |
| **Display** | No | No | Text, Card |

Containers hold children. Leaves handle keyboard input. Display components just render.

### Layouts

`Panel` (and `Window`) supports three layout strategies:

- **Vertical** — stacks children top-to-bottom at their natural/preferred sizes. No flexing — remaining space is unused. Use for forms and stacked content.
- **Horizontal** — stacks children left-to-right at their natural/preferred sizes (use `SetSpacing` for gaps). No flexing — remaining space is unused. Use for toolbars and button rows.
- **TCB** — Top-Center-Bottom: top and bottom use natural/preferred sizes; **center gets all remaining space**. This is the only layout that stretches a component. Use whenever you need a component to fill available height.

Use `Add(child, Position)` to place children. `Next` appends sequentially for all layouts. For TCB, use `TCBTop`, `TCBCenter`, `TCBBottom` to place components in specific slots.

**Important:** Only TCB stretches. If you need a component to fill the remaining space in a container — a Table that should grow to fill a panel, for example — use TCB and put it in the center slot.

### Preferred Sizes

Use `SetPreferredWidth(w)` and `SetPreferredHeight(h)` to give a component a fixed size hint. The layout engine uses this instead of measuring the component's natural rendered size. Leave unset (the default) to let the layout measure naturally.

### The Component Interface

Every component implements `widget.Component`:

```go
type Component interface {
    ID() string
    View() string
    SetSize(w, h int)
    Size() (w, h int)
    SetPosition(x, y int)
    Position() (x, y int)
    Visible() bool
    SetVisible(bool)
    Enabled() bool
    SetEnabled(bool)
    Active() bool
    Focusable() bool
    Focused() bool
    SetFocused(bool) tea.Cmd
    Parent() Container
    SetParent(Container)
    KeyBindings() []key.Binding
    RegisterKeyBinding(key.Binding, func() tea.Cmd)
    ResolveKeyBinding(tea.KeyMsg) (func() tea.Cmd, bool)
    HandleEvent(event Event) (tea.Cmd, bool)
    PreferredWidth() int
    PreferredHeight() int
    SetPreferredWidth(int)
    SetPreferredHeight(int)
}
```

You rarely implement this yourself. Instead, you compose your UI from the provided widgets.

### Active vs Enabled vs Visible

- **Visible** — whether the component renders. Invisible components are skipped in layout.
- **Enabled** — whether the component is locally enabled. Set with `SetEnabled(bool)`.
- **Active** — whether the component is *effectively* enabled. A component is active only if it is enabled AND all its ancestors are enabled. Disabling a Panel disables everything inside it.

```go
panel.SetEnabled(false)
// Now panel.Active() == false
// And every child, grandchild, etc. has Active() == false
// They remain individually Enabled, but are not Active
```

### Focus

Only one leaf component has focus at a time. The focused component receives keyboard input. Focus is managed by `input.FocusManager`, which walks the component tree to find focusable, active leaves.

Tab and Shift-Tab cycle focus. Mouse clicks change focus. You can also focus programmatically.

### Event Flow

Mate has two separate event paths:

1. **Keyboard events** are routed by `BaseWindow`:
   - Tab/Shift-Tab handled by FocusManager for focus cycling
   - Registered key bindings checked via KeyBindingResolver (children-first)
   - Remaining keys sent to the focused leaf's `Update()`
   - `OnKeyPress` callback fires for keys not consumed by any of the above

2. **Mouse events** flow through hit testing:
   - Hit testing finds the target component
   - Focus changes if it's a left-click on a focusable component
   - `HandleEvent(MouseClickEvent{})` dispatched to the target, bubbles up parents

### Windows and Popups

`MainWindow` is a full-screen component tree. `PopupWindow` is an overlay that sits on top of the main window. The internal `Stack` manages focus and rendering — you interact with it via `win.ShowPopup(popup)` and `popup.Close(result)`. `App` wraps a `MainWindow` and implements `tea.Model`.

## Building Your First Application

### Step 1: Create a Window

Every Mate application starts with a `MainWindow`. `NewWindow` defaults to TCB layout:

```go
win := window.NewWindow("my-window")

// Register a fallthrough key handler for application-level shortcuts
win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
    if msg.String() == "ctrl+c" {
        return tea.Quit
    }
    return nil
})
```

### Step 2: Build a Component Tree

Add widgets to the window using `Add(child, position)`:

```go
win := window.NewWindow("my-window")

// A panel with a text input and a button
panel := widget.NewPanel("form")
panel.SetBorder(widget.DefaultBorder())

nameInput := widget.NewTextInput("name", 30)
nameInput.WithPlaceholder("Type here...")
nameField := widget.NewField("name-field", "Name", nameInput, widget.DefaultFieldStyles())
panel.Add(nameField, widget.Next)

btn := widget.NewButton("ok", "OK", widget.DefaultButtonStyles())
btn.OnPress(func() tea.Cmd {
    return tea.Quit
})
panel.Add(btn, widget.Next)

// TCB: place the form in the center so it expands to fill available space
win.Add(panel, widget.TCBCenter)
```

### Step 3: Wire Into Bubble Tea

Create an `App` with your window and pass it to `tea.NewProgram`:

```go
func main() {
    win := window.NewWindow("my-window")
    // ... build component tree ...

    app := window.NewApp(win)
    tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run()
}
```

`NewApp`:
- Creates a `FocusManager` and focuses the first leaf automatically
- Handles `tea.WindowSizeMsg` and sizes the window to fill the terminal
- Routes all messages through the internal stack (base window + any popups)
- Implements `tea.Model` so you do not need to write a model struct yourself

### Step 4: Add Interactivity

Respond to user actions with callbacks:

```go
nameInput.OnSubmit(func(value string) tea.Cmd {
    // Enter was pressed in the text input
    return nil
})

nameInput.OnChange(func(value string) tea.Cmd {
    // Text changed
    return nil
})

btn.OnPress(func() tea.Cmd {
    // Button was pressed (space, enter, or mouse click)
    return nil
})
```

## Next Steps

- [Components](components.md) — Learn about every available widget
- [Layout](layout.md) — Build complex layouts with containers
- [Focus and Keyboard](focus-and-keyboard.md) — Understand the keyboard event flow
- [Windows and Popups](windows-and-popups.md) — Add popup dialogs
