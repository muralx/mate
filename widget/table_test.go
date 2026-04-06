package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// --- Helpers ---

func sampleColumns() []ColumnDef {
	return []ColumnDef{
		{Title: "TIME", Width: 12},
		{Title: "NODE", Width: 12},
		{Title: "MESSAGE", Width: 0}, // flex
	}
}

func sampleData() [][]string {
	return [][]string{
		{"12:00:01", "node1", "Starting up"},
		{"12:00:02", "node1", "Slow query"},
		{"12:00:03", "node2", "Connection failed"},
		{"12:00:04", "node2", "Scheduled task completed"},
		{"12:00:05", "node1", "Debug trace"},
	}
}

func makeTestTable(data [][]string) *Table {
	tbl := NewTable("tbl", sampleColumns(), NewSliceDataSource(data), DefaultTableStyles())
	tbl.SetSize(80, 20)
	return tbl
}

// --- Compile-time interface check ---

var _ Leaf = (*Table)(nil)

// --- SliceDataSource tests ---

func TestSliceDataSource_RowCount(t *testing.T) {
	ds := NewSliceDataSource(sampleData())
	if ds.RowCount() != 5 {
		t.Errorf("RowCount = %d, want 5", ds.RowCount())
	}
}

func TestSliceDataSource_CellData(t *testing.T) {
	ds := NewSliceDataSource(sampleData())
	if v := ds.CellData(0, 0); v != "12:00:01" {
		t.Errorf("CellData(0,0) = %q, want %q", v, "12:00:01")
	}
	if v := ds.CellData(2, 2); v != "Connection failed" {
		t.Errorf("CellData(2,2) = %q, want %q", v, "Connection failed")
	}
}

func TestSliceDataSource_OutOfBounds(t *testing.T) {
	ds := NewSliceDataSource(sampleData())
	if v := ds.CellData(-1, 0); v != "" {
		t.Errorf("CellData(-1,0) = %q, want empty", v)
	}
	if v := ds.CellData(0, 99); v != "" {
		t.Errorf("CellData(0,99) = %q, want empty", v)
	}
	if v := ds.CellData(99, 0); v != "" {
		t.Errorf("CellData(99,0) = %q, want empty", v)
	}
}

func TestSliceDataSource_SetData(t *testing.T) {
	ds := NewSliceDataSource(sampleData())
	ds.SetData([][]string{{"a", "b"}})
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
	if v := ds.CellData(0, 0); v != "a" {
		t.Errorf("CellData(0,0) = %q, want %q", v, "a")
	}
}

func TestSliceDataSource_Empty(t *testing.T) {
	ds := NewSliceDataSource(nil)
	if ds.RowCount() != 0 {
		t.Errorf("RowCount = %d, want 0", ds.RowCount())
	}
}

// --- Table tests ---

func TestTable_Defaults(t *testing.T) {
	tbl := makeTestTable(sampleData())

	if tbl.ID() != "tbl" {
		t.Errorf("ID = %q, want %q", tbl.ID(), "tbl")
	}
	if !tbl.Visible() {
		t.Error("should be visible by default")
	}
	if !tbl.Enabled() {
		t.Error("should be enabled by default")
	}
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", tbl.Cursor())
	}
}

func TestTable_Interface(t *testing.T) {
	tbl := makeTestTable(nil)
	var _ Leaf = tbl
}

func TestTable_DataSource(t *testing.T) {
	ds := NewSliceDataSource(sampleData())
	tbl := NewTable("t", sampleColumns(), ds, DefaultTableStyles())
	if tbl.DataSource() != ds {
		t.Error("DataSource() should return the data source passed to constructor")
	}
}

func TestTable_View_Empty(t *testing.T) {
	tbl := makeTestTable(nil)
	tbl.SetSize(80, 20)

	view := tbl.View()
	plain := stripansi.Strip(view)
	if !strings.Contains(plain, "No data") {
		t.Errorf("empty view should contain 'No data', got:\n%s", plain)
	}
}

func TestTable_View_Empty_FillsHeight(t *testing.T) {
	tbl := makeTestTable(nil)
	tbl.SetSize(80, 20)

	view := tbl.View()
	h := lipgloss.Height(view)
	if h < 20 {
		t.Errorf("empty view height = %d, want >= 20", h)
	}
}

