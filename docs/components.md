# Components

All interactive components (leaves) handle keyboard input when focused and can respond to mouse clicks. Each component has default styles and can be customized.

## Button

A focusable action button that responds to Space, Enter, and mouse clicks.

```go
btn := widget.NewButton("save", "Save", widget.DefaultButtonStyles())
btn.OnPress(func() tea.Cmd {
    return saveDocument()
})
```

Renders as `[ Save ]` with appropriate styling for normal, focused, and inactive states.

### Constructor

```go
func NewButton(id, label string, styles ButtonStyles) *Button
```

### Methods

| Method | Description |
|--------|-------------|
| `OnPress(fn func() tea.Cmd)` | Set the callback for button press |
| `BindDefaultActionToKey(binding key.Binding)` | Bind a global shortcut that triggers OnPress |

### Styles

```go
type ButtonStyles struct {
    Normal  lipgloss.Style  // unfocused appearance
    Focused lipgloss.Style  // focused appearance
}
```

Built-in styles: `DefaultButtonStyles()`, `DefaultPopupButtonStyles()` (dimmer, for trigger buttons).

### Keyboard

| Key | Action |
|-----|--------|
| Space | Trigger OnPress |
| Enter | Trigger OnPress |

### Global Shortcuts

Bind a global key that triggers the button's action from anywhere:

```go
btn.BindDefaultActionToKey(key.NewBinding(
    key.WithKeys("ctrl+s"),
    key.WithHelp("ctrl+s", "Save"),
))
```

When Ctrl+S is pressed anywhere in the window, the button receives focus and its OnPress fires.

---

## TextInput

A focusable text input field wrapping Bubble Tea's `textinput` component. Handles character input, cursor movement, and clipboard operations.

```go
input := widget.NewTextInput("email", 40)
input.WithPlaceholder("user@example.com")
input.WithCharLimit(100)
input.OnSubmit(func(value string) tea.Cmd {
    return validateEmail(value)
})
input.OnChange(func(value string) tea.Cmd {
    return nil // react to every keystroke
})
```

### Constructor

```go
func NewTextInput(id string, inputWidth int) *TextInput
```

`inputWidth` sets the visual width of the input area in columns.

### Methods

| Method | Description |
|--------|-------------|
| `WithPlaceholder(p string) *TextInput` | Set placeholder text (chainable) |
| `WithCharLimit(n int) *TextInput` | Set max character count (chainable) |
| `OnSubmit(fn func(string) tea.Cmd)` | Callback when Enter is pressed |
| `OnChange(fn func(string) tea.Cmd)` | Callback on every value change |
| `SetValue(v string)` | Set the input value programmatically |
| `Value() string` | Get the current value |
| `SetError(e string)` | Set a validation error message |
| `Error() string` | Get the current error message |

### Keyboard

When focused, TextInput consumes all key input. It supports standard text editing: arrow keys, Home/End, Ctrl+A (select all), Ctrl+K (kill line), backspace, delete, and paste.

| Key | Action |
|-----|--------|
| Enter | Trigger OnSubmit |
| All other keys | Handled by the inner text input |

### Focus and Cursor

When a TextInput gains focus, `SetFocused(true)` returns a `tea.Cmd` that starts the cursor blink timer. Always batch this command into your Bubble Tea update loop.

---

## FormattedTextInput

Extends TextInput with validation and formatting that run automatically on blur (when the input loses focus).

```go
dateInput := widget.NewFormattedTextInput("date", 12)
dateInput.WithPlaceholder("YYYY-MM-DD")

dateInput.WithValidation(func(s string) error {
    _, err := time.Parse("2006-01-02", s)
    return err
})

dateInput.WithFormat(func(s string) string {
    t, err := time.Parse("2006-01-02", s)
    if err != nil { return s }
    return t.Format("2006-01-02") // normalize format
})
```

### Constructor

```go
func NewFormattedTextInput(id string, inputWidth int) *FormattedTextInput
```

### Methods

