# Windows and Popups

## Window

`MainWindow` is the entry point for your application. Create one with `NewWindow`, add children to it, and pass it to `NewApp`.

`NewWindow` defaults to `TCB` layout, which makes it easy to build apps with a tab bar on top, a main content area that fills the remaining space, and an optional status bar at the bottom.

```go
win := window.NewWindow("main")

tabs := widget.NewTabComponent("tabs", widget.DefaultTabBarStyles())
tabs.AddTab("Overview", overviewPanel)
tabs.AddTab("Settings", settingsPanel)
statusBar := widget.NewText("status", "Ready", lipgloss.NewStyle())

win.Add(tabs, widget.TCBCenter)
win.Add(statusBar, widget.TCBBottom)

win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
    if msg.String() == "ctrl+q" {
        return tea.Quit
    }
    return nil
})
```

For a simpler layout, pass a layout explicitly:

```go
win := window.NewWindow("main", widget.Vertical)
panel := widget.NewPanel("settings")
panel.SetBorder(widget.DefaultBorder())
panel.Add(field1, widget.Next)
panel.Add(submitBtn, widget.Next)
win.Add(panel, widget.Next)
```

`OnKeyPress` registers a fallthrough handler for keys that are not consumed by a registered key binding or the focused leaf. It is the right place for global application-level keys like quit.

## PopupWindow

`PopupWindow` is an overlay popup with a title border, Escape-to-close, and an `OnResult` callback.

```go
popup := window.NewPopupWindow("confirm", "Confirm Delete", window.DefaultPopupStyles())
msg := widget.NewText("msg", "Delete all items?", lipgloss.NewStyle())
popup.Add(msg, widget.Next)
popup.OnResult(func(result any) tea.Cmd {
    if result != nil {
        return deleteItems()
    }
    return nil
})
```

### Opening a Popup

Call `ShowPopup` from any callback that has access to the window:

```go
deleteBtn.OnPress(func() tea.Cmd {
    return win.ShowPopup(popup)
})
```

### Closing a Popup

Call `Close` from inside the popup, passing the result value:

```go
yesBtn.OnPress(func() tea.Cmd {
    return popup.Close(true)
})
noBtn.OnPress(func() tea.Cmd {
    return popup.Close(nil)  // nil = cancelled
})
```

Pressing Escape automatically closes the popup with a `nil` result.

### Full Popup Example

```go
popup := window.NewPopupWindow("confirm", "Confirm Delete", window.DefaultPopupStyles())

msg := widget.NewText("msg", "Delete all items?", lipgloss.NewStyle())

yesBtn := widget.NewButton("yes", "Yes", widget.DefaultPopupButtonStyles())
noBtn := widget.NewButton("no", "No", widget.DefaultPopupButtonStyles())

buttons := widget.NewPanel("buttons", widget.Horizontal)
buttons.SetSpacing(2)
buttons.Add(yesBtn, widget.Next)
buttons.Add(noBtn, widget.Next)

popup.Add(msg, widget.Next)
popup.Add(buttons, widget.Next)

yesBtn.OnPress(func() tea.Cmd {
    return popup.Close(true)
})
noBtn.OnPress(func() tea.Cmd {
    return popup.Close(nil)
})

popup.OnResult(func(result any) tea.Cmd {
    if confirmed, ok := result.(bool); ok && confirmed {
        return doDelete()
    }
    return nil
})

// Open it:
deleteBtn.OnPress(func() tea.Cmd {
    return win.ShowPopup(popup)
})
```

## App

`App` bridges `MainWindow` to Bubble Tea's `tea.Model` interface. Create one and pass it to `tea.NewProgram`:

```go
app := window.NewApp(win)
p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion())
p.Run()
```

`App` handles `tea.WindowSizeMsg` internally, so you do not need to manage screen dimensions yourself.

## Event Flow

`BaseWindow` (embedded by both `MainWindow` and `PopupWindow`) routes events in this order:

1. **Tab / Shift-Tab** — handled by `FocusManager` for focus cycling
2. **Registered key bindings** — resolved by `KeyBindingResolver` (children-first, window last)
3. **Focused leaf's `Update()`** — the focused widget handles the key
4. **`OnKeyPress` callback** — fallthrough for keys not consumed by steps 1–3

Mouse events are routed through hit testing: the component under the cursor receives a `MouseClickEvent`, and focus changes if the target is focusable.

## Popup Lifecycle

1. `win.ShowPopup(popup)` — pushes the popup, disables the base window, re-roots `FocusManager` to the popup, focuses its first leaf
2. All events route to the popup while it is on top
3. `popup.Close(result)` — pops the popup, re-enables the base window, restores focus, calls the `OnResult` callback with `result`
4. Pressing Escape automatically calls `popup.Close(nil)`

## Multiple Popups

You can push popups from inside a popup. Each push deactivates the previous window:

```go
// From inside popup1:
popup2 := window.NewPopupWindow("nested", "Are you sure?", window.DefaultPopupStyles())
confirmBtn.OnPress(func() tea.Cmd {
    return popup1.ShowPopup(popup2)
})
```

Closing restores the previous window. The base window can never be closed this way.

## Window-Level Key Bindings

Use `RegisterKeyBinding` (inherited from `BaseComponent` through `BaseContainer`) for keys that should trigger actions regardless of which leaf has focus. These are resolved in step 2 of the event flow, before the focused leaf sees the key:

```go
win := window.NewWindow("main")

win.RegisterKeyBinding("ctrl+n", "New", func() tea.Cmd {
    return createItem()
})
```

For application-level fallthrough keys (like quit), use `OnKeyPress` instead — it runs after the focused leaf has had a chance to handle the key.

## Reacting to State Changes

Use `OnUpdate` to react after any event changes state (focus, key bindings, etc.). This is the right place to update status bars or other reactive UI:

```go
win.OnUpdate(func() tea.Cmd {
    // Build status hints from active key bindings
    var parts []string
    parts = append(parts, "Tab: focus")
    for _, b := range win.ActiveKeyBindings() {
        h := b.Help()
        if h.Key != "" && h.Desc != "" {
            parts = append(parts, h.Key+": "+h.Desc)
        }
    }
    statusBar.SetText(strings.Join(parts, " | "))
    return nil
})
```

`ActiveKeyBindings()` returns all registered key bindings from visible, active components in the current window's tree. Bindings from disabled or hidden components are excluded.