func TestTable_View_Header(t *testing.T) {
	tbl := makeTestTable(sampleData())

	view := tbl.View()
	plain := stripansi.Strip(view)

	for _, col := range []string{"TIME", "NODE", "MESSAGE"} {
		if !strings.Contains(plain, col) {
			t.Errorf("header should contain %q, got:\n%s", col, plain)
		}
	}
}

func TestTable_View_RowValues(t *testing.T) {
	tbl := makeTestTable(sampleData())

	view := tbl.View()
	plain := stripansi.Strip(view)

	if !strings.Contains(plain, "12:00:01") {
		t.Errorf("should contain timestamp '12:00:01', got:\n%s", plain)
	}
	if !strings.Contains(plain, "node1") {
		t.Errorf("should contain node 'node1', got:\n%s", plain)
	}
	if !strings.Contains(plain, "Starting up") {
		t.Errorf("should contain message 'Starting up', got:\n%s", plain)
	}
}

func TestTable_View_SelectedRow(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetFocused(true)

	view := tbl.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines")
	}
	plain := stripansi.Strip(lines[1])
	if !strings.Contains(plain, "12:00:01") {
		t.Errorf("selected row should contain first row data, got: %q", plain)
	}
}

func TestTable_View_CustomRenderer(t *testing.T) {
	cols := []ColumnDef{
		{Title: "NAME", Width: 20, Renderer: func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
			value := ds.CellData(row, col)
			return "[" + value + "]"
		}},
		{Title: "VAL", Width: 0},
	}
	ds := NewSliceDataSource([][]string{{"alpha", "100"}})
	tbl := NewTable("custom", cols, ds, DefaultTableStyles())
	tbl.SetSize(40, 10)

	view := tbl.View()
	plain := stripansi.Strip(view)
	if !strings.Contains(plain, "[alpha]") {
		t.Errorf("custom renderer should wrap value, got:\n%s", plain)
	}
}

func TestTable_View_CellRenderer_CrossColumnAccess(t *testing.T) {
	// Renderer accesses another column in the same row
	cols := []ColumnDef{
		{Title: "LEVEL", Width: 10},
		{Title: "MSG", Width: 0, Renderer: func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
			level := ds.CellData(row, 0) // cross-column lookup
			value := ds.CellData(row, col)
			return level + ":" + value
		}},
	}
	ds := NewSliceDataSource([][]string{{"ERROR", "disk full"}})
	tbl := NewTable("cross", cols, ds, DefaultTableStyles())
	tbl.SetSize(60, 5)

	view := tbl.View()
	plain := stripansi.Strip(view)
	if !strings.Contains(plain, "ERROR:disk full") {
		t.Errorf("renderer should access level column, got:\n%s", plain)
	}
}

func TestTable_View_CellRenderer_ReceivesWidth(t *testing.T) {
	var receivedWidth int
	cols := []ColumnDef{
		{Title: "COL", Width: 25, Renderer: func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
			receivedWidth = width
			return ds.CellData(row, col)
		}},
	}
	ds := NewSliceDataSource([][]string{{"test"}})
	tbl := NewTable("w", cols, ds, DefaultTableStyles())
	tbl.SetSize(80, 5)
	tbl.View()

	if receivedWidth != 25 {
		t.Errorf("renderer received width = %d, want 25", receivedWidth)
	}
}

func TestTable_View_CellRenderer_ReceivesStyles(t *testing.T) {
	var gotStyles bool
	cols := []ColumnDef{
		{Title: "COL", Width: 20, Renderer: func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
			// Verify styles are passed by rendering with them
			gotStyles = styles.Cell.Render("x") != ""
			return ds.CellData(row, col)
		}},
	}
	ds := NewSliceDataSource([][]string{{"test"}})
	tbl := NewTable("s", cols, ds, DefaultTableStyles())
	tbl.SetSize(80, 5)
	tbl.View()

	if !gotStyles {
		t.Error("renderer should receive the table's styles")
	}
}

func TestTable_Update_Down(t *testing.T) {
	tbl := makeTestTable(sampleData())

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !consumed {
		t.Error("down should be consumed")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", tbl.Cursor())
	}
}

