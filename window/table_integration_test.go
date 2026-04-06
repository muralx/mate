package window

import (
	"fmt"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/muralx/mate/widget"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// --- Test data helpers ---

func testColumns() []widget.ColumnDef {
	return []widget.ColumnDef{
		{Title: "TIME", Width: 12},
		{Title: "NODE", Width: 10},
		{Title: "LEVEL", Width: 7},
		{Title: "MESSAGE", Width: 0}, // flex
	}
}

func testRows(n int) [][]string {
	rows := make([][]string, n)
	for i := range rows {
		msg := "Short msg"
		if i%3 == 0 {
			msg = strings.Repeat("Very long message that exceeds column width ", 5)
		}
		if i%5 == 0 {
			msg = "EXACT"
		}
		rows[i] = []string{
			"14:23:" + padInt(i),
			"node-" + padInt(i%3),
			[]string{"INFO", "WARN", "ERROR"}[i%3],
			msg,
		}
	}
	return rows
}

func padInt(n int) string {
	return fmt.Sprintf("%02d", n)
}

// makeTablePopup creates a popup with a TCB panel containing a table in center.
func makeTablePopup(data [][]string) (*App, *widget.Table, *PopupWindow) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "Open", widget.DefaultButtonStyles()), widget.TCBCenter)
	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	popup := NewPopupWindow("popup", "Data View", DefaultPopupStyles())
	popup.SetPreferredWidth(100)
	popup.SetPreferredHeight(30)

	panel := widget.NewPanel("table-panel", widget.TCB)
	ds := widget.NewSliceDataSource(data)
	table := widget.NewTable("tbl", testColumns(), ds, widget.DefaultTableStyles())
	panel.Add(table, widget.TCBCenter)
	popup.Add(panel, widget.TCBCenter)

	win.ShowPopup(popup)
	app.View() // trigger layout

	return app, table, popup
}

// --- Tests ---

func TestTablePopup_NoRows(t *testing.T) {
	app, table, _ := makeTablePopup(nil)

	view := app.View()
	plain := stripansi.Strip(view)

	if !strings.Contains(plain, "No data") {
		t.Error("empty table should show 'No data'")
	}

	w, h := table.Size()
	if w == 0 {
		t.Errorf("table width = 0, should be set by layout")
	}
	if h == 0 {
		t.Errorf("table height = 0, should be set by layout")
	}
}

func TestTablePopup_FewRows_HeaderVisible(t *testing.T) {
	app, _, _ := makeTablePopup(testRows(3))

	view := app.View()
	plain := stripansi.Strip(view)

	// Headers should be visible
	for _, col := range []string{"TIME", "NODE", "LEVEL", "MESSAGE"} {
		if !strings.Contains(plain, col) {
			t.Errorf("header %q not visible", col)
		}
	}

	// All 3 rows should be visible
	for _, ts := range []string{"14:23:00", "14:23:01", "14:23:02"} {
		if !strings.Contains(plain, ts) {
			t.Errorf("row with %q not visible", ts)
		}
	}
}

func TestTablePopup_FewRows_ContentTruncated(t *testing.T) {
	app, table, _ := makeTablePopup(testRows(3))

	view := app.View()
	plain := stripansi.Strip(view)

	tw, _ := table.Size()
	if tw == 0 {
		t.Fatal("table width = 0")
	}

	// The long message should NOT appear in full (it's ~200 chars, table is ~100 wide)
	longMsg := strings.Repeat("Very long message that exceeds column width ", 5)
	if strings.Contains(plain, longMsg) {
		t.Error("long message should be truncated, not shown in full")
	}

	// Each rendered line in the table area should not exceed table width
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		lw := lipgloss.Width(line)
		if lw > 120 { // viewport width
			t.Errorf("line %d width = %d, exceeds viewport 120", i, lw)
		}
	}
}