| Method | Description |
|--------|-------------|
| `WithValidation(fn func(string) error)` | Set validator (runs on blur) |
| `WithFormat(fn func(string) string)` | Set formatter (runs on blur after validation passes) |

Inherits all TextInput methods (`WithPlaceholder`, `Value`, `SetValue`, etc.).

### Behavior

When the input loses focus:
1. Validation runs. If it returns an error, `SetError()` is called with the message.
2. If validation passes, the format function runs and replaces the value.

---

## Toggle

A boolean on/off switch with two rendering modes.

```go
// On/Off mode: renders "Feature:[on]" or "Feature:[off]"
toggle := widget.NewToggle("toggle1", "Feature", false, widget.ToggleModeOnOff, widget.DefaultToggleStyles())

// Radio mode: renders "[Live] [Cache]" with one highlighted
modeToggle := widget.NewToggle("mode", "Source", true, widget.ToggleModeRadio, widget.DefaultToggleStyles())
modeToggle.SetLabels("[Live]", "[Cache]")

toggle.OnChange(func(on bool) tea.Cmd {
    return applyFilter(on)
})
```

### Constructor

```go
func NewToggle(id, label string, initial bool, mode ToggleMode, styles ToggleStyles) *Toggle
```

- `label` — prefix text (rendered as "Label:")
- `initial` — starting state
- `mode` — `ToggleModeOnOff` or `ToggleModeRadio`

### Methods

| Method | Description |
|--------|-------------|
| `SetLabels(onLabel, offLabel string)` | Customize the on/off display text |
| `OnChange(fn func(bool) tea.Cmd)` | Callback when state changes |
| `On() bool` | Get current state |
| `SetOn(v bool)` | Set state programmatically |
| `BindDefaultActionToKey(binding key.Binding)` | Global shortcut that toggles state |

### Keyboard

| Key | Action |
|-----|--------|
| Space | Toggle on/off |
| Enter | Toggle on/off |

### Rendering Modes

**OnOff mode** — Shows one value at a time, padded for stable width:
```
Feature:[on]     (focused: yellow)
Feature:[off]    (unfocused: dim)
```

**Radio mode** — Shows both labels, highlights the active one:
```
Source:[Live] [Cache]    (Live is on, highlighted)
```

---

## CheckboxList

A scrollable list of checkable items with keyboard navigation.

```go
items := []widget.CheckboxItem{
    {Label: "node-1", Value: "n1"},
    {Label: "node-2", Value: "n2", Checked: true},
    {Label: "node-3", Value: "n3"},
    {Label: "--- Group ---", Value: "", IsGroup: true},
    {Label: "node-4", Value: "n4"},
}
list := widget.NewCheckboxList("nodes", items, widget.DefaultCheckboxListStyles())
list.OnChange(func(items []widget.CheckboxItem) tea.Cmd {
    return updateFilter(items)
})
```

### Constructor

```go
func NewCheckboxList(id string, items []CheckboxItem, styles CheckboxListStyles) *CheckboxList
```

### CheckboxItem

```go
type CheckboxItem struct {
    Label   string  // display text
    Value   string  // programmatic value
    Checked bool    // checked state
    IsGroup bool    // group headers (rendered differently)
}
```

### Methods

| Method | Description |
|--------|-------------|
| `OnChange(fn func([]CheckboxItem) tea.Cmd)` | Callback when any item is toggled |
| `Items() []CheckboxItem` | Get all items with current state |
| `Cursor() int` | Get cursor position |
| `Selected() []string` | Get values of all checked items |
| `BindDefaultActionToKey(binding key.Binding)` | Global shortcut to toggle cursor item |

### Keyboard

| Key | Action |
|-----|--------|
| Up / k | Move cursor up |
| Down / j | Move cursor down |
| Space | Toggle item at cursor |

### Mouse

Clicking on an item moves the cursor to it and toggles its checked state. The click Y coordinate is used to determine which item was clicked.

---

## TabBar

A horizontal tab selector with distinct cursor and active states.