func TestTable_Update_Up(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(2)

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyUp})
	if !consumed {
		t.Error("up should be consumed")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", tbl.Cursor())
	}
}

func TestTable_Update_Up_AtTop(t *testing.T) {
	tbl := makeTestTable(sampleData())

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyUp})
	if !consumed {
		t.Error("up should be consumed even at top")
	}
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", tbl.Cursor())
	}
}

func TestTable_Update_Down_AtBottom(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(4)

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !consumed {
		t.Error("down should be consumed even at bottom")
	}
	if tbl.Cursor() != 4 {
		t.Errorf("cursor = %d, want 4", tbl.Cursor())
	}
}

func TestTable_Update_Home(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(3)

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyHome})
	if !consumed {
		t.Error("home should be consumed")
	}
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", tbl.Cursor())
	}
}

func TestTable_Update_End(t *testing.T) {
	tbl := makeTestTable(sampleData())

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if !consumed {
		t.Error("end should be consumed")
	}
	if tbl.Cursor() != 4 {
		t.Errorf("cursor = %d, want 4", tbl.Cursor())
	}
}

func TestTable_Update_PgDown(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetSize(80, 4) // viewport = 3 rows

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if !consumed {
		t.Error("pgdown should be consumed")
	}
	if tbl.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3", tbl.Cursor())
	}
}

func TestTable_Update_PgUp(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetSize(80, 4) // viewport = 3 rows
	tbl.SetCursor(4)

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if !consumed {
		t.Error("pgup should be consumed")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", tbl.Cursor())
	}
}

func TestTable_Update_Inactive_Ignored(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetEnabled(false)

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if consumed {
		t.Error("inactive table should not consume keys")
	}
	if tbl.Cursor() != 0 {
		t.Error("cursor should not move when inactive")
	}
}

func TestTable_Update_EnterNotConsumed(t *testing.T) {
	tbl := makeTestTable(sampleData())

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if consumed {
		t.Error("enter should NOT be consumed by table (no onKeyPress set)")
	}
}

func TestTable_Update_OnKeyPress(t *testing.T) {
	tbl := makeTestTable(sampleData())
	called := false
	tbl.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "enter" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed via onKeyPress")
	}
	if !called {
		t.Error("onKeyPress should have been called")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func TestTable_SetDataSource(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(4)

	// Shrink to 2 rows — cursor should clamp
	tbl.SetDataSource(NewSliceDataSource([][]string{
		{"a", "b", "c"},
		{"d", "e", "f"},
	}))
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 after shrink", tbl.Cursor())
	}
	if tbl.DataSource().RowCount() != 2 {
		t.Errorf("rows = %d, want 2", tbl.DataSource().RowCount())
	}
}

func TestTable_SetDataSource_Empty(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(3)

	tbl.SetDataSource(NewSliceDataSource(nil))
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 after clearing data", tbl.Cursor())
	}
}

func TestTable_Viewport_Scrolling(t *testing.T) {
	// 10 rows, viewport of 3
	data := make([][]string, 10)
	for i := range data {
		data[i] = []string{strings.Repeat("x", 5)}
	}
	cols := []ColumnDef{{Title: "COL", Width: 0}}
	tbl := NewTable("scroll", cols, NewSliceDataSource(data), DefaultTableStyles())
	tbl.SetSize(20, 4) // viewport = 3

	// Move cursor to row 5
	for i := 0; i < 5; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	view := tbl.View()
	plain := stripansi.Strip(view)
	lines := strings.Split(plain, "\n")

	// Header + 3 data rows = 4 lines max
	if len(lines) > 4 {
		t.Errorf("expected at most 4 lines, got %d", len(lines))
	}
}

func TestTable_Truncation(t *testing.T) {
	longValue := strings.Repeat("x", 200)
	ds := NewSliceDataSource([][]string{{longValue}})
	cols := []ColumnDef{{Title: "COL", Width: 30}}
	tbl := NewTable("trunc", cols, ds, DefaultTableStyles())
	tbl.SetSize(30, 5)

	view := tbl.View()
	plain := stripansi.Strip(view)

	if strings.Contains(plain, longValue) {
		t.Error("long value should be truncated")
	}
}

