# API Reference

Complete reference for all exported types, methods, and functions.

## Package `widget`

### Interfaces

#### Component

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

#### Leaf

```go
type Leaf interface {
    Component
    Update(msg tea.KeyMsg) (tea.Cmd, bool)
}
```

#### Container

```go
type Container interface {
    Component
    Children() []Component
    AddChild(child Component)
    InnerFocused() bool
}
```

#### Event

```go
type Event interface{ isEvent() }
```

### Types

#### Layout

```go
type Layout int

const (
    Vertical   Layout = iota // stack children top-to-bottom, natural sizes
    Horizontal               // stack children left-to-right, natural sizes
    TCB                      // Top-Center-Bottom: center flexes to fill remaining space
)
```

#### Position

```go
type Position int

const (
    Next      Position = iota // sequential: append for V/H, fill next slot for TCB
    TCBTop                    // TCB only: place in top slot
    TCBCenter                 // TCB only: place in center slot
    TCBBottom                 // TCB only: place in bottom slot
)
```

#### BorderConfig

```go
type BorderConfig struct {
    Type        BorderType     // NoBorder or RoundedBorder
    Color       lipgloss.Color // border color when no descendant is focused
    ActiveColor lipgloss.Color // border color when a descendant has focus
    Padding     int            // horizontal padding inside border (columns per side)
}
```

Constructor functions:

```go
func SingleLineBorder(color, activeColor string) BorderConfig
func DefaultBorder() BorderConfig
```

`DefaultBorder()` returns a rounded border with a dark blue inactive color and cyan active color.

#### MouseClickEvent

```go
type MouseClickEvent struct {
    X, Y   int
    Button tea.MouseButton
}
```

Implements `Event`. Dispatched to components on mouse button press (not motion or release) via `HandleEvent`.

#### Alignment

```go
type Alignment int

const (
    AlignLeft   Alignment = iota
    AlignCenter
    AlignRight
)
```

#### ToggleMode

```go
type ToggleMode int

const (
    ToggleModeOnOff ToggleMode = iota
    ToggleModeRadio
)
```

#### CheckboxItem

```go
type CheckboxItem struct {
    Label   string
    Value   string
    Checked bool
    IsGroup bool
}
```

#### TableDataSource

```go
type TableDataSource interface {
    RowCount() int
    CellData(row, col int) string
}
```

#### SliceDataSource

```go
func NewSliceDataSource(data [][]string) *SliceDataSource
```

| Method | Signature |
|--------|-----------|
| SetData | `func(data [][]string)` |

#### CellRenderer

```go
type CellRenderer func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string
func DefaultCellRenderer(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string
```

#### ColumnDef

```go
type ColumnDef struct {
    Title    string
    Width    int           // 0 = flex (take remaining)
    Renderer CellRenderer  // nil = DefaultCellRenderer
}
```

### Base Types

#### BaseComponent

```go
func NewBaseComponent(id string) *BaseComponent
```

Methods: `ID`, `View`, `SetSize`, `Size`, `SetPosition`, `Position`, `Visible`, `SetVisible`, `Enabled`, `SetEnabled`, `Active`, `Focusable` (returns false), `Focused`, `SetFocused` (returns nil), `Parent`, `SetParent`, `KeyBindings` (returns only registered bindings), `RegisterKeyBinding`, `RemoveKeyBinding`, `ResolveKeyBinding`, `HandleEvent` (bubbles to parent), `SetAlignment`, `Alignment`, `RenderInSize`, `PreferredWidth`, `PreferredHeight`, `SetPreferredWidth`, `SetPreferredHeight`.

#### FocusableComponent

```go
func NewFocusableComponent(id string) FocusableComponent
```

Embeds `BaseComponent`. Overrides: `Focusable()` returns true.

Additional methods:
- `OnKeyPress(fn func(tea.KeyMsg) tea.Cmd)` — set callback for unconsumed keys

#### BaseContainer

```go
func NewBaseContainer(id string, self Container) *BaseContainer
```

`self` — pass the concrete container type, or nil for the default.

Embeds `BaseComponent`. Overrides: `Focusable()` returns false, `View()` returns empty.

Additional methods: `AddChild(child Component)`, `Children() []Component`, `InnerFocused() bool`.

### Widgets

#### Button

```go
func NewButton(id, label string, styles ButtonStyles) *Button
func DefaultButtonStyles() ButtonStyles
func DefaultPopupButtonStyles() ButtonStyles
```

| Method | Signature |
|--------|-----------|
| OnPress | `func(fn func() tea.Cmd)` |
| BindDefaultActionToKey | `func(binding key.Binding)` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### TextInput

```go
func NewTextInput(id string, inputWidth int) *TextInput
```

