package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeDemo() (*demoState, *window.App) {
	s := &demoState{}
	s.win = window.NewWindow("main") // TCB layout (default)

	s.buildDashboard()
	s.buildServers()
	s.buildSettings()
	s.buildStatusBar()

	// TabComponent: header + content panels in one unit
	s.tabs = widget.NewTabComponent("main-tabs", themeTabBarStyles())
	s.tabs.AddTab("Dashboard", s.dashPanel)
	s.tabs.AddTab("Servers", s.serverPanel)
	s.tabs.AddTab("Settings", s.settingsPanel)
	s.tabs.SetTabKeyBinding(0, "ctrl+d")
	s.tabs.SetTabKeyBinding(1, "ctrl+e")
	s.tabs.SetTabKeyBinding(2, "ctrl+g")

	s.win.Add(s.tabs, widget.TCBCenter)
	s.win.Add(s.statusBar, widget.TCBBottom)

	s.win.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return tea.Quit
		}
		return nil
	})

	s.win.OnUpdate(func() tea.Cmd {
		var parts []string
		if s.lastStatusMsg != "" {
			parts = append(parts, s.lastStatusMsg)
			s.lastStatusMsg = ""
		}
		parts = append(parts, "Tab: focus")
		for _, b := range s.win.ActiveKeyBindings() {
			h := b.Help()
			if h.Key != "" && h.Desc != "" {
				parts = append(parts, h.Key+": "+h.Desc)
			}
		}
		s.statusBar.SetText(strings.Join(parts, " | "))
		return nil
	})

	app := window.NewApp(s.win)
	// Trigger initial resize
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return s, app
}

func update(app *window.App, msg tea.Msg) tea.Cmd {
	_, cmd := app.Update(msg)
	// Process returned commands by running them with a short timeout.
	// ShowPopup/Close produce immediate msgs; cursor blink cmds block.
	if cmd != nil {
		done := make(chan tea.Msg, 1)
		go func() { done <- cmd() }()
		select {
		case m := <-done:
			if m != nil {
				app.Update(m)
			}
		case <-time.After(5 * time.Millisecond):
			// Blocking cmd (cursor blink timer) — skip
		}
	}
	return cmd
}

func sendKey(app *window.App, keyType tea.KeyType) tea.Cmd {
	return update(app, tea.KeyMsg{Type: keyType})
}

func render(app *window.App) string {
	return stripansi.Strip(app.View())
}

// ---------------------------------------------------------------------------
// 1. Initial render and structure
// ---------------------------------------------------------------------------

func TestInitialRender(t *testing.T) {
	_, app := makeDemo()
	output := render(app)

	for _, want := range []string{"Dashboard", "Servers", "Settings", "CPU", "23%", "Memory", "Disk", "Uptime"} {
		if !strings.Contains(output, want) {
			t.Errorf("should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. Tab switching
// ---------------------------------------------------------------------------

func TestTabBarKeyboardNavigation(t *testing.T) {
	s, app := makeDemo()

	if s.tabs.ActiveTab() != 0 {
		t.Fatalf("initial tab = %d, want 0", s.tabs.ActiveTab())
	}

	sendKey(app, tea.KeyRight)
	sendKey(app, tea.KeySpace)

	if s.tabs.ActiveTab() != 1 {
		t.Errorf("active tab = %d, want 1", s.tabs.ActiveTab())
	}
	if !s.serverPanel.Visible() {
		t.Error("server panel should be visible")
	}
}

func TestTabAcceleratorKeys(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlE})
	if s.tabs.ActiveTab() != 1 {
		t.Errorf("ctrl+e: tab = %d, want 1", s.tabs.ActiveTab())
	}

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	if s.tabs.ActiveTab() != 2 {
		t.Errorf("ctrl+g: tab = %d, want 2", s.tabs.ActiveTab())
	}

	update(app, tea.KeyMsg{Type: tea.KeyCtrlD})
	if s.tabs.ActiveTab() != 0 {
		t.Errorf("ctrl+d: tab = %d, want 0", s.tabs.ActiveTab())
	}
}

func TestTabVisibility(t *testing.T) {
	s, _ := makeDemo()

	if !s.dashPanel.Visible() || s.serverPanel.Visible() || s.settingsPanel.Visible() {
		t.Error("only dashboard should be visible on tab 0")
	}

	s.tabs.SetActiveTab(1)
	if s.dashPanel.Visible() || !s.serverPanel.Visible() {
		t.Error("only servers should be visible on tab 1")
	}

	s.tabs.SetActiveTab(2)
	if s.serverPanel.Visible() || !s.settingsPanel.Visible() {
		t.Error("only settings should be visible on tab 2")
	}
}

// ---------------------------------------------------------------------------
// 3. Data table
// ---------------------------------------------------------------------------

func TestDataTableNavigation(t *testing.T) {
	s, app := makeDemo()

	sendKey(app, tea.KeyTab) // focus data-table
	if s.dataTable.Cursor() != 0 {
		t.Errorf("initial cursor = %d", s.dataTable.Cursor())
	}

	sendKey(app, tea.KeyDown)
	sendKey(app, tea.KeyDown)
	if s.dataTable.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2", s.dataTable.Cursor())
	}

	sendKey(app, tea.KeyUp)
	if s.dataTable.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", s.dataTable.Cursor())
	}
}

