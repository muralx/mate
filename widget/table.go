package widget

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// TableDataSource provides row data to a Table. Implementations must return
// consistent results within a single render cycle. To change the data, call
// Table.SetDataSource with a new (or updated) source.
type TableDataSource interface {
	// RowCount returns the total number of rows.
	RowCount() int
	// CellData returns the raw (unstyled) string for the given cell.
	// The Table guarantees row < RowCount() and col < len(columns).
	CellData(row, col int) string
}

// SliceDataSource is a TableDataSource backed by a [][]string.
type SliceDataSource struct {
	data [][]string
}

// NewSliceDataSource creates a SliceDataSource wrapping the given data.
func NewSliceDataSource(data [][]string) *SliceDataSource {
	return &SliceDataSource{data: data}
}

func (s *SliceDataSource) RowCount() int { return len(s.data) }

func (s *SliceDataSource) CellData(row, col int) string {
	if row < 0 || row >= len(s.data) {
		return ""
	}
	r := s.data[row]
	if col < 0 || col >= len(r) {
		return ""
	}
	return r[col]
}

// SetData replaces the backing data. The Table re-reads on the next render.
func (s *SliceDataSource) SetData(data [][]string) {
	s.data = data
}

// CellRenderer renders a single table cell. The renderer receives the data
// source (for cross-column lookups), position, selection state, the resolved
// cell width, and the table's styles. It returns a fully styled string.
// Use PrepareCell to get a truncated value safe for single-line rendering.
type CellRenderer func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string

// PrepareCell returns the cell value truncated to width, ready for styling.
// Use this in custom renderers to prevent lipgloss.Width() from wrapping
// long text into multiple lines.
func PrepareCell(ds TableDataSource, row, col, width int) string {
	value := strings.ReplaceAll(ds.CellData(row, col), "\n", `\n`)
	return ansi.Truncate(value, width, "")
}

// DefaultCellRenderer is the built-in cell renderer used when ColumnDef.Renderer
// is nil. It applies styles.Selected (with full width background) for selected
// rows, and styles.Cell otherwise.
func DefaultCellRenderer(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
	value := PrepareCell(ds, row, col, width)
	if selected {
		return styles.Selected.Width(width).Render(value)
	}
	return styles.Cell.Render(value)
}

// ColumnDef defines a single column in the table.
type ColumnDef struct {
	Title    string
	Width    int          // fixed width; 0 = take remaining space
	Renderer CellRenderer // nil = default renderer
}

// TableStyles defines the styles used by the Table.
type TableStyles struct {
	Header   lipgloss.Style
	Selected lipgloss.Style
	Cell     lipgloss.Style // default cell style when no column renderer
}

// DefaultTableStyles returns a TableStyles with sensible defaults.
func DefaultTableStyles() TableStyles {
	return TableStyles{
		Header:   lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true),
		Selected: lipgloss.NewStyle().Background(lipgloss.Color("#2a2a3e")),
		Cell:     lipgloss.NewStyle(),
	}
}

// Table is a generic, column-aware scrollable data table with cursor selection
// and per-cell rendering. Data is provided via a TableDataSource.
type Table struct {
	FocusableComponent
	columns       []ColumnDef
	ds            TableDataSource
	cursor        int
	offset        int
	styles        TableStyles
	columnSpacing int
	onRowKeyPress func(row int, msg tea.KeyMsg) tea.Cmd
	onRowClick    func(row int) tea.Cmd
}

// NewTable creates a new Table with the given ID, columns, data source, and styles.
func NewTable(id string, columns []ColumnDef, ds TableDataSource, styles TableStyles) *Table {
	t := &Table{
		columns:       columns,
		columnSpacing: 1,
		ds:            ds,
		styles:        styles,
	}
	t.FocusableComponent = NewFocusableComponent(id)
	return t
}

// SetDataSource replaces the data source and clamps the cursor.
func (t *Table) SetDataSource(ds TableDataSource) {
	t.ds = ds
	t.clampCursor()
}

// DataSource returns the current data source.
func (t *Table) DataSource() TableDataSource { return t.ds }

// SetColumnSpacing sets the number of spaces between columns. Default is 1.
func (t *Table) SetColumnSpacing(spacing int) { t.columnSpacing = spacing }