| Method | Signature |
|--------|-----------|
| WithPlaceholder | `func(p string) *TextInput` |
| WithCharLimit | `func(n int) *TextInput` |
| OnSubmit | `func(fn func(string) tea.Cmd)` |
| OnChange | `func(fn func(string) tea.Cmd)` |
| SetValue | `func(v string)` |
| Value | `func() string` |
| SetError | `func(e string)` |
| Error | `func() string` |
| SetFocused | `func(v bool) tea.Cmd` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| SetSize | `func(w, h int)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### FormattedTextInput

```go
func NewFormattedTextInput(id string, inputWidth int) *FormattedTextInput
```

Embeds `TextInput`. Inherits all TextInput methods.

| Method | Signature |
|--------|-----------|
| WithValidation | `func(fn func(string) error)` |
| WithFormat | `func(fn func(string) string)` |
| SetFocused | `func(focused bool) tea.Cmd` |

#### Toggle

```go
func NewToggle(id, label string, initial bool, mode ToggleMode, styles ToggleStyles) *Toggle
func DefaultToggleStyles() ToggleStyles
```

| Method | Signature |
|--------|-----------|
| SetLabels | `func(onLabel, offLabel string)` |
| OnChange | `func(fn func(bool) tea.Cmd)` |
| On | `func() bool` |
| SetOn | `func(v bool)` |
| BindDefaultActionToKey | `func(binding key.Binding)` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### CheckboxList

```go
func NewCheckboxList(id string, items []CheckboxItem, styles CheckboxListStyles) *CheckboxList
func DefaultCheckboxListStyles() CheckboxListStyles
```

| Method | Signature |
|--------|-----------|
| OnChange | `func(fn func([]CheckboxItem) tea.Cmd)` |
| Items | `func() []CheckboxItem` |
| Cursor | `func() int` |
| Selected | `func() []string` |
| BindDefaultActionToKey | `func(binding key.Binding)` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### TabBar

```go
func NewTabBar(id string, labels []string, styles TabBarStyles) *TabBar
func DefaultTabBarStyles() TabBarStyles
```

| Method | Signature |
|--------|-----------|
| OnChange | `func(fn func(int) tea.Cmd)` |
| ActiveTab | `func() int` |
| SetActiveTab | `func(i int)` |
| CursorTab | `func() int` |
| SetTabKeyBinding | `func(index int, keys string, description ...string)` |
| BindDefaultActionToKey | `func(binding key.Binding)` |
| SetFocused | `func(v bool) tea.Cmd` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### Table

```go
func NewTable(id string, columns []ColumnDef, ds TableDataSource, styles TableStyles) *Table
func DefaultTableStyles() TableStyles
```

| Method | Signature |
|--------|-----------|
| SetDataSource | `func(ds TableDataSource)` |
| DataSource | `func() TableDataSource` |
| Cursor | `func() int` |
| SetCursor | `func(c int)` |
| OnRowKeyPress | `func(fn func(row int, msg tea.KeyMsg) tea.Cmd)` |
| OnRowClick | `func(fn func(row int) tea.Cmd)` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |
| HandleEvent | `func(event Event) (tea.Cmd, bool)` |

#### Panel

Panel is the universal container with configurable layout.

```go
func NewPanel(id string, layout ...Layout) *Panel
```

`layout` defaults to `Vertical` if omitted.

| Method | Signature |
|--------|-----------|
| Add | `func(child Component, position Position)` |
| SetBorder | `func(config BorderConfig)` |
| SetTitle | `func(title string)` |
| SetTitleStyle | `func(style lipgloss.Style)` |
| SetSpacing | `func(spacing int)` |
| View | `func() string` |

`Add` accepts `Next` for all layouts. For `TCB` layout, also accepts `TCBTop`, `TCBCenter`, `TCBBottom`.
`SetSpacing` sets the column gap between children in `Horizontal` layout.

Embeds `BaseContainer`. Inherits: `AddChild`, `Children`, `InnerFocused`.

#### Card

```go
func NewCard(id, title, value string, styles CardStyles) *Card
func DefaultCardStyles() CardStyles
```

| Method | Signature |
|--------|-----------|
| SetValue | `func(value string)` |
| Value | `func() string` |
| SetAlert | `func(alert bool)` |
| View | `func() string` |

#### Text

```go
func NewText(id, text string, style lipgloss.Style) *Text
```

| Method | Signature |
|--------|-----------|
| SetText | `func(text string)` |
| GetText | `func() string` |
| SetStyle | `func(s lipgloss.Style)` |
| Style | `func() lipgloss.Style` |
| View | `func() string` |

#### Field

```go
func NewField(id, labelText string, input Component, styles FieldStyles) *Field
func DefaultFieldStyles() FieldStyles
```

| Method | Signature |
|--------|-----------|
| Label | `func() *Text` |
| Separator | `func() *Text` |
| Input | `func() Component` |
| View | `func() string` |

Embeds `BaseContainer`. Inherits: `AddChild`, `Children`, `InnerFocused`.

#### ScrollableText

```go
func NewScrollableText(id string, styles ScrollableTextStyles) *ScrollableText
func DefaultScrollableTextStyles() ScrollableTextStyles
```

| Method | Signature |
|--------|-----------|
| SetContent | `func(text string)` |
| Content | `func() string` |
| SetWrap | `func(bool)` |
| ScrollTo | `func(line int)` |
| ScrollTop | `func()` |
| Update | `func(msg tea.KeyMsg) (tea.Cmd, bool)` |
| View | `func() string` |

#### TabComponent

A container that manages tab switching using TCB layout internally: TabBar header at the top, active tab content filling remaining height.

```go
func NewTabComponent(id string, styles TabBarStyles) *TabComponent
```

| Method | Signature |
|--------|-----------|
| AddTab | `func(label string, content Component)` |
| SetTabKeyBinding | `func(index int, keys string, description ...string)` |
| ActiveTab | `func() int` |
| SetActiveTab | `func(index int)` |
| TabBar | `func() *TabBar` |
| View | `func() string` |

Embeds `BaseContainer`. Inherits: `AddChild`, `Children`, `InnerFocused`.

---

## Package `input`

#### FocusManager

```go
func NewFocusManager(root widget.Container) *FocusManager
```

| Method | Signature | Description |
|--------|-----------|-------------|
| SetRoot | `func(root widget.Container)` | Re-root to a different component tree |
| Leaves | `func() []widget.Leaf` | All focusable, active leaves in tree order |
| FocusedLeaf | `func() widget.Leaf` | Currently focused leaf, or nil |
| Next | `func() tea.Cmd` | Focus next leaf (Tab) |
| Prev | `func() tea.Cmd` | Focus previous leaf (Shift-Tab) |
| FocusByID | `func(id string) (bool, tea.Cmd)` | Focus leaf by ID |
| FocusFirst | `func() tea.Cmd` | Focus first available leaf |
| HitTest | `func(x, y int) widget.Leaf` | Find leaf at screen coordinates |
| IsFocusChangingEvent | `func(msg tea.MouseMsg) bool` | Check if mouse event should change focus |
| CanFocusTo | `func(leaf widget.Leaf) bool` | Check if leaf can receive focus |
| ChangeFocusTo | `func(leaf widget.Leaf) (bool, tea.Cmd)` | Blur current, focus new |
| ResolveKeyBinding | `func(msg tea.KeyMsg) (widget.Component, func() tea.Cmd, bool)` | Check global key bindings |
| FocusedKeyBindings | `func() []key.Binding` | Key bindings from focused leaf |
| AllActiveKeyBindings | `func() []key.Binding` | Key bindings from all visible, active components in tree order |

#### KeyBindingResolver

```go
func NewKeyBindingResolver(root widget.Container) *KeyBindingResolver
```

| Method | Signature | Description |
|--------|-----------|-------------|
| SetRoot | `func(root widget.Container)` | Change root container |
| Resolve | `func(msg tea.KeyMsg) (widget.Component, func() tea.Cmd, bool)` | Find matching binding in tree (children-first, root last) |

---

## Package `window`

#### BaseWindow (internal)

`BaseWindow` is embedded by `MainWindow` and `PopupWindow`. It is not instantiated directly by users.

Embeds `widget.BaseContainer`. Has an internal content `Panel` with configurable layout (default `TCB`).

| Method | Signature | Description |
|--------|-----------|-------------|
| Add | `func(child widget.Component, position widget.Position)` | Place a child at the given position in the content panel |
| OnKeyPress | `func(fn func(tea.KeyMsg) tea.Cmd)` | Fallthrough handler for unhandled keys |
| OnUpdate | `func(fn func() tea.Cmd)` | Called after every event (focus change, key press, etc.) |
| ActiveKeyBindings | `func() []key.Binding` | All registered key bindings from visible, active components |
| ShowPopup | `func(popup *PopupWindow) tea.Cmd` | Push a popup onto the stack |
| View | `func() string` | Render child components |

#### MainWindow

```go
func NewWindow(id string, layout ...widget.Layout) *MainWindow
```

The main application window. Layout defaults to `TCB` if omitted. Embeds `BaseWindow`. Inherits all `BaseWindow` methods: `Add`, `OnKeyPress`, `ShowPopup`, `View`.

#### PopupWindow

```go
func NewPopupWindow(id, title string, styles PopupStyles) *PopupWindow
func DefaultPopupStyles() PopupStyles
```

An overlay popup with a title border and Escape-to-close. Embeds `BaseWindow`. Inherits: `Add`, `OnKeyPress`, `ShowPopup`, `View`.

| Method | Signature | Description |
|--------|-----------|-------------|
| Title | `func() string` | The popup's border title |
| OnResult | `func(fn func(value any) tea.Cmd)` | Register result callback, called when popup closes |
| Close | `func(result any) tea.Cmd` | Close popup and deliver result to OnResult callback |

#### App

```go
func NewApp(win *MainWindow) *App
```

Bridges `MainWindow` to `tea.Model`. Handles `tea.WindowSizeMsg` internally.

| Method | Signature | Description |
|--------|-----------|-------------|
| Init | `func() tea.Cmd` | Implements `tea.Model` |
| Update | `func(msg tea.Msg) (tea.Model, tea.Cmd)` | Implements `tea.Model` |
| View | `func() string` | Implements `tea.Model` |

#### OverlayOffset

```go
type OverlayOffset struct {
    X, Y int
}
```

#### RenderOverlay

```go
func RenderOverlay(content, title string, width, height int) (string, OverlayOffset)
```

Renders popup content centered on a dimmed background. Returns the rendered screen and the content offset for mouse coordinate adjustment.
