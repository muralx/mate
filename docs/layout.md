# Layout

Mate uses Bubble Tea's convention: layout happens in `View()`. There is no separate layout pass. Containers compute child positions during rendering because that's the only point where they know their own dimensions.

## Layout Types

Panel (and Window) support three layout strategies:

| Layout | Behavior |
|--------|----------|
| `Vertical` | Stacks children top-to-bottom. Each child gets the container's full content width and its preferred/natural height. **No flexing** — remaining space below the last child is unused. |
| `Horizontal` | Stacks children left-to-right. Each child gets the container's full content height and its preferred/natural width. **No flexing** — remaining space to the right is unused. |
| `TCB` | Top-Center-Bottom. Top and Bottom get their preferred/natural size. **Center gets ALL remaining space** — this is the only layout that stretches a component. |

> **Key rule: Only TCB stretches.** If you need a child to fill the available height of a container, use TCB layout and put the component in the center slot. Vertical and Horizontal never stretch their children.

## When to Use Which Layout

| Situation | Layout to use |
|-----------|--------------|
| Form with stacked fields and buttons | `Vertical` |
| Toolbar or row of buttons | `Horizontal` |
| Cards displayed side by side | `Horizontal` |
| Main content area that must fill remaining height | `TCB` (put it in center) |
| Window with tab bar at top and content filling the rest | `TCB` |
| Bordered panel wrapping a Table | `TCB` with the Table as center (not Vertical — see note below) |
| Dashboard: cards top, table center, detail bottom | `TCB` |

**Note on bordered panels around a Table:** Use TCB layout, not Vertical. With Vertical layout, the Panel measures the Table's natural height, which for an empty Table is just 1 row. With TCB, the Table is in the center slot and gets all remaining space regardless of its natural height.

## Containers

### Panel

Panel is the universal container. It has a configurable layout, an optional border, an optional title, and configurable spacing.

```go
// Vertical layout (default) — form with stacked fields
panel := widget.NewPanel("settings")
panel.SetBorder(widget.DefaultBorder())
panel.SetTitle("Settings")

panel.Add(field1, widget.Next)
panel.Add(field2, widget.Next)
panel.Add(submitBtn, widget.Next)
```

```go
// Horizontal layout — toolbar with buttons
toolbar := widget.NewPanel("toolbar", widget.Horizontal)
toolbar.SetSpacing(2)

toolbar.Add(saveBtn, widget.Next)
toolbar.Add(cancelBtn, widget.Next)
toolbar.Add(statusText, widget.Next)
```

```go
// TCB layout — center fills remaining height
win := window.NewWindow("main") // Window defaults to TCB

tabBar := widget.NewTabBar("tabs", []string{"Overview", "Settings"}, widget.DefaultTabBarStyles())
contentPanel := widget.NewPanel("content")
statusBar := widget.NewText("status", "Ready", lipgloss.NewStyle())

win.Add(tabBar, widget.TCBTop)
win.Add(contentPanel, widget.TCBCenter)
win.Add(statusBar, widget.TCBBottom)
```

Panel renders children according to its layout. When any descendant is focused and a border is set, the border color switches to its active color.

**Positioning and sizing:** Panel automatically sets each child's position and size during `View()`. You do not need to manually position children inside a Panel.

### Add Method and Position

Use `Add(child, Position)` to place children:

```go
panel.Add(child, widget.Next)        // append sequentially (Vertical/Horizontal/TCB)
panel.Add(child, widget.TCBTop)      // TCB only: place in top slot
panel.Add(child, widget.TCBCenter)   // TCB only: place in center slot
panel.Add(child, widget.TCBBottom)   // TCB only: place in bottom slot
```

For `Vertical` and `Horizontal` layouts, only `Next` is valid. For `TCB`, you can use `Next` to fill slots in order (Top → Center → Bottom), or specify the slot explicitly.

### SetBorder

```go
// Use the default border (rounded, blue/cyan on focus)
panel.SetBorder(widget.DefaultBorder())

// Or a custom border
panel.SetBorder(widget.SingleLineBorder("#444444", "#4fc3f7"))
```

See `BorderConfig` in the [API Reference](api-reference.md) for the full type definition.

### Preferred Sizes

Use `SetPreferredWidth` and `SetPreferredHeight` to give a component a fixed size hint. The layout engine uses this instead of measuring the component's natural size:

```go
sidebar := widget.NewPanel("sidebar")
sidebar.SetPreferredWidth(30)
// sidebar will always be 30 columns wide in a Horizontal parent

header := widget.NewPanel("header")
header.SetPreferredHeight(3)
// header will always be 3 rows tall in a Vertical or TCB parent
```

If no preferred size is set (the default), the layout engine measures the component's natural rendered size.