// clampCursor ensures the cursor is within the data source bounds.
func (t *Table) clampCursor() {
	n := t.ds.RowCount()
	if t.cursor >= n {
		t.cursor = n - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
}

// OnRowKeyPress sets a handler called when the table has focus and receives a key
// that the table's internal navigation doesn't consume. Unlike OnKeyPress, this
// passes the current cursor row so the handler has row context.
func (t *Table) OnRowKeyPress(fn func(row int, msg tea.KeyMsg) tea.Cmd) {
	t.onRowKeyPress = fn
}

// OnRowClick sets a handler called when a data row is clicked.
// The handler receives the zero-based row index.
func (t *Table) OnRowClick(fn func(row int) tea.Cmd) {
	t.onRowClick = fn
}

// Cursor returns the current cursor index.
func (t *Table) Cursor() int { return t.cursor }

// SetCursor sets the cursor to c if in range.
func (t *Table) SetCursor(c int) {
	if c >= 0 && c < t.ds.RowCount() {
		t.cursor = c
	}
}

// viewportRows returns how many data rows fit in the viewport (total height minus header).
func (t *Table) viewportRows() int {
	vp := t.height - 1 // header row
	if vp < 1 {
		vp = 1
	}
	return vp
}

// rowCount returns the current row count from the data source.
func (t *Table) rowCount() int {
	return t.ds.RowCount()
}

// moveCursor shifts the cursor by delta rows (negative=up, positive=down),
// clamps to valid range, and scrolls the viewport to keep it visible.
func (t *Table) moveCursor(delta int) {
	n := t.rowCount()
	t.cursor += delta
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= n {
		t.cursor = n - 1
	}
	t.ensureVisible(t.viewportRows())
}

// Update handles key input for table navigation.
func (t *Table) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !t.Active() {
		return nil, false
	}
	vp := t.viewportRows()
	switch msg.String() {
	case "up", "k":
		t.moveCursor(-1)
		return nil, true
	case "down", "j":
		t.moveCursor(1)
		return nil, true
	case "pgup":
		t.moveCursor(-vp)
		return nil, true
	case "pgdown":
		t.moveCursor(vp)
		return nil, true
	case "home":
		t.cursor = 0
		t.offset = 0
		return nil, true
	case "end":
		t.moveCursor(t.rowCount())
		return nil, true
	}
	if t.onRowKeyPress != nil {
		if cmd := t.onRowKeyPress(t.cursor, msg); cmd != nil {
			return cmd, true
		}
	}
	if t.onKeyPress != nil {
		if cmd := t.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// ensureVisible adjusts the scroll offset so the cursor is within the viewport.
func (t *Table) ensureVisible(vp int) {
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+vp {
		t.offset = t.cursor - vp + 1
	}
}

// View renders the table: header row followed by visible data rows.
func (t *Table) View() string {
	n := t.rowCount()
	if n == 0 {
		return lipgloss.NewStyle().
			Width(t.width).
			Height(t.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No data")
	}

	var lines []string
	lines = append(lines, t.renderHeader())

	vp := t.viewportRows()
	end := t.offset + vp
	if end > n {
		end = n
	}

	for i := t.offset; i < end; i++ {
		selected := i == t.cursor && t.Focused()
		lines = append(lines, t.renderRow(i, selected))
	}

	// Pad with empty lines to fill allocated height
	for len(lines) < t.height {
		if t.width > 0 {
			lines = append(lines, strings.Repeat(" ", t.width))
		} else {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the column titles.
func (t *Table) renderHeader() string {
	var parts []string
	remaining := t.width
	sep := strings.Repeat(" ", t.columnSpacing)
	for i, col := range t.columns {
		if i > 0 {
			remaining -= t.columnSpacing
		}
		w := col.Width
		if w == 0 {
			w = remaining
		}
		if w > remaining {
			w = remaining
		}
		remaining -= w
		title := ansi.Truncate(col.Title, w, "")
		parts = append(parts, t.styles.Header.Render(ansiPadRight(title, w)))
	}
	return strings.Join(parts, sep)
}

// renderRow renders a single data row. The renderer is responsible for all
// styling including selection background. When no custom renderer is set,
// the default renderer applies styles.Selected or styles.Cell as appropriate.
func (t *Table) renderRow(row int, selected bool) string {
	var parts []string
	remaining := t.width
	sep := strings.Repeat(" ", t.columnSpacing)
	for col, colDef := range t.columns {
		if col > 0 {
			remaining -= t.columnSpacing
		}
		w := colDef.Width
		if w == 0 {
			w = remaining
		}
		if w > remaining {
			w = remaining
		}
		remaining -= w

		renderer := colDef.Renderer
		if renderer == nil {
			renderer = DefaultCellRenderer
		}
		cell := renderer(t.ds, row, col, selected, w, t.styles)
		// Truncate and pad to cell width
		cell = ansi.Truncate(cell, w, "")
		if lipgloss.Width(cell) < w {
			cell = ansiPadRight(cell, w)
		}
		parts = append(parts, cell)
	}
	return strings.Join(parts, sep)
}

// ansiPadRight pads a (possibly ANSI-styled) string with spaces to width w.
func ansiPadRight(s string, w int) string {
	current := lipgloss.Width(s)
	if current >= w {
		return s
	}
	return s + strings.Repeat(" ", w-current)
}

// HandleEvent handles mouse clicks and scroll wheel events.
// Clicks select the row under the cursor. Scroll wheel moves the cursor
// up/down by 3 rows (matching typical terminal scroll speed).
func (t *Table) HandleEvent(event Event) (tea.Cmd, bool) {
	if scroll, ok := event.(MouseScrollEvent); ok {
		if !t.Active() {
			return nil, false
		}
		t.moveCursor(scroll.Direction * 3)
		return nil, true
	}
	if click, ok := event.(MouseClickEvent); ok {
		if !t.Active() {
			return nil, false
		}
		_, ty := t.Position()
		row := click.Y - ty - 1 + t.offset // -1 for header
		if row < 0 || row >= t.rowCount() {
			return nil, false
		}
		t.cursor = row
		t.ensureVisible(t.viewportRows())
		if t.onRowClick != nil {
			cmd := t.onRowClick(row)
			return cmd, true
		}
		return nil, true
	}
	return t.BaseComponent.HandleEvent(event)
}