```go
tabs := widget.NewTabBar("tabs", []string{"Overview", "Settings", "About"}, widget.DefaultTabBarStyles())
tabs.OnChange(func(index int) tea.Cmd {
    return switchTab(index)
})
```

### Constructor

```go
func NewTabBar(id string, labels []string, styles TabBarStyles) *TabBar
```

### Methods

| Method | Description |
|--------|-------------|
| `OnChange(fn func(int) tea.Cmd)` | Callback when active tab changes |
| `ActiveTab() int` | Get the currently active tab index |
| `SetActiveTab(i int)` | Set active tab and move cursor to match |
| `CursorTab() int` | Get the cursor position (may differ from active) |
| `SetTabKeyBinding(index int, keys string, desc ...string)` | Bind a key to activate a specific tab |
| `BindDefaultActionToKey(binding key.Binding)` | Global shortcut to activate cursor tab |

### Keyboard

| Key | Action |
|-----|--------|
| Left / h | Move cursor left |
| Right / l | Move cursor right |
| Space | Activate tab under cursor |
| Enter | Activate tab under cursor |

### Mouse

Clicking on a tab activates it directly and fires `OnChange`. If the clicked tab is already active, nothing happens.

### Per-Tab Accelerator Keys

Bind keyboard shortcuts to activate specific tabs directly, without moving the cursor first:

```go
tabs.SetTabKeyBinding(0, "ctrl+d")             // help text defaults to tab label ("Overview")
tabs.SetTabKeyBinding(1, "ctrl+e")             // "Settings"
tabs.SetTabKeyBinding(2, "ctrl+g", "About")    // explicit description
```

When the shortcut fires, the tab activates and `OnChange` fires. If the tab is already active, nothing happens. Calling `SetTabKeyBinding` again on the same index replaces the previous binding.

### Two-State Design

