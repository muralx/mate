# Styling

All Mate components use [Lip Gloss](https://github.com/charmbracelet/lipgloss) for terminal styling. Styles are passed to components via constructor config structs and can be customized to create any visual theme.

## Style Structs

Each component type has its own styles struct. All components provide a `Default*Styles()` function that returns sensible defaults.

### ButtonStyles

```go
type ButtonStyles struct {
    Normal  lipgloss.Style  // unfocused, active
    Focused lipgloss.Style  // focused
}
```

Inactive buttons render with `Faint(true)` applied automatically.

```go
// Default
styles := widget.DefaultButtonStyles()

// Custom
styles := widget.ButtonStyles{
    Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4fc3f7")),
    Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#fff")).Background(lipgloss.Color("#1565c0")).Bold(true),
}

btn := widget.NewButton("save", "Save", styles)
```

There is also `DefaultPopupButtonStyles()` which provides a dimmer style suitable for trigger buttons like `[▾]`.

### ToggleStyles

```go
type ToggleStyles struct {
    Label       lipgloss.Style  // label prefix
    OnActive    lipgloss.Style  // on value, unfocused
    OnFocused   lipgloss.Style  // on value, focused
    OffActive   lipgloss.Style  // off value, unfocused
    OffFocused  lipgloss.Style  // off value, focused
    OffInactive lipgloss.Style  // unselected value in Radio mode
}
```

### CheckboxListStyles

```go
type CheckboxListStyles struct {
    Cursor    lipgloss.Style  // "> " marker on current item
    Checked   lipgloss.Style  // "[x] " indicator
    Unchecked lipgloss.Style  // "[ ] " indicator
    Item      lipgloss.Style  // normal item label
    Group     lipgloss.Style  // group header label
    Dim       lipgloss.Style  // non-cursor item prefix ("  ")
}
```

### TabBarStyles

```go
type TabBarStyles struct {
    Active   lipgloss.Style  // selected tab
    Inactive lipgloss.Style  // non-selected tab
    Focused  lipgloss.Style  // keyboard cursor on an inactive tab
}
```

### TableStyles

```go
type TableStyles struct {
    Header   lipgloss.Style  // column header row
    Selected lipgloss.Style  // selected row highlight
    Cell     lipgloss.Style  // default cell style (when no column Renderer)
}
```

For per-cell custom styling, use `ColumnDef.Renderer` instead of relying on `TableStyles.Cell`. See [Components - Table](components.md#table).

### BorderConfig

Panel borders are configured with `BorderConfig` rather than a style struct. Use `SetBorder` to apply it:

```go
// Default border (rounded, blue inactive / cyan active)
panel.SetBorder(widget.DefaultBorder())

// Custom colors
panel.SetBorder(widget.SingleLineBorder("#444444", "#4fc3f7"))

// Full control
panel.SetBorder(widget.BorderConfig{
    Type:        widget.RoundedBorder,
    Color:       lipgloss.Color("#444444"),
    ActiveColor: lipgloss.Color("#4fc3f7"),
    Padding:     1,
})
```

`NoBorder` (the default when `SetBorder` is not called) renders no border at all — the panel's content fills its allocated space without chrome.

### CardStyles

```go
type CardStyles struct {
    Border lipgloss.Style  // outer border
    Title  lipgloss.Style  // title text
    Value  lipgloss.Style  // normal value
    Alert  lipgloss.Style  // alert value (when SetAlert(true))
}
```

### FieldStyles

```go
type FieldStyles struct {
    Label     lipgloss.Style  // label when no child focused
    LabelHot  lipgloss.Style  // label when a child has focus
    Separator lipgloss.Style  // the ": " between label and input
}
```

## Creating a Theme

Define a cohesive set of colors and apply them across all component styles:

```go
// Define your palette
var (
    primary   = lipgloss.Color("#4fc3f7")
    secondary = lipgloss.Color("#81c784")
    warning   = lipgloss.Color("#ffb74d")
    danger    = lipgloss.Color("#ef5350")
    text      = lipgloss.Color("#e0e0e0")
    dimText   = lipgloss.Color("#888888")
    bg        = lipgloss.Color("#2a2a3e")
    highlight = lipgloss.Color("#ffeb3b")
)

// Build styles from the palette
func myButtonStyles() widget.ButtonStyles {
    return widget.ButtonStyles{
        Normal:  lipgloss.NewStyle().Foreground(primary).Bold(true),
        Focused: lipgloss.NewStyle().Foreground(highlight).Background(bg).Bold(true),
    }
}

func myBorder() widget.BorderConfig {
    return widget.BorderConfig{
        Type:        widget.RoundedBorder,
        Color:       dimText,
        ActiveColor: primary,
        Padding:     1,
    }
}

func myFieldStyles() widget.FieldStyles {
    return widget.FieldStyles{
        Label:     lipgloss.NewStyle().Foreground(dimText),
        LabelHot:  lipgloss.NewStyle().Foreground(highlight),
        Separator: lipgloss.NewStyle().Foreground(dimText),
    }
}
```

## Inactive Rendering

When a component becomes inactive (a parent is disabled), it automatically renders with `Faint(true)`. This applies to:
- Button: faint label
- Toggle: faint entire output
- CheckboxList: faint entire list
- Panel: faint the entire bordered area
- Text: faint text
- Card: faint the entire card

You do not need to handle inactive styling yourself unless you write a custom component.

## Table Cell Renderers

The Table widget supports per-column custom renderers via `CellRenderer`. Renderers receive the full data source, enabling cross-column lookups (e.g., color a cell based on another column's value):

```go
columns := []widget.ColumnDef{
    {
        Title: "STATUS",
        Width: 8,
        Renderer: func(ds widget.TableDataSource, row, col int, selected bool, width int, styles widget.TableStyles) string {
            value := ds.CellData(row, col)
            style := styles.Cell
            if selected {
                style = styles.Selected.Width(width)
            }
            switch value {
            case "OK":
                style = style.Foreground(lipgloss.Color("#81c784"))
            case "FAIL":
                style = style.Foreground(lipgloss.Color("#ef5350")).Bold(true)
            }
            return style.Render(value)
        },
    },
    {Title: "NAME", Width: 20},  // nil renderer = DefaultCellRenderer
    {Title: "DETAILS", Width: 0}, // flex column
}
```

The renderer is responsible for all styling including the selection background. Use the `width` parameter with `styles.Selected.Width(width)` to ensure the selection highlight covers the full cell width. When `Renderer` is nil, `DefaultCellRenderer` handles this automatically.

## Color Profiles

Lip Gloss automatically adapts colors to the terminal's color profile (TrueColor, 256-color, 16-color, or no color). You can force a profile for testing:

```go
import "github.com/muesli/termenv"

lipgloss.SetColorProfile(termenv.Ascii)     // no color
lipgloss.SetColorProfile(termenv.TrueColor) // full color
```

## Width and Sizing

Components sized by the layout engine will pad their rendered output to fill the allocated width using `RenderInSize()`. This uses `lipgloss.NewStyle().Width(w).MaxWidth(w)` to ensure consistent widths regardless of content length.

Use `SetPreferredWidth(w)` to give a component a fixed width hint. The `SetAlignment` method controls where content sits within the padded width: left (default), center, or right.