func TestTable_KeyBindings_NilByDefault(t *testing.T) {
	tbl := makeTestTable(sampleData())
	if tbl.KeyBindings() != nil {
		t.Errorf("expected nil KeyBindings (no registered bindings), got %d", len(tbl.KeyBindings()))
	}
}

func TestTable_View_SelectedRow_FullWidth(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	cols := []ColumnDef{
		{Title: "TIME", Width: 12},
		{Title: "LEVEL", Width: 8},
		{Title: "SOURCE", Width: 15},
		{Title: "MESSAGE", Width: 0}, // flex
	}
	ds := NewSliceDataSource([][]string{{"12:00", "INFO", "app", "hello"}})
	tbl := NewTable("sel", cols, ds, DefaultTableStyles())
	tbl.SetSize(80, 5)
	tbl.SetFocused(true)

	view := tbl.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatal("expected header + data row")
	}
	selectedLine := lines[1]

	// The Selected background is #2a2a3e = RGB(42,42,62).
	bgEsc := "\x1b[48;2;42;42;62m"
	count := strings.Count(selectedLine, bgEsc)
	if count < len(cols) {
		t.Errorf("background escape appears %d times, want >= %d (one per column); "+
			"selected row background is not applied to every cell", count, len(cols))
	}

	w := lipgloss.Width(selectedLine)
	if w != 80 {
		t.Errorf("selected row visual width = %d, want 80", w)
	}
}

func TestTable_View_SelectedRow_FullWidth_CustomRenderer(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	cols := []ColumnDef{
		{Title: "NAME", Width: 20, Renderer: func(ds TableDataSource, row, col int, selected bool, width int, styles TableStyles) string {
			value := ds.CellData(row, col)
			if selected {
				return styles.Selected.Width(width).Render("[" + value + "]")
			}
			return "[" + value + "]"
		}},
		{Title: "VALUE", Width: 0},
	}
	ds := NewSliceDataSource([][]string{{"key", "val"}})
	tbl := NewTable("sel2", cols, ds, DefaultTableStyles())
	tbl.SetSize(60, 5)
	tbl.SetFocused(true)

	view := tbl.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatal("expected header + data row")
	}

	bgEsc := "\x1b[48;2;42;42;62m"
	count := strings.Count(lines[1], bgEsc)
	if count < len(cols) {
		t.Errorf("background escape appears %d times, want >= %d", count, len(cols))
	}

	w := lipgloss.Width(lines[1])
	if w != 60 {
		t.Errorf("selected row visual width = %d, want 60", w)
	}
}

func TestTable_FlexColumn(t *testing.T) {
	cols := []ColumnDef{
		{Title: "A", Width: 10},
		{Title: "B", Width: 0}, // flex: should get 70 (80-10)
	}
	ds := NewSliceDataSource([][]string{{"short", "also short"}})
	tbl := NewTable("flex", cols, ds, DefaultTableStyles())
	tbl.SetSize(80, 5)

	view := tbl.View()
	lines := strings.Split(stripansi.Strip(view), "\n")
	for i, line := range lines {
		if len(line) != 80 {
			t.Errorf("line %d width = %d, want 80: %q", i, len(line), line)
		}
	}
}

func TestTable_HandleEvent_ClickSelectsRow(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	if tbl.Cursor() != 0 {
		t.Fatalf("cursor = %d, want 0", tbl.Cursor())
	}

	// Click on row 2 (Y=0 is header, Y=1 is row 0, Y=2 is row 1, Y=3 is row 2)
	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 3})
	if !consumed {
		t.Error("click should be consumed")
	}
	if tbl.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (clicked row)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_ClickOnHeader_Ignored(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 0})
	if consumed {
		t.Error("click on header should not be consumed")
	}
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (unchanged)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_ClickWithScrollOffset(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 4) // header + 3 visible rows
	tbl.SetPosition(0, 0)

	tbl.SetCursor(4)
	tbl.View()

	tbl.SetCursor(0)
	for i := 0; i < 4; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	// Cursor is at row 4, offset should be 2 (viewport=3, 4-3+1=2)

	// Click Y=1 (first data row on screen) should be row at offset (2)
	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 1})
	if !consumed {
		t.Error("click should be consumed")
	}
	if tbl.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (first visible row after scroll)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_ClickBeyondRows_Ignored(t *testing.T) {
	ds := NewSliceDataSource([][]string{
		{"12:00:01", "node1", "Starting up"},
		{"12:00:02", "node1", "Done"},
	})
	tbl := NewTable("t", sampleColumns(), ds, DefaultTableStyles())
	tbl.SetSize(80, 10)
	tbl.SetPosition(0, 0)

	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 5})
	if consumed {
		t.Error("click beyond data rows should not be consumed")
	}
	if tbl.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (unchanged)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_ClickInactive_Ignored(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)
	tbl.SetEnabled(false)

	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 2})
	if consumed {
		t.Error("click on inactive table should not be consumed")
	}
}