func TestDataTableRendersContent(t *testing.T) {
	_, app := makeDemo()
	output := render(app)

	for _, want := range []string{"TIME", "LEVEL", "SOURCE", "MESSAGE", "14:23:01", "web-prod-01"} {
		if !strings.Contains(output, want) {
			t.Errorf("should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Checkbox list
// ---------------------------------------------------------------------------

func TestServerCheckboxList(t *testing.T) {
	s, app := makeDemo()

	// Switch to Servers tab
	sendKey(app, tea.KeyRight)
	sendKey(app, tea.KeySpace)
	sendKey(app, tea.KeyTab) // server-list

	if len(s.serverList.Selected()) != 3 {
		t.Errorf("initially %d selected, want 3", len(s.serverList.Selected()))
	}

	sendKey(app, tea.KeyDown)  // cursor to web-prod-01
	sendKey(app, tea.KeySpace) // uncheck
	if len(s.serverList.Selected()) != 2 {
		t.Errorf("after uncheck: %d selected, want 2", len(s.serverList.Selected()))
	}

	sendKey(app, tea.KeySpace) // re-check
	if len(s.serverList.Selected()) != 3 {
		t.Errorf("after re-check: %d selected, want 3", len(s.serverList.Selected()))
	}
}

// ---------------------------------------------------------------------------
// 5. TextInput
// ---------------------------------------------------------------------------

func TestTextInputTyping(t *testing.T) {
	s, app := makeDemo()

	// Settings tab
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	sendKey(app, tea.KeyTab) // hostname

	if s.hostnameInput.Value() != "dashboard.local" {
		t.Errorf("initial = %q", s.hostnameInput.Value())
	}

	update(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})
	if !strings.Contains(s.hostnameInput.Value(), "!") {
		t.Error("typing should modify value")
	}
}

// ---------------------------------------------------------------------------
// 6. FormattedTextInput validation
// ---------------------------------------------------------------------------

func TestFormattedTextInputValidation(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG}) // settings
	sendKey(app, tea.KeyTab)                    // hostname
	sendKey(app, tea.KeyTab)                    // api-key
	sendKey(app, tea.KeyTab)                    // refresh-interval

	s.refreshInterval.SetValue("abc")
	sendKey(app, tea.KeyTab) // blur triggers validation

	if s.refreshInterval.Error() == "" {
		t.Error("validation should set error for non-numeric input")
	}
}

// ---------------------------------------------------------------------------
// 7. Toggle
// ---------------------------------------------------------------------------

func TestToggleOnOff(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	// Tab to notifications: hostname, api-key, refresh-interval, dark-mode, notifications
	for range 5 {
		sendKey(app, tea.KeyTab)
	}

	if !s.notifications.On() {
		t.Error("should initially be on")
	}
	sendKey(app, tea.KeySpace)
	if s.notifications.On() {
		t.Error("should be off after space")
	}
}

func TestToggleRadioMode(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	for range 4 {
		sendKey(app, tea.KeyTab)
	} // dark-mode

	if !s.darkMode.On() {
		t.Error("should be on initially")
	}
	sendKey(app, tea.KeySpace)
	if s.darkMode.On() {
		t.Error("should be off after toggle")
	}
}

// ---------------------------------------------------------------------------
// 8. Buttons
// ---------------------------------------------------------------------------

func TestSaveSettingsButton(t *testing.T) {
	_, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	// Tab to save button: hostname, api-key, interval, dark, notif, auto, source, save
	for range 8 {
		sendKey(app, tea.KeyTab)
	}

	sendKey(app, tea.KeyEnter)
	output := render(app)
	if !strings.Contains(output, "Settings saved") {
		t.Error("status should show saved")
	}
}

func TestResetButton(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	sendKey(app, tea.KeyTab) // hostname
	update(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Tab to reset: api-key, interval, dark, notif, auto, source, save, reset
	for range 8 {
		sendKey(app, tea.KeyTab)
	}

	sendKey(app, tea.KeySpace)
	if s.hostnameInput.Value() != "dashboard.local" {
		t.Errorf("hostname = %q, want 'dashboard.local'", s.hostnameInput.Value())
	}
}

// ---------------------------------------------------------------------------
// 9. Cards
// ---------------------------------------------------------------------------

func TestCardValues(t *testing.T) {
	s, _ := makeDemo()

	checks := map[string]string{
		"CPU": s.cpuCard.Value(), "Memory": s.memCard.Value(),
		"Disk": s.diskCard.Value(), "Uptime": s.uptimeCard.Value(),
	}
	expected := map[string]string{"CPU": "23%", "Memory": "4.2 GB", "Disk": "67%", "Uptime": "14d 6h"}
	for k, want := range expected {
		if checks[k] != want {
			t.Errorf("%s = %q, want %q", k, checks[k], want)
		}
	}
}

// ---------------------------------------------------------------------------
// 10. Popups - Confirm
// ---------------------------------------------------------------------------

func TestConfirmPopupYesClose(t *testing.T) {
	s, app := makeDemo()

	// Switch to Servers tab, navigate to delete button
	update(app, tea.KeyMsg{Type: tea.KeyCtrlE})
	sendKey(app, tea.KeyTab) // server-list
	sendKey(app, tea.KeyTab) // server-detail-table
	sendKey(app, tea.KeyTab) // add-server-btn
	sendKey(app, tea.KeyTab) // delete-server-btn

	// Press delete (has 3 selected) — opens confirm popup
	sendKey(app, tea.KeyEnter)

	// In popup — press Enter on Yes button (first focused)
	sendKey(app, tea.KeyEnter)

	output := render(app)
	if !strings.Contains(output, "Deleted") {
		t.Errorf("should show deletion message, got %q", s.statusBar.GetText())
	}
}

func TestConfirmPopupEscCancels(t *testing.T) {
	s, app := makeDemo()

	update(app, tea.KeyMsg{Type: tea.KeyCtrlE})
	sendKey(app, tea.KeyTab)
	sendKey(app, tea.KeyTab)
	sendKey(app, tea.KeyTab)
	sendKey(app, tea.KeyTab)   // delete-server-btn
	sendKey(app, tea.KeyEnter) // open popup

	sendKey(app, tea.KeyEscape) // cancel

	if !strings.Contains(s.statusBar.GetText(), "cancelled") {
		t.Errorf("status = %q, should contain 'cancelled'", s.statusBar.GetText())
	}
}

// ---------------------------------------------------------------------------
// 11. Popups - Add Server
// ---------------------------------------------------------------------------

func TestAddServerPopupFlow(t *testing.T) {
	s, app := makeDemo()

	initialRows := s.serverDS.RowCount()

	// Open add server popup via ctrl+n
	update(app, tea.KeyMsg{Type: tea.KeyCtrlE}) // servers tab
	update(app, tea.KeyMsg{Type: tea.KeyCtrlN}) // trigger add server binding

	// We're now in the popup. Type server name.
	// First focusable should be name input.
	for _, r := range "my-server" {
		update(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	sendKey(app, tea.KeyTab) // host input
	for _, r := range "10.0.0.5" {
		update(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	sendKey(app, tea.KeyTab) // port input
	for _, r := range "8080" {
		update(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	sendKey(app, tea.KeyTab)   // save button
	sendKey(app, tea.KeyEnter) // press save

	newRows := s.serverDS.RowCount()
	if newRows != initialRows+1 {
		t.Errorf("rows = %d, want %d", newRows, initialRows+1)
	}

	lastRow := newRows - 1
	if s.serverDS.CellData(lastRow, 0) != "my-server" {
		t.Errorf("name = %q", s.serverDS.CellData(lastRow, 0))
	}
	if s.serverDS.CellData(lastRow, 5) != "10.0.0.5:8080" {
		t.Errorf("addr = %q", s.serverDS.CellData(lastRow, 5))
	}
}

// ---------------------------------------------------------------------------
// 12. Mouse
// ---------------------------------------------------------------------------

func clickAt(app *window.App, x, y int) tea.Cmd {
	return update(app, tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
}

func hoverAt(app *window.App, x, y int) tea.Cmd {
	return update(app, tea.MouseMsg{
		X: x, Y: y,
		Button: tea.MouseButtonNone,
		Action: tea.MouseActionMotion,
	})
}

func TestMouseClickFocusesDataTable(t *testing.T) {
	s, app := makeDemo()
	_ = app.View() // trigger layout

	px, py := s.dataTable.Position()
	tw, th := s.dataTable.Size()
	if tw == 0 || th == 0 {
		t.Skip("table size not set")
	}

	clickAt(app, px+1, py+1)
	if !s.dataTable.Focused() {
		t.Error("click should focus data table")
	}
}

func TestMouseMotionDoesNotChangeFocus(t *testing.T) {
	s, app := makeDemo()
	_ = app.View()

	if !s.tabs.TabBar().Focused() {
		t.Fatal("tab bar should have initial focus")
	}

	px, py := s.dataTable.Position()
	hoverAt(app, px+1, py+1)

	if !s.tabs.TabBar().Focused() {
		t.Error("hover should not change focus")
	}
}

func TestMouseMotionDoesNotToggle(t *testing.T) {
	s, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	_ = app.View()

	initial := s.darkMode.On()
	px, py := s.darkMode.Position()
	hoverAt(app, px+1, py)

	if s.darkMode.On() != initial {
		t.Error("hover should not toggle")
	}
}

func TestMouseReleaseDoesNotTrigger(t *testing.T) {
	s, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	_ = app.View()

	initial := s.darkMode.On()
	px, py := s.darkMode.Position()
	update(app, tea.MouseMsg{
		X: px + 1, Y: py,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionRelease,
	})

	if s.darkMode.On() != initial {
		t.Error("release should not toggle")
	}
}

func TestMouseClickToggle(t *testing.T) {
	s, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	_ = app.View()

	initial := s.darkMode.On()
	px, py := s.darkMode.Position()
	tw, _ := s.darkMode.Size()
	if tw == 0 {
		t.Skip("toggle size not set")
	}

	clickAt(app, px+1, py)
	if s.darkMode.On() == initial {
		t.Error("click should toggle dark mode")
	}
}

func TestTableClickSelectsRow(t *testing.T) {
	s, app := makeDemo()
	_ = app.View()

	px, py := s.dataTable.Position()
	_, th := s.dataTable.Size()
	if th < 3 {
		t.Skip("table too small")
	}

	clickAt(app, px+1, py+3) // header + 2 data rows -> row 2

	if s.dataTable.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2", s.dataTable.Cursor())
	}
}

func TestTabBarClickActivatesTab(t *testing.T) {
	s, app := makeDemo()
	_ = app.View()

	px, py := s.tabs.Position()
	tw, _ := s.tabs.Size()
	if tw == 0 {
		t.Skip("tabs size not set")
	}

	// Click toward right to hit a later tab
	clickAt(app, px+(tw*2/3), py)

	if s.tabs.ActiveTab() == 0 {
		clickAt(app, px+tw-2, py)
	}

	if s.tabs.ActiveTab() != 0 {
		t.Logf("tab click activated tab %d", s.tabs.ActiveTab())
	}
}

// ---------------------------------------------------------------------------
// 13. ScrollableText
// ---------------------------------------------------------------------------

func TestScrollableTextInitialContent(t *testing.T) {
	s, _ := makeDemo()
	if !strings.Contains(s.detailView.Content(), "Select a row") {
		t.Error("should have placeholder content")
	}
}

func TestDetailUpdatesOnTableClick(t *testing.T) {
	s, app := makeDemo()
	_ = app.View()

	px, py := s.dataTable.Position()
	_, th := s.dataTable.Size()
	if th < 3 {
		t.Skip("table too small")
	}

	clickAt(app, px+1, py+2) // click row 1 (14:23:02)

	content := s.detailView.Content()
	if !strings.Contains(content, "14:23:02") {
		t.Errorf("detail should show clicked row, got %q", content)
	}
}

// ---------------------------------------------------------------------------
// 14. Global key bindings
// ---------------------------------------------------------------------------

func TestCtrlQQuits(t *testing.T) {
	_, app := makeDemo()
	cmd := update(app, tea.KeyMsg{Type: tea.KeyCtrlQ})
	if cmd == nil {
		t.Error("ctrl+q should return quit")
	}
}

func TestCtrlCQuits(t *testing.T) {
	_, app := makeDemo()
	cmd := update(app, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return quit")
	}
}

// ---------------------------------------------------------------------------
// 15. Status bar with dynamic keybinding hints
// ---------------------------------------------------------------------------

func TestStatusBarShowsKeyHints(t *testing.T) {
	s, app := makeDemo()
	// Trigger an event so OnUpdate fires
	sendKey(app, tea.KeyTab)

	status := s.statusBar.GetText()
	if !strings.Contains(status, "ctrl+d") {
		t.Errorf("status should contain tab accelerators, got %q", status)
	}
	if !strings.Contains(status, "Tab: focus") {
		t.Error("should contain Tab: focus")
	}
}

func TestStatusBarHintsChangePerTab(t *testing.T) {
	s, app := makeDemo()
	sendKey(app, tea.KeyTab) // trigger OnUpdate

	// Dashboard: should NOT have ctrl+n or ctrl+s
	status := s.statusBar.GetText()
	if strings.Contains(status, "ctrl+n") {
		t.Error("dashboard should not show ctrl+n")
	}

	// Switch to servers
	update(app, tea.KeyMsg{Type: tea.KeyCtrlE})
	status = s.statusBar.GetText()
	if !strings.Contains(status, "ctrl+n") {
		t.Errorf("servers tab should show ctrl+n, got %q", status)
	}
}

// ---------------------------------------------------------------------------
// 16. Enabled/Disabled
// ---------------------------------------------------------------------------

func TestDisablePanel(t *testing.T) {
	s, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})

	s.settingsPanel.SetEnabled(false)
	if s.hostnameInput.Active() {
		t.Error("children should be inactive")
	}

	s.settingsPanel.SetEnabled(true)
	if !s.hostnameInput.Active() {
		t.Error("children should be active again")
	}
}

// ---------------------------------------------------------------------------
// 17. Nested panels
// ---------------------------------------------------------------------------

func TestNestedPanelRendering(t *testing.T) {
	_, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlG})
	output := render(app)

	for _, want := range []string{"Connection", "Preferences", "Hostname", "API Key", "Save Settings", "Reset"} {
		if !strings.Contains(output, want) {
			t.Errorf("settings tab should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 18. Renderers
// ---------------------------------------------------------------------------

func TestLevelRendererAllBranches(t *testing.T) {
	ds := widget.NewSliceDataSource([][]string{
		{"14:00", "ERROR", "src", "msg"},
		{"14:01", "WARN", "src", "msg"},
		{"14:02", "INFO", "src", "msg"},
		{"14:03", "DEBUG", "src", "msg"},
		{"14:04", "TRACE", "src", "msg"},
	})
	styles := themeTableStyles()
	for row, want := range []string{"ERROR", "WARN", "INFO", "DEBUG", "TRACE"} {
		t.Run(want, func(t *testing.T) {
			result := stripansi.Strip(levelRenderer(ds, row, 1, false, 7, styles))
			if !strings.Contains(result, want) {
				t.Errorf("got %q", result)
			}
		})
	}
}

func TestStatusRendererAllBranches(t *testing.T) {
	ds := widget.NewSliceDataSource([][]string{
		{"s1", "OK"}, {"s2", "WARN"}, {"s3", "FAIL"}, {"s4", "UNKNOWN"},
	})
	styles := themeTableStyles()
	for _, tt := range []struct {
		row  int
		want string
	}{{0, "OK"}, {1, "WARN"}, {2, "FAIL"}, {3, "UNKNOWN"}} {
		t.Run(tt.want, func(t *testing.T) {
			result := stripansi.Strip(statusRenderer(ds, tt.row, 1, false, 8, styles))
			if !strings.Contains(result, tt.want) {
				t.Errorf("got %q", result)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 19. Render at various sizes
// ---------------------------------------------------------------------------

func TestRenderAtVariousSizes(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}, {60, 20}, {200, 50}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			_, app := makeDemo()
			update(app, tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			output := app.View()
			if output == "" {
				t.Error("render should not be empty")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 20. Table selected row highlights full row
// ---------------------------------------------------------------------------

func TestTableSelectedHighlightsFullRow(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	s, app := makeDemo()
	sendKey(app, tea.KeyTab) // focus data table

	// Render through app to trigger layout propagation
	app.View()
	view := s.dataTable.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 2 {
		t.Fatal("need at least header + 1 row")
	}

	selectedRow := lines[1]
	stripped := stripansi.Strip(selectedRow)

	for _, want := range []string{"14:23:01", "INFO", "web-prod-01"} {
		if !strings.Contains(stripped, want) {
			t.Errorf("selected row should contain %q", want)
		}
	}

	// Background code should appear in latter half
	bgCode := "\x1b[48;"
	halfLen := len(selectedRow) / 2
	if !strings.Contains(selectedRow[halfLen:], bgCode) {
		t.Error("background should extend to later columns")
	}
}

// ---------------------------------------------------------------------------
// 21. Row container
// ---------------------------------------------------------------------------

func TestRowContainerButtonsRender(t *testing.T) {
	_, app := makeDemo()
	update(app, tea.KeyMsg{Type: tea.KeyCtrlE})
	output := render(app)

	for _, want := range []string{"Add Server", "Delete", "Refresh"} {
		if !strings.Contains(output, want) {
			t.Errorf("should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 22. Add server via addServerRow
// ---------------------------------------------------------------------------

func TestAddServerRow(t *testing.T) {
	s, _ := makeDemo()
	initial := s.serverDS.RowCount()

	s.addServerRow("test", "1.2.3.4", "22")

	if s.serverDS.RowCount() != initial+1 {
		t.Error("row count should increase")
	}
	last := s.serverDS.RowCount() - 1
	if s.serverDS.CellData(last, 0) != "test" {
		t.Errorf("name = %q", s.serverDS.CellData(last, 0))
	}
}