`SetSize` is used internally by the layout engine — you do not call it directly except when building custom containers.

### Field

A horizontal composition of label, separator, and input component. The label automatically highlights when any child has focus.

```go
nameInput := widget.NewTextInput("name", 30)
nameField := widget.NewField("name-field", "Name", nameInput, widget.DefaultFieldStyles())
```

Renders as: `Name: [text input here]`

Field uses Horizontal layout internally. It manages three children:
1. A `Text` component for the label
2. A `Text` component for the separator (": ")
3. The input component you provide

You can add more children after construction:

```go
// Add a popup button after the input
popupBtn := widget.NewButton("name-popup", "[▾]", widget.DefaultPopupButtonStyles())
nameField.AddChild(popupBtn)
// Renders: Name: [text input here][▾]
```

#### FieldStyles

```go
type FieldStyles struct {
    Label     lipgloss.Style  // label when no child is focused
    LabelHot  lipgloss.Style  // label when a child has focus (highlighted)
    Separator lipgloss.Style  // separator style (constant)
}
```

**Positioning:** Field sets child positions horizontally during `View()`, measuring each child's rendered width and placing the next one to the right. This enables correct mouse hit testing across all children.

#### Accessing Field Parts

```go
field.Label()     // *Text — the label component
field.Separator() // *Text — the separator component
field.Input()     // Component — the input you provided
```

## Common Patterns

### Bordered Panel Around a Table

Use TCB layout — not Vertical. With Vertical, the Panel measures the Table's natural height (1 row for an empty table). With TCB, the Table expands to fill all available space.

```go
tablePanel := widget.NewPanel("table-panel", widget.TCB)
tablePanel.SetBorder(widget.DefaultBorder())
tablePanel.SetTitle("Results")

table := widget.NewTable("results", columns, ds, widget.DefaultTableStyles())
tablePanel.Add(table, widget.TCBCenter)  // Table fills all remaining height
```

### Dashboard Layout (TCB: cards top, table center, detail bottom)

```go
win := window.NewWindow("dashboard") // defaults to TCB

// Top: a row of summary cards
cardsRow := widget.NewPanel("cards", widget.Horizontal)
cardsRow.SetSpacing(2)
cardsRow.Add(widget.NewCard("cpu", "CPU", "42%", widget.DefaultCardStyles()), widget.Next)
cardsRow.Add(widget.NewCard("mem", "Memory", "3.1 GB", widget.DefaultCardStyles()), widget.Next)
cardsRow.Add(widget.NewCard("err", "Errors", "0", widget.DefaultCardStyles()), widget.Next)

// Center: data table fills remaining height
tablePanel := widget.NewPanel("table-panel", widget.TCB)
tablePanel.SetBorder(widget.DefaultBorder())
tablePanel.SetTitle("Events")
table := widget.NewTable("events", columns, ds, widget.DefaultTableStyles())
tablePanel.Add(table, widget.TCBCenter)

// Bottom: detail / status line
detailText := widget.NewText("detail", "Select a row to view details", lipgloss.NewStyle())

win.Add(cardsRow, widget.TCBTop)
win.Add(tablePanel, widget.TCBCenter)
win.Add(detailText, widget.TCBBottom)
```

### Form Layout (Vertical: fields stacked)

```go
form := widget.NewPanel("form", widget.Vertical)
form.SetBorder(widget.DefaultBorder())
form.SetTitle("New User")

form.Add(widget.NewField("name-f", "Name",  widget.NewTextInput("name", 30), widget.DefaultFieldStyles()), widget.Next)
form.Add(widget.NewField("email-f", "Email", widget.NewTextInput("email", 30), widget.DefaultFieldStyles()), widget.Next)
form.Add(widget.NewField("role-f", "Role",  roleToggle, widget.DefaultFieldStyles()), widget.Next)

buttons := widget.NewPanel("buttons", widget.Horizontal)
buttons.SetSpacing(2)
buttons.Add(submitBtn, widget.Next)
buttons.Add(cancelBtn, widget.Next)
form.Add(buttons, widget.Next)
```

### Toolbar (Horizontal: buttons with spacing)

```go
toolbar := widget.NewPanel("toolbar", widget.Horizontal)
toolbar.SetSpacing(2)

toolbar.Add(newBtn, widget.Next)
toolbar.Add(editBtn, widget.Next)
toolbar.Add(deleteBtn, widget.Next)
```

## Building Component Trees

### Parent-Child Relationships

When you call `Add()`, the child's parent is set automatically. This enables:
- `Active()` propagation — disabling a parent makes all descendants inactive
- `InnerFocused()` — containers know when any descendant has focus
- Event bubbling — mouse events bubble up the parent chain