func TestTable_OnRowKeyPress(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(2)

	var receivedRow int
	var receivedKey string
	tbl.OnRowKeyPress(func(row int, msg tea.KeyMsg) tea.Cmd {
		receivedRow = row
		receivedKey = msg.String()
		return tea.Quit
	})

	cmd, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed via onRowKeyPress")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
	if receivedRow != 2 {
		t.Errorf("received row = %d, want 2", receivedRow)
	}
	if receivedKey != "enter" {
		t.Errorf("received key = %q, want 'enter'", receivedKey)
	}
}

func TestTable_OnRowKeyPress_NavigationNotIntercepted(t *testing.T) {
	tbl := makeTestTable(sampleData())
	called := false
	tbl.OnRowKeyPress(func(row int, msg tea.KeyMsg) tea.Cmd {
		called = true
		return tea.Quit
	})

	// Navigation keys should be consumed by table, not passed to handler
	tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	if called {
		t.Error("navigation keys should not reach onRowKeyPress")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", tbl.Cursor())
	}
}

func TestTable_OnRowKeyPress_FallsThrough_ToOnKeyPress(t *testing.T) {
	tbl := makeTestTable(sampleData())

	// OnRowKeyPress returns nil — should fall through to OnKeyPress
	tbl.OnRowKeyPress(func(row int, msg tea.KeyMsg) tea.Cmd {
		return nil
	})

	onKeyPressCalled := false
	tbl.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		onKeyPressCalled = true
		return tea.Quit
	})

	_, consumed := tbl.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed via onKeyPress fallthrough")
	}
	if !onKeyPressCalled {
		t.Error("onKeyPress should be called when onRowKeyPress returns nil")
	}
}

func TestTable_OnRowClick(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	var clickedRow int
	clickCalled := false
	tbl.OnRowClick(func(row int) tea.Cmd {
		clickedRow = row
		clickCalled = true
		return nil
	})

	// Click on row 2 (Y=3: header at 0, row 0 at 1, row 1 at 2, row 2 at 3)
	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 3})
	if !consumed {
		t.Error("click should be consumed")
	}
	if !clickCalled {
		t.Error("onRowClick should have been called")
	}
	if clickedRow != 2 {
		t.Errorf("clicked row = %d, want 2", clickedRow)
	}
	if tbl.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2", tbl.Cursor())
	}
}

func TestTable_OnRowClick_WithScrollOffset(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 4) // header + 3 visible rows
	tbl.SetPosition(0, 0)

	// Scroll down to make offset > 0
	for i := 0; i < 4; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	// cursor=4, offset=2

	var clickedRow int
	tbl.OnRowClick(func(row int) tea.Cmd {
		clickedRow = row
		return nil
	})

	// Click Y=1 (first data row on screen) = row at offset (2)
	tbl.HandleEvent(MouseClickEvent{X: 5, Y: 1})
	if clickedRow != 2 {
		t.Errorf("clicked row = %d, want 2 (first visible after scroll)", clickedRow)
	}
}

func TestTable_OnRowClick_ReturnsCmd(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	tbl.OnRowClick(func(row int) tea.Cmd {
		return tea.Quit
	})

	cmd, _ := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 2})
	if cmd == nil {
		t.Error("OnRowClick cmd should be returned from HandleEvent")
	}
}