func TestTablePopup_ManyRows_Scrolling(t *testing.T) {
	app, table, _ := makeTablePopup(testRows(100))

	app.View()

	_, th := table.Size()
	if th == 0 {
		t.Fatal("table height = 0")
	}

	// Viewport should be less than 100 rows
	viewport := th - 1 // minus header
	if viewport >= 100 {
		t.Fatalf("viewport = %d, should be less than 100 rows", viewport)
	}

	// Initial cursor at 0
	if table.Cursor() != 0 {
		t.Errorf("initial cursor = %d, want 0", table.Cursor())
	}

	// Navigate down past viewport
	for i := 0; i < viewport+5; i++ {
		app.Update(tea.KeyMsg{Type: tea.KeyTab}) // focus table first time
		app.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Cursor should have moved
	if table.Cursor() == 0 {
		t.Error("cursor should have moved after Down keys")
	}

	// View should still render correctly
	view := app.View()
	if view == "" {
		t.Error("view should not be empty after scrolling")
	}
}

func TestTablePopup_PageDown_PageUp(t *testing.T) {
	app, table, _ := makeTablePopup(testRows(200))

	// Focus the table
	app.View()
	// Tab to find the table (it's the only focusable in the popup)
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Page down
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	afterPgDown := table.Cursor()
	if afterPgDown == 0 {
		t.Error("PgDown should move cursor")
	}

	// Page up
	app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	afterPgUp := table.Cursor()
	if afterPgUp >= afterPgDown {
		t.Errorf("PgUp should move cursor back: was %d, now %d", afterPgDown, afterPgUp)
	}

	// Home
	app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if table.Cursor() != 0 {
		t.Errorf("Home: cursor = %d, want 0", table.Cursor())
	}

	// End
	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if table.Cursor() != 199 {
		t.Errorf("End: cursor = %d, want 199", table.Cursor())
	}
}

func TestTablePopup_RowWidth_MatchesTableWidth(t *testing.T) {
	app, table, _ := makeTablePopup(testRows(5))

	view := app.View()
	plain := stripansi.Strip(view)
	_ = plain

	tw, _ := table.Size()
	if tw == 0 {
		t.Fatal("table width = 0, layout didn't propagate")
	}

	// Render the table directly and check each line width
	tableView := table.View()
	tableLines := strings.Split(tableView, "\n")

	for i, line := range tableLines {
		lw := lipgloss.Width(line)
		if lw != tw {
			t.Errorf("table line %d width = %d, want %d (table width)", i, lw, tw)
			if i > 5 {
				t.Log("(skipping remaining lines)")
				break
			}
		}
	}
}

func TestTablePopup_HeaderAlwaysFirstLine(t *testing.T) {
	app, table, _ := makeTablePopup(testRows(100))

	// Scroll down
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	for i := 0; i < 50; i++ {
		app.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Render and check header is still first line of table
	tableView := table.View()
	firstLine := stripansi.Strip(strings.Split(tableView, "\n")[0])

	if !strings.Contains(firstLine, "TIME") {
		t.Errorf("header should always be first line, got: %q", firstLine)
	}
	if !strings.Contains(firstLine, "NODE") {
		t.Errorf("header should contain NODE, got: %q", firstLine)
	}
}

func TestTablePopup_LongMessages_RowsAreSingleLine(t *testing.T) {
	// Simulates real-world log viewer with very long messages (like the screenshot:
	// MQM session lines, Java stack traces, etc.)
	data := [][]string{
		{"08:00:56", "node-0", "INFO", "NioProcessor-5 o.a.a.AcceptorHandler] MQM session created: local/192.168.14.101:9875, class org.apache.mina.transport.socket.nio.NioSocketSession, remote/118.228.18.91:21008"},
		{"08:01:17", "node-1", "ERROR", "NioProcessor-5 o.a.a.AcceptorHandler] tested [1:126.29.76.96:6891]: java.net.ssl.SSLException: Dnpager close riate: State = NK RemoteAddress = MEM_MAP"},
		{"08:01:21", "node-0", "WARN", "bytesTransmit = 8 BytesProduced = 7 requestModifier = 8  java.net.ssl.SSLException: Dnpager close riate: State = DN RemoteAddress = MEM3_WRAP"},
		{"08:03:42", "node-2", "INFO", "NioProcessor-5 o.a.a.AcceptorHandler] MQM session created: local/192.168.14.101:9875, class org.apache.mina.transport.socket.nio.NioSocketSession, remote/110.8.76.144:24594"},
		{"08:04:56", "node-0", "INFO", "NioProcessor-5 o.a.a.AcceptorHandler] MQM session created: local/192.168.14.101:9875, class org.apache.mina.transport.socket.nio.NioSocketSession, remote/119.8.72.61:41308"},
	}

	app, table, _ := makeTablePopup(data)

	// Focus table and select a row
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app.Update(tea.KeyMsg{Type: tea.KeyDown})

	app.View()

	tableView := table.View()
	tw, th := table.Size()
	if tw == 0 {
		t.Fatal("table width = 0")
	}

	lines := strings.Split(tableView, "\n")

	// Table should have exactly th lines (header + rows + padding)
	if len(lines) != th {
		t.Errorf("table rendered %d lines, expected %d (height)", len(lines), th)
		for i, line := range lines {
			plain := stripansi.Strip(line)
			t.Logf("  line %d: w=%d len=%d %q", i, lipgloss.Width(line), len(plain), plain[:min(80, len(plain))])
		}
	}

	// Every line must be exactly table width — no overflow, no wrapping
	for i, line := range lines {
		lw := lipgloss.Width(line)
		if lw > tw {
			plain := stripansi.Strip(line)
			t.Errorf("line %d width = %d, exceeds table width %d: %q", i, lw, tw, plain[:min(80, len(plain))])
		}
	}

	// Each data row should be a single line (no embedded newlines from lipgloss wrapping)
	headerAndRows := lines[:min(len(data)+1, len(lines))] // header + data rows
	for i, line := range headerAndRows {
		if strings.Contains(line, "\n") {
			t.Errorf("line %d contains embedded newline (wrapping detected)", i)
		}
	}
}

func TestTablePopup_WidthPropagation(t *testing.T) {
	// Verify the full chain: Window → popup → panel(TCB) → table
	app, table, popup := makeTablePopup(testRows(5))

	app.View()

	pw, ph := popup.Size()
	tw, th := table.Size()

	if pw == 0 || ph == 0 {
		t.Errorf("popup size = (%d, %d), should not be zero", pw, ph)
	}
	if tw == 0 || th == 0 {
		t.Errorf("table size = (%d, %d), should not be zero", tw, th)
	}
	if tw > pw {
		t.Errorf("table width %d exceeds popup width %d", tw, tw)
	}

	t.Logf("popup=(%d,%d) table=(%d,%d)", pw, ph, tw, th)
}
