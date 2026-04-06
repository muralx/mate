# Focus and Keyboard

## Focus Model

Only one leaf component has focus at a time. The focused leaf receives keyboard input. Focus is managed by `input.FocusManager`, which the `Stack` creates and provides to your window's `Update()` method.

### Which Components Can Receive Focus?

A leaf can receive focus if all three conditions are met:
1. It implements the `Leaf` interface (all interactive widgets do)
2. `Focusable()` returns true (all leaves return true; containers return false)
3. `Active()` returns true (the component and all its ancestors are enabled)

### Focus Order

FocusManager walks the component tree depth-first to collect focusable, active leaves. The order matches the order components were added to their containers:

```go
panel.Add(field1, widget.Next)    // field1's input is leaf 0
panel.Add(field2, widget.Next)    // field2's input is leaf 1
panel.Add(submitBtn, widget.Next) // submitBtn is leaf 2
```

Tab cycles through leaves in this order. Shift-Tab cycles in reverse.

## Keyboard Event Flow

When a key arrives at your window's `Update()`, delegate to `BaseWindow.HandleKey()`. The flow is:

```
Key arrives at BaseWindow.HandleKey(msg, fm):

1. Tab / Shift-Tab
   -> FocusManager.Next() / Prev()
   -> Returns tea.Cmd (may include cursor blink start)
   -> DONE

2. KeyBindingResolver.Resolve(msg)
   -> Walks entire component tree looking for registered bindings
   -> If match found on a focusable component: focus it first, then call action
   -> If match found on a non-focusable component: call action without focus change
   -> Returns tea.Cmd
   -> DONE

3. FocusedLeaf.Update(msg)
   -> The focused leaf handles the key internally
   -> Component may call its OnKeyPress callback for unconsumed keys
   -> Returns (tea.Cmd, consumed bool)
   -> DONE
```

### Window-Level Keys

Handle window-specific keys (Esc, Ctrl+C, etc.) in your window's `Update()` **before** calling `HandleKey()`:

```go
func (w *MyWindow) Update(msg tea.Msg, fm *input.FocusManager) (tea.Cmd, *window.Result) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Window-level shortcuts first
        switch msg.String() {
        case "ctrl+c":
            return tea.Quit, nil
        case "esc":
            return nil, &window.Result{Value: "cancelled"}
        }
        // Standard routing
        cmd, _ := w.HandleKey(msg, fm)
        return cmd, nil
    }
    return nil, nil
}
```

## Global Key Bindings

Register key bindings on any component. The KeyBindingResolver checks these before routing to the focused leaf.

### RegisterKeyBinding

Register a custom action for a key on any component:

```go
// On a window (BaseWindow embeds BaseContainer which embeds BaseComponent)
w.RegisterKeyBinding(
    key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("ctrl+x", "Quit")),
    func() tea.Cmd { return tea.Quit },
)

// On any component
panel.RegisterKeyBinding(
    key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "New")),
    func() tea.Cmd { return createNew() },
)
```

### BindDefaultActionToKey

For interactive components, bind a global shortcut to the component's default action:

```go
// Button: shortcut triggers OnPress
saveBtn.OnPress(func() tea.Cmd { return save() })
saveBtn.BindDefaultActionToKey(key.NewBinding(
    key.WithKeys("ctrl+s"),
    key.WithHelp("ctrl+s", "Save"),
))

// Toggle: shortcut toggles on/off
toggle.BindDefaultActionToKey(key.NewBinding(
    key.WithKeys("ctrl+g"),
    key.WithHelp("ctrl+g", "Toggle option"),
))

// CheckboxList: shortcut toggles the item at cursor
nodeList.BindDefaultActionToKey(key.NewBinding(
    key.WithKeys("ctrl+a"),
    key.WithHelp("ctrl+a", "Toggle all"),
))

// TabBar: shortcut activates the tab at cursor
tabs.BindDefaultActionToKey(key.NewBinding(
    key.WithKeys("ctrl+t"),
    key.WithHelp("ctrl+t", "Switch tab"),
))
```

`BindDefaultActionToKey` is available on: **Button**, **Toggle**, **CheckboxList**, **TabBar**.

When the shortcut is pressed:
1. If the component is focusable and active, it receives focus first
2. Then the default action fires (same as pressing Space/Enter on the component)

### RemoveKeyBinding

Remove a previously registered key binding:

```go
binding := key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "New"))
panel.RegisterKeyBinding(binding, func() tea.Cmd { return createNew() })

// Later, remove it
panel.RemoveKeyBinding(binding)
```

Matches by key combination (the `Keys()` on the binding). No-op if the binding is not found.

### TabBar Per-Tab Accelerators

Bind keyboard shortcuts to activate specific tabs by index:

```go
tabs := widget.NewTabBar("tabs", []string{"Overview", "Details", "Settings"}, widget.DefaultTabBarStyles())

tabs.SetTabKeyBinding(0, "ctrl+d")                // help desc defaults to "Overview"
tabs.SetTabKeyBinding(1, "ctrl+e")                // help desc defaults to "Details"
tabs.SetTabKeyBinding(2, "ctrl+g", "Settings")    // explicit help description
```

When the shortcut fires, the tab is activated directly (no cursor movement needed). `OnChange` fires as normal. If the tab is already active, it's a no-op.

Calling `SetTabKeyBinding` again on the same index replaces the previous binding. Panics if the index is out of range.

### Binding Resolution Rules

- Only visible and active components' bindings are checked
- Children are checked before their parents (most-specific wins)
- Root container bindings are checked last (natural place for app-level shortcuts like Quit, Help)
- The resolver checks `ResolveKeyBinding()` on each component, which looks at the component's registered bindings
- Tab/Shift-Tab are handled before binding resolution (they always cycle focus)

### key.Binding

Mate uses `key.Binding` from `github.com/charmbracelet/bubbles/key`:

```go
import "github.com/charmbracelet/bubbles/key"

// Multiple keys for one binding
binding := key.NewBinding(
    key.WithKeys("ctrl+s", "f5"),
    key.WithHelp("ctrl+s/f5", "Save"),
)

// Disable a binding at runtime
binding.SetEnabled(false)

// Check help text
h := binding.Help()
fmt.Println(h.Key, h.Desc) // "ctrl+s/f5", "Save"
```

## OnKeyPress Callback

Each focusable component has an `OnKeyPress` callback that fires for keys the component doesn't handle internally:

```go
nameInput := widget.NewTextInput("name", 30)
nameInput.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
    // TextInput consumes all keys, so this rarely fires
    // But for other components:
    return nil
})

btn := widget.NewButton("ok", "OK", widget.DefaultButtonStyles())
btn.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
    // Fires for any key that isn't Space or Enter
    if msg.String() == "?" {
        return showHelp()
    }
    return nil
})
```

OnKeyPress is called inside the component's `Update()` method after the component's own key handling. If OnKeyPress returns a non-nil `tea.Cmd`, the key is considered consumed.

## Focus Management API

The `FocusManager` is provided to your window's `Update()` method. You can use it for programmatic focus control.

### Cycling Focus

```go
cmd := fm.Next()  // move to next leaf (Tab)
cmd := fm.Prev()  // move to previous leaf (Shift-Tab)
```

Both return `tea.Cmd` — always batch these into your update return.

### Focus by ID

```go
ok, cmd := fm.FocusByID("email")
if !ok {
    // Component not found or not focusable/active
}
```

### Focus First

```go
cmd := fm.FocusFirst()  // focus the first available leaf
```

### Query Focus State

```go
leaf := fm.FocusedLeaf()  // currently focused leaf, or nil
if leaf != nil {
    fmt.Println("Focused:", leaf.ID())
}
```

### Get Focused Component's Registered Bindings

```go
bindings := fm.FocusedKeyBindings()  // []key.Binding registered on the focused leaf
for _, b := range bindings {
    h := b.Help()
    fmt.Printf("%s: %s  ", h.Key, h.Desc)
}
```

Only returns bindings from `RegisterKeyBinding` / `BindDefaultActionToKey`. Internal widget keys (space/enter on Button, up/down on Table) are handled in `Update()` and are not exposed through `KeyBindings()`.

### Get All Active Key Bindings

```go
bindings := fm.AllActiveKeyBindings()  // registered bindings from all visible, active components
for _, b := range bindings {
    h := b.Help()
    fmt.Printf("%s: %s  ", h.Key, h.Desc)
}
```

`AllActiveKeyBindings()` walks the component tree depth-first, collecting registered bindings from every visible and active component. Hidden or disabled components (and their subtrees) are skipped. Duplicate key combinations are deduplicated (first-match wins, matching `KeyBindingResolver.Resolve` behavior).

This is designed for building status bars that show global shortcuts:
```
ctrl+q: Quit | ctrl+n: New item | ctrl+r: Refresh
```

## Cmd Threading

All focus operations return `tea.Cmd`. This is important because some components (like TextInput) return a command from `SetFocused(true)` that starts the cursor blink timer. If you discard these commands, the cursor won't blink.

```go
// Good: return the cmd from HandleKey
cmd, _ := w.HandleKey(msg, fm)
return cmd, nil

// Bad: discard the cmd
w.HandleKey(msg, fm)
return nil, nil  // cursor blink may not start
```

The same applies to `fm.Next()`, `fm.Prev()`, `fm.FocusByID()`, and `fm.FocusFirst()` — always return their commands.
