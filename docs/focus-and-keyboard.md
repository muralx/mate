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

The framework handles all keyboard routing automatically inside `BaseWindow`. The flow is:

```
Key arrives at BaseWindow (internal):

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

4. Window OnKeyPress callback
   -> Fallthrough for keys not consumed by steps 1-3
   -> DONE
```

### Window-Level Keys

Use `RegisterKeyBinding` for global shortcuts that should fire before the focused leaf sees the key. Use `OnKeyPress` for fallthrough keys that should only fire if nothing else consumed them:

```go
win := window.NewWindow("main")

// Global shortcut — fires before focused leaf (step 2)
win.RegisterKeyBinding("ctrl+n", "New", func() tea.Cmd {
    return createItem()
})

// Fallthrough — fires after focused leaf (step 4)
win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
    if msg.String() == "ctrl+c" {
        return tea.Quit
    }
    return nil
})
```

## Global Key Bindings

Register key bindings on any component. The KeyBindingResolver checks these before routing to the focused leaf.

### RegisterKeyBinding

Register a custom action for a key on any component:

```go
// On a window
win.RegisterKeyBinding("ctrl+x", "Quit", func() tea.Cmd {
    return tea.Quit
})

// On any component
panel.RegisterKeyBinding("ctrl+n", "New", func() tea.Cmd {
    return createNew()
})
```

The first argument is the key combination (e.g., `"ctrl+s"`, `"space"`, `"enter"`). The second is a description used for help text display (pass `""` to hide from status bars).

### BindDefaultActionToKey

For interactive components, bind a global shortcut to the component's default action:

```go
// Button: shortcut triggers OnPress
saveBtn.OnPress(func() tea.Cmd { return save() })
saveBtn.BindDefaultActionToKey("ctrl+s", "Save")

// Toggle: shortcut toggles on/off
toggle.BindDefaultActionToKey("ctrl+g", "Toggle option")

// CheckboxList: shortcut toggles the item at cursor
nodeList.BindDefaultActionToKey("ctrl+a", "Toggle all")

// TabBar: shortcut activates the tab at cursor
tabs.BindDefaultActionToKey("ctrl+t", "Switch tab")
```

The description is optional — if omitted, it defaults to the component's label.

`BindDefaultActionToKey` is available on: **Button**, **Toggle**, **CheckboxList**, **TabBar**.

When the shortcut is pressed:
1. If the component is focusable and active, it receives focus first
2. Then the default action fires (same as pressing Space/Enter on the component)

### RemoveKeyBinding

Remove a previously registered key binding. Pass a `key.Binding` that matches the key combination to remove:

```go
import "github.com/charmbracelet/bubbles/key"

panel.RegisterKeyBinding("ctrl+n", "New", func() tea.Cmd { return createNew() })

// Later, remove it by matching the key
panel.RemoveKeyBinding(key.NewBinding(key.WithKeys("ctrl+n")))
```

Matches by key combination. No-op if the binding is not found.

### TabComponent Per-Tab Accelerators

Bind keyboard shortcuts to activate specific tabs by index:

```go
tabs := widget.NewTabComponent("tabs", widget.DefaultTabBarStyles())
tabs.AddTab("Overview", overviewPanel)
tabs.AddTab("Details", detailsPanel)
tabs.AddTab("Settings", settingsPanel)

tabs.SetTabKeyBinding(0, "ctrl+d")                // help desc defaults to "Overview"
tabs.SetTabKeyBinding(1, "ctrl+e")                // help desc defaults to "Details"
tabs.SetTabKeyBinding(2, "ctrl+g", "Settings")    // explicit help description
```

When the shortcut fires, the tab activates and the corresponding content panel is shown. If the tab is already active, it's a no-op.

Calling `SetTabKeyBinding` again on the same index replaces the previous binding. Panics if the index is out of range.

### Binding Resolution Rules

- Only visible and active components' bindings are checked
- Children are checked before their parents (most-specific wins)
- Root container bindings are checked last (natural place for app-level shortcuts like Quit, Help)
- The resolver checks `ResolveKeyBinding()` on each component, which looks at the component's registered bindings
- Tab/Shift-Tab are handled before binding resolution (they always cycle focus)

### key.Binding (internal)

Internally, Mate uses `key.Binding` from `github.com/charmbracelet/bubbles/key` to store bindings. You generally don't need to create `key.Binding` values directly — `RegisterKeyBinding` and `BindDefaultActionToKey` accept simple strings. The main place you'll encounter `key.Binding` is when reading bindings back (e.g., `KeyBindings()`, `AllActiveKeyBindings()`) or when calling `RemoveKeyBinding`.

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

The `FocusManager` is created automatically by `App` and manages focus internally. You rarely need to interact with it directly, but it's useful for understanding how focus works.

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

All focus operations return `tea.Cmd`. This is important because some components (like TextInput) return a command from `SetFocused(true)` that starts the cursor blink timer. The framework handles cmd threading automatically when you use `App` and callbacks. If you use `FocusManager` programmatically (e.g., `fm.FocusByID()` inside a callback), return the resulting `tea.Cmd` from your callback so it gets batched correctly.