```go
panel := widget.NewPanel("panel")
panel.SetBorder(widget.DefaultBorder())
field := widget.NewField("field", "Name", input, widget.DefaultFieldStyles())
panel.Add(field, widget.Next)

// Now:
// field.Parent() == panel
// input.Parent() == field
```

### Nesting

Containers can be nested to any depth:

```go
outerPanel := widget.NewPanel("outer")
outerPanel.SetBorder(widget.DefaultBorder())

innerPanel1 := widget.NewPanel("inner1")
innerPanel1.SetBorder(widget.DefaultBorder())
innerPanel1.SetTitle("Connection")
innerPanel1.Add(hostField, widget.Next)
innerPanel1.Add(portField, widget.Next)

innerPanel2 := widget.NewPanel("inner2")
innerPanel2.SetBorder(widget.DefaultBorder())
innerPanel2.SetTitle("Authentication")
innerPanel2.Add(userField, widget.Next)
innerPanel2.Add(passField, widget.Next)

outerPanel.Add(innerPanel1, widget.Next)
outerPanel.Add(innerPanel2, widget.Next)
outerPanel.Add(connectBtn, widget.Next)
```

Focus, active state, and mouse events all propagate correctly through any nesting depth.

### Visibility

Hide components with `SetVisible(false)`. Hidden components:
- Are not rendered by their parent's `View()`
- Are skipped by the FocusManager (their leaves cannot receive focus)
- Do not receive events

```go
advancedPanel.SetVisible(false) // hide the panel and all its children

// Later...
advancedPanel.SetVisible(true)  // show it again
```

### Enabling/Disabling

Disable components with `SetEnabled(false)`. Disabled components:
- Render with a faint/dimmed style
- Cannot receive focus (FocusManager skips inactive leaves)
- Do not respond to keyboard or mouse input
- Propagate disabled state to all descendants via `Active()`

```go
// Disable a specific field
notesField.SetEnabled(false)

// Disable an entire section
settingsPanel.SetEnabled(false) // everything inside becomes inactive
```

A component is `Active()` only when it is `Enabled()` AND all its ancestors are `Enabled()`. This means you can disable a single component or an entire subtree.

## Size and Alignment

### Preferred vs Computed Sizes

There are two ways a component gets its size:

- **Preferred size** — set by you with `SetPreferredWidth(w)` / `SetPreferredHeight(h)`. The layout engine uses this directly.
- **Natural size** — measured by the layout engine by calling `View()` and measuring the output. Used when no preferred size is set.

In TCB layout, the center-slot component always gets all remaining height, regardless of any preferred size it has.

`SetSize(w, h)` is used internally by the layout engine to apply the computed size before final rendering. You should not call it directly.

### Components Must Render at Their Allocated Size

When the layout engine calls `SetSize(w, h)` on a component, the component must render at exactly that size. `RenderInSize()` (available on `BaseComponent`) handles this automatically for leaf components. Custom containers must respect the dimensions set by the layout engine in their own `View()` implementation.

### Alignment

Leaf components support horizontal alignment within their allocated width:

```go
btn.SetPreferredWidth(40)
btn.SetAlignment(widget.AlignCenter)  // centers "[ OK ]" in 40 columns
btn.SetAlignment(widget.AlignRight)   // right-aligns
btn.SetAlignment(widget.AlignLeft)    // left-aligns (default)
```

Available alignments: `AlignLeft`, `AlignCenter`, `AlignRight`.

## Custom Containers

To create a custom container, embed `BaseContainer` and pass `self` to the constructor:

```go
type Sidebar struct {
    widget.BaseContainer
    styles SidebarStyles
}

func NewSidebar(id string, styles SidebarStyles) *Sidebar {
    s := &Sidebar{styles: styles}
    s.BaseContainer = *widget.NewBaseContainer(id, s)
    return s
}

func (s *Sidebar) View() string {
    // Custom layout logic
    var parts []string
    px, py := s.Position()
    yOffset := py
    for _, child := range s.Children() {
        if !child.Visible() { continue }
        child.SetPosition(px, yOffset)
        rendered := child.View()
        parts = append(parts, rendered)
        yOffset += lipgloss.Height(rendered)
    }
    return s.styles.Border.Render(
        lipgloss.JoinVertical(lipgloss.Left, parts...),
    )
}
```

The `self` parameter ensures that when children are added, their `Parent()` returns your concrete type (not `BaseContainer`). This is required for correct `Active()` propagation and event bubbling.

When creating standalone `BaseContainer` instances for testing or simple grouping, pass `nil` as `self` — it defaults to the BaseContainer itself:

```go
group := widget.NewBaseContainer("group", nil)
group.AddChild(btn1)
group.AddChild(btn2)
```
