# Mouse

Mate supports mouse interaction through hit testing, click-to-focus, and event dispatch. Mouse support requires enabling mouse events in your Bubble Tea program.

## Enabling Mouse

```go
p := tea.NewProgram(model{}, tea.WithMouseAllMotion())
```

## Mouse Event Flow

When a mouse event arrives at your window's `Update()`, delegate to `BaseWindow.HandleMouse()`:

```go
func (w *MyWindow) Update(msg tea.Msg, fm *input.FocusManager) (tea.Cmd, *window.Result) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        cmd, _ := w.HandleMouse(msg, fm)
        return cmd, nil
    }
    return nil, nil
}
```

The flow inside `HandleMouse`:

```
Mouse event arrives at BaseWindow.HandleMouse(msg, fm):

1. HitTest: find the leaf at (X, Y)
   -> FocusManager walks all focusable+active leaves
   -> Checks bounds: leaf.Position() and leaf.Size()
   -> Returns the topmost matching leaf, or nil
   -> If no match: DONE (nil, false)

2. Focus change (left button press only)
   -> Only on tea.MouseActionPress
   -> If target is focusable and active: blur old leaf, focus new leaf
   -> Returns tea.Cmd from focus change (cursor blink etc.)
   -> If target is not focusable: DONE (nil, false)

3. Dispatch MouseScrollEvent (wheel events)
   -> If button is WheelUp/WheelDown/WheelLeft/WheelRight
   -> target.HandleEvent(MouseScrollEvent{X, Y, Direction})
   -> Direction: -1 for up, +1 for down
   -> DONE (scroll events do not also fire MouseClickEvent)

4. Dispatch MouseClickEvent (button press only)
   -> Only on tea.MouseActionPress (non-wheel)
   -> target.HandleEvent(MouseClickEvent{X, Y, Button})
   -> Event may bubble up the parent chain
   -> Returns (tea.Cmd, consumed)
   -> Motion and release events do NOT dispatch MouseClickEvent
```

## Hit Testing

Hit testing determines which component is at a given screen coordinate. It works because:

1. Containers set their children's positions during `View()` (Panel accounts for border/padding, Field arranges children horizontally)
2. Components store their position via `SetPosition(x, y)` and size via `SetSize(w, h)`
3. FocusManager checks each leaf's bounds during `HitTest(x, y)`

Hit testing walks the leaf list in reverse order (last-added children checked first), so overlapping components resolve to the topmost one.

### Position Requirements

For mouse support to work, containers must set child positions in their `View()` method. The built-in containers (Panel, Field) do this automatically. If you write a custom container, you must set positions:

```go
func (c *MyContainer) View() string {
    px, py := c.Position()
    yOffset := py
    for _, child := range c.Children() {
        if !child.Visible() { continue }
        child.SetPosition(px, yOffset)
        rendered := child.View()
        child.SetSize(lipgloss.Width(rendered), lipgloss.Height(rendered))
        yOffset += lipgloss.Height(rendered)
    }
    // ...
}
```

## MouseClickEvent

When a leaf receives a mouse click, `HandleEvent` is called with a `MouseClickEvent`:

```go
type MouseClickEvent struct {
    X, Y   int              // screen coordinates of the click
    Button tea.MouseButton  // which button was pressed
}
```

### Default Click Behavior

Each component handles clicks differently:

| Component | Click Action |
|-----------|-------------|
| Button | Fires OnPress |
| Toggle | Toggles on/off, fires OnChange |
| CheckboxList | Toggles the clicked item (uses Y coordinate to determine which item) |
| TextInput | Focus only (no specific click action) |
| TabBar | Activates the clicked tab, fires OnChange |
| Table | Moves cursor to the clicked row (accounts for header and scroll offset) |

## MouseScrollEvent

When the mouse wheel is scrolled over a leaf, `HandleEvent` is called with a `MouseScrollEvent`:

```go
type MouseScrollEvent struct {
    X, Y      int  // screen coordinates
    Direction int  // -1 = up, +1 = down
}
```

### Default Scroll Behavior

| Component | Scroll Action |
|-----------|--------------|
| Table | Moves cursor up/down by 3 rows |
| ScrollableText | Scrolls viewport up/down by 3 lines |
| Other components | Not consumed (bubbles to parent) |

Scroll events do not change focus — the component under the cursor receives the event regardless of which component has keyboard focus.

### Event Bubbling

If a component's `HandleEvent` returns `consumed = false`, the event bubbles up to the parent. This continues up the tree until a component consumes it or the root is reached.

```go
// Default BaseComponent.HandleEvent: bubble to parent
func (bc *BaseComponent) HandleEvent(event Event) (tea.Cmd, bool) {
    if bc.parent != nil {
        return bc.parent.HandleEvent(event)
    }
    return nil, false
}
```

This means a Panel or Window can handle mouse events that no child consumed.

## Popup Mouse Coordinates

When a popup is displayed via the Stack, mouse coordinates are automatically adjusted. The Stack tracks the popup's overlay offset and subtracts it from mouse coordinates before routing to the popup window:

```go
// This happens automatically inside Stack.Update():
if mouse, ok := msg.(tea.MouseMsg); ok && s.HasPopup() {
    mouse.X -= s.popupOffset.X
    mouse.Y -= s.popupOffset.Y
    msg = mouse
}
```

Components inside popups receive coordinates relative to the popup's content area, so their hit testing works correctly without any special handling.

## Custom HandleEvent

Override `HandleEvent` on your component to add custom mouse handling:

```go
type ClickableLabel struct {
    widget.BaseComponent
    text    string
    onClick func() tea.Cmd
}

func (cl *ClickableLabel) HandleEvent(event widget.Event) (tea.Cmd, bool) {
    if _, ok := event.(widget.MouseClickEvent); ok {
        if cl.onClick != nil {
            return cl.onClick(), true
        }
        return nil, true
    }
    return cl.BaseComponent.HandleEvent(event)
}
```

Always fall through to `BaseComponent.HandleEvent` for unhandled events to preserve bubbling.