The TabBar separates **cursor** (which tab the keyboard highlight is on) from **active** (which tab's content is displayed). When the TabBar gains focus, the cursor resets to the active tab. Moving left/right moves the cursor. Pressing Space/Enter activates the tab under the cursor and fires OnChange.

Rendering:
- Active tab: bold, highlighted background
- Cursor tab (focused, not active): yellow highlight
- Active + cursor (focused, cursor on active tab): active style with underline
- Other tabs: dim/inactive

---

## Table

A generic, column-aware data table with scrolling, cursor selection, and per-cell rendering. Data is provided via a `TableDataSource` interface, enabling lazy data provision and cross-column rendering.

```go
columns := []widget.ColumnDef{
    {Title: "TIME", Width: 12},
    {Title: "LEVEL", Width: 5, Renderer: levelRenderer},
    {Title: "MESSAGE", Width: 0}, // 0 = take remaining space
}

ds := widget.NewSliceDataSource([][]string{
    {"12:00:01", "INFO", "Task completed successfully"},
    {"12:00:02", "WARN", "Slow response detected"},
    {"12:00:03", "ERROR", "Connection refused"},
})

table := widget.NewTable("data-table", columns, ds, widget.DefaultTableStyles())
```

### Constructor

```go
func NewTable(id string, columns []ColumnDef, ds TableDataSource, styles TableStyles) *Table
```

### TableDataSource

```go
type TableDataSource interface {
    RowCount() int
    CellData(row, col int) string
}
```

The Table queries the data source each render cycle. To change the data, call `SetDataSource` with a new source. The Table guarantees it will only call `CellData` with valid indices (`row < RowCount()`, `col < len(columns)`).

### SliceDataSource

A built-in adapter for `[][]string`:

```go
func NewSliceDataSource(data [][]string) *SliceDataSource
```

`SetData(data [][]string)` replaces the backing data. The Table re-reads on the next render.

### ColumnDef

```go
type ColumnDef struct {
    Title    string        // header text
    Width    int           // fixed width; 0 = take remaining space
    Renderer CellRenderer  // custom cell renderer; nil = DefaultCellRenderer
}
```

### CellRenderer

```go
type CellRenderer func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string
```

The renderer receives the data source (for cross-column lookups), position, selection state, the resolved cell width, and the table's styles. It returns a fully styled string. The renderer is responsible for all styling including selection background. Example:

```go
func levelRenderer(ds widget.TableDataSource, row, col int, selected bool, width int, styles widget.TableStyles) string {
    value := ds.CellData(row, col)
    style := styles.Cell
    if selected {
        style = styles.Selected.Width(width)
    }
    switch value {
    case "ERROR":
        style = style.Foreground(lipgloss.Color("#ef5350")).Bold(true)
    case "WARN":
        style = style.Foreground(lipgloss.Color("#ffb74d"))
    }
    return style.Render(value)
}
```

When `Renderer` is nil, the built-in `DefaultCellRenderer` is used, which applies `styles.Selected` with full-width background for selected rows and `styles.Cell` otherwise.

### Methods

| Method | Description |
|--------|-------------|
| `SetDataSource(ds TableDataSource)` | Replace the data source (cursor is clamped to range) |
| `DataSource() TableDataSource` | Get the current data source |
| `Cursor() int` | Get cursor row index |
| `SetCursor(c int)` | Set cursor position |
| `OnRowKeyPress(fn func(row int, msg tea.KeyMsg) tea.Cmd)` | Handler for unconsumed keys, receives cursor row |
| `OnRowClick(fn func(row int) tea.Cmd)` | Handler for row clicks |

### Keyboard

| Key | Action |
|-----|--------|
| Up / k | Move cursor up |
| Down / j | Move cursor down |
| Page Up | Move cursor up by viewport height |
| Page Down | Move cursor down by viewport height |
| Home | Jump to first row |
| End | Jump to last row |

### Mouse

Clicking on a data row moves the cursor to that row. Clicks on the header row are ignored. The click position accounts for the scroll offset, so clicking the first visible row selects the correct underlying data row.

Mouse wheel scrolling moves the cursor up/down by 3 rows per scroll tick, keeping the cursor within bounds.

### Scrolling

The table automatically scrolls to keep the cursor visible. The viewport height is `component height - 1` (subtracting the header row). When the cursor moves above the visible area, the view scrolls up. When it moves below, the view scrolls down.

### Flex Columns

Set `Width: 0` on a column to make it fill the remaining horizontal space after fixed-width columns are allocated. Only one flex column is supported.

---

## Card

A non-focusable bordered display box for summaries. Shows a title and a value.

```go
card := widget.NewCard("cpu", "CPU Usage", "42%", widget.DefaultCardStyles())
card.SetPreferredWidth(20)
card.SetPreferredHeight(4)

// Highlight when value exceeds threshold
card.SetAlert(true) // switches to alert style
card.SetValue("98%")
```

### Constructor

```go
func NewCard(id, title, value string, styles CardStyles) *Card
```

### Methods

| Method | Description |
|--------|-------------|
| `SetValue(value string)` | Update displayed value |
| `Value() string` | Get current value |
| `SetAlert(alert bool)` | Toggle alert styling |

Card is not focusable. It renders a bordered box with the title on the first line and the value on the second.

---

## Text

A non-focusable styled text display component. Used internally by Field for labels and separators, but also useful on its own.

```go
txt := widget.NewText("status", "Ready", lipgloss.NewStyle().Foreground(lipgloss.Color("#81c784")))
txt.SetText("Processing...")
txt.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb74d")))
```

### Constructor

```go
func NewText(id, text string, style lipgloss.Style) *Text
```

### Methods

| Method | Description |
|--------|-------------|
| `SetText(text string)` | Update displayed text |
| `GetText() string` | Get current text |
| `SetStyle(s lipgloss.Style)` | Change rendering style |
| `Style() lipgloss.Style` | Get current style |

Text is not focusable. When inactive (a parent is disabled), it renders with a faint style. The Text component respects `SetPreferredWidth` for width and `SetAlignment` for horizontal alignment.

---

## TabComponent

A container that manages tab switching. Internally uses TCB layout: the TabBar header sits at the top, and the active tab's content panel fills all remaining height (center slot). This means a TabComponent placed in a TCB center slot will correctly expand to fill available space, and the active content panel within it will also expand.

```go
tabs := widget.NewTabComponent("main-tabs", widget.DefaultTabBarStyles())

dashPanel := widget.NewPanel("dash", widget.TCB)
// ... populate dashPanel ...

settingsPanel := widget.NewPanel("settings", widget.Vertical)
// ... populate settingsPanel ...

tabs.AddTab("Dashboard", dashPanel)
tabs.AddTab("Settings", settingsPanel)

// Optional: bind keyboard shortcuts to switch tabs directly
tabs.SetTabKeyBinding(0, "ctrl+d", "Dashboard")
tabs.SetTabKeyBinding(1, "ctrl+e", "Settings")

win.Add(tabs, widget.TCBCenter)
```

### Constructor

```go
func NewTabComponent(id string, styles TabBarStyles) *TabComponent
```

### Methods

| Method | Description |
|--------|-------------|
| `AddTab(label string, content Component)` | Add a tab with the given label and content component |
| `SetTabKeyBinding(index int, keys string, description ...string)` | Bind a keyboard shortcut to activate a specific tab |
| `ActiveTab() int` | Get the index of the currently active tab |
| `SetActiveTab(index int)` | Switch to the tab at the given index |
| `TabBar() *TabBar` | Get the underlying TabBar leaf component |

### Nesting

TabComponent supports nesting — a TabComponent can be the content of another TabComponent's tab:

```go
innerTabs := widget.NewTabComponent("inner-tabs", widget.DefaultTabBarStyles())
innerTabs.AddTab("Sub A", subPanelA)
innerTabs.AddTab("Sub B", subPanelB)

outerTabs := widget.NewTabComponent("outer-tabs", widget.DefaultTabBarStyles())
outerTabs.AddTab("Overview", overviewPanel)
outerTabs.AddTab("Details", innerTabs)  // nested TabComponent as tab content

win.Add(outerTabs, widget.TCBCenter)
```

### How It Stretches

TabComponent uses TCB layout internally. When you place a TabComponent in a TCB center slot (e.g., in a window), the layout engine gives it all remaining height. TabComponent then gives all remaining height (after the TabBar header row) to the active content panel. This chain means content always fills available space without any manual size hints.

---

## ScrollableText

A focusable, scrollable, read-only text display area. Renders text within its allocated bounds and scrolls with keyboard input when focused. Content may contain ANSI color codes.

```go
st := widget.NewScrollableText("detail", widget.DefaultScrollableTextStyles())
st.SetPreferredWidth(80)
st.SetPreferredHeight(20)
st.SetContent(longText)
```

### Constructor

```go
func NewScrollableText(id string, styles ScrollableTextStyles) *ScrollableText
func DefaultScrollableTextStyles() ScrollableTextStyles
```

### ScrollableTextStyles

```go
type ScrollableTextStyles struct {
    Normal  lipgloss.Style  // unfocused appearance
    Focused lipgloss.Style  // focused appearance
}
```

### Methods

| Method | Description |
|--------|-------------|
| `SetContent(text string)` | Set the full text content (resets scroll position) |
| `Content() string` | Get the current content |
| `SetWrap(bool)` | Set line wrapping (default: true). When false, lines are truncated |
| `ScrollTo(line int)` | Scroll to a specific line (clamped) |
| `ScrollTop()` | Scroll to the top |

### Keyboard

| Key | Action |
|-----|--------|
| Up / k | Scroll up one line |
| Down / j | Scroll down one line |
| Page Up | Scroll up by viewport height |
| Page Down | Scroll down by viewport height |
| Home | Scroll to top |
| End | Scroll to bottom |

### Mouse

Mouse wheel scrolling moves the viewport up/down by 3 lines per scroll tick. The component does not need keyboard focus to receive scroll events — scrolling targets the component under the cursor.