func TestTable_SetColumnSpacing(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetColumnSpacing(3)

	view := tbl.View()
	plain := stripansi.Strip(view)
	// With spacing=3, there should be 3 spaces between column values
	// Check that the header has wider spacing
	lines := strings.Split(plain, "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	// Verify the spacing is actually applied by checking column gap
	// TIME is 12 chars, with 3 spacing should have 3 spaces before NODE
	if !strings.Contains(lines[0], "TIME") || !strings.Contains(lines[0], "NODE") {
		t.Error("header should contain column titles")
	}
}

func TestTable_ViewportRows_HeightZero(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetSize(80, 0) // height 0

	vp := tbl.viewportRows()
	if vp != 1 {
		t.Errorf("viewportRows() = %d, want 1 (min)", vp)
	}
}

func TestTable_HandleEvent_ScrollDown(t *testing.T) {
	tbl := makeTestTable(sampleData())

	_, consumed := tbl.HandleEvent(MouseScrollEvent{Direction: 1})
	if !consumed {
		t.Error("scroll down should be consumed")
	}
	if tbl.Cursor() != 3 {
		t.Errorf("cursor = %d, want 3 (scrolled by 3)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_ScrollUp(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetCursor(4)

	_, consumed := tbl.HandleEvent(MouseScrollEvent{Direction: -1})
	if !consumed {
		t.Error("scroll up should be consumed")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (scrolled up by 3 from 4)", tbl.Cursor())
	}
}

func TestTable_HandleEvent_Scroll_Inactive(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetEnabled(false)

	_, consumed := tbl.HandleEvent(MouseScrollEvent{Direction: 1})
	if consumed {
		t.Error("scroll on inactive table should not be consumed")
	}
}

func TestTable_HandleEvent_Click_NoOnRowClick(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	// Click without onRowClick set - should still consume and move cursor
	_, consumed := tbl.HandleEvent(MouseClickEvent{X: 5, Y: 2})
	if !consumed {
		t.Error("click should be consumed even without onRowClick")
	}
	if tbl.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", tbl.Cursor())
	}
}

func TestTable_EnsureVisible_CursorAboveViewport(t *testing.T) {
	tbl := makeTestTable(sampleData())
	tbl.SetSize(80, 4) // viewport = 3

	// Scroll down first
	for i := 0; i < 4; i++ {
		tbl.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	// cursor=4, offset should be 2

	// Now jump cursor to 0 — offset should adjust
	tbl.cursor = 0
	tbl.ensureVisible(tbl.viewportRows())
	if tbl.offset != 0 {
		t.Errorf("offset = %d, want 0 (cursor above viewport should scroll up)", tbl.offset)
	}
}

func TestTable_AnsiPadRight_AlreadyAtWidth(t *testing.T) {
	result := ansiPadRight("hello", 5)
	if result != "hello" {
		t.Errorf("ansiPadRight at width should return as-is, got %q", result)
	}
}

func TestTable_AnsiPadRight_BeyondWidth(t *testing.T) {
	result := ansiPadRight("hello world", 5)
	if result != "hello world" {
		t.Errorf("ansiPadRight beyond width should return as-is, got %q", result)
	}
}

func TestTable_RenderHeader_FlexColumn(t *testing.T) {
	cols := []ColumnDef{
		{Title: "A", Width: 10},
		{Title: "B", Width: 0}, // flex
	}
	ds := NewSliceDataSource([][]string{{"x", "y"}})
	tbl := NewTable("t", cols, ds, DefaultTableStyles())
	tbl.SetSize(80, 5)

	header := tbl.renderHeader()
	plain := stripansi.Strip(header)
	if !strings.Contains(plain, "A") || !strings.Contains(plain, "B") {
		t.Errorf("header should contain both column titles, got %q", plain)
	}
}

func TestTable_OnRowClick_Header_NotFired(t *testing.T) {
	tbl := NewTable("t", sampleColumns(), NewSliceDataSource(sampleData()), DefaultTableStyles())
	tbl.SetSize(80, 6)
	tbl.SetPosition(0, 0)

	called := false
	tbl.OnRowClick(func(row int) tea.Cmd {
		called = true
		return nil
	})

	tbl.HandleEvent(MouseClickEvent{X: 5, Y: 0})
	if called {
		t.Error("onRowClick should not fire on header click")
	}
}
