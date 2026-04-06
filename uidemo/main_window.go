package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

// demoState holds references to components we need to interact with from callbacks.
type demoState struct {
	win       *window.MainWindow
	tabs      *widget.TabComponent
	statusBar *widget.Text

	// Dashboard
	cpuCard    *widget.Card
	memCard    *widget.Card
	diskCard   *widget.Card
	uptimeCard *widget.Card
	dataTable  *widget.Table
	detailView *widget.ScrollableText

	// Servers
	serverList      *widget.CheckboxList
	serverTable     *widget.Table
	serverDS        *widget.SliceDataSource
	serverData      [][]string
	addServerBtn    *widget.Button
	deleteServerBtn *widget.Button
	refreshBtn      *widget.Button

	// Settings
	hostnameInput   *widget.TextInput
	apiKeyInput     *widget.TextInput
	refreshInterval *widget.FormattedTextInput
	darkMode        *widget.Toggle
	notifications   *widget.Toggle
	autoRefresh     *widget.Toggle
	sourceToggle    *widget.Toggle
	saveBtn         *widget.Button
	resetBtn        *widget.Button

	// Tab panels
	dashPanel     *widget.Panel
	serverPanel   *widget.Panel
	settingsPanel *widget.Panel

	lastStatusMsg string
}

func buildMainWindow() *window.MainWindow {
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

	// Global quit binding
	s.win.RegisterKeyBinding("ctrl+q", "Quit", func() tea.Cmd { return tea.Quit })
	s.win.RegisterKeyBinding("ctrl+c", "", func() tea.Cmd { return tea.Quit })

	// Update status bar hints after every event
	s.win.OnUpdate(func() tea.Cmd {
		var parts []string
		if s.lastStatusMsg != "" {
			parts = append(parts, s.lastStatusMsg)
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

	return s.win
}

func (s *demoState) setStatus(msg string) {
	s.lastStatusMsg = msg
}

func (s *demoState) clearStatus() {
	s.lastStatusMsg = ""
}

// buildTabs and showTab are no longer needed — TabComponent handles tab switching

func (s *demoState) buildStatusBar() {
	style := lipgloss.NewStyle().Foreground(colorDim).Background(lipgloss.Color("#2a2a3e"))
	s.statusBar = widget.NewText("status-bar", "Ready", style)
	s.statusBar.SetPreferredHeight(1)
}

func (s *demoState) buildDashboard() {
	s.dashPanel = widget.NewPanel("dash-panel", widget.TCB)
	s.dashPanel.SetBorder(themeBorder())
	s.dashPanel.SetTitle("System Overview")

	// Cards in a horizontal row
	s.cpuCard = widget.NewCard("cpu-card", "CPU", "23%", themeCardStyles())
	s.cpuCard.SetPreferredWidth(18)
	s.cpuCard.SetPreferredHeight(4)
	s.memCard = widget.NewCard("mem-card", "Memory", "4.2 GB", themeCardStyles())
	s.memCard.SetPreferredWidth(18)
	s.memCard.SetPreferredHeight(4)
	s.diskCard = widget.NewCard("disk-card", "Disk", "67%", themeCardStyles())
	s.diskCard.SetPreferredWidth(18)
	s.diskCard.SetPreferredHeight(4)
	s.diskCard.SetAlert(true)
	s.uptimeCard = widget.NewCard("uptime-card", "Uptime", "14d 6h", themeCardStyles())
	s.uptimeCard.SetPreferredWidth(18)
	s.uptimeCard.SetPreferredHeight(4)

	cardRow := widget.NewPanel("card-row", widget.Horizontal)
	cardRow.SetSpacing(1)
	cardRow.SetPreferredHeight(4)
	cardRow.Add(s.cpuCard, widget.Next)
	cardRow.Add(s.memCard, widget.Next)
	cardRow.Add(s.diskCard, widget.Next)
	cardRow.Add(s.uptimeCard, widget.Next)
	s.dashPanel.Add(cardRow, widget.TCBTop)

	// Data table — fills center (remaining space)
	columns := []widget.ColumnDef{
		{Title: "TIME", Width: 12},
		{Title: "LEVEL", Width: 7, Renderer: levelRenderer},
		{Title: "SOURCE", Width: 15},
		{Title: "MESSAGE", Width: 0},
	}
	tableDS := widget.NewSliceDataSource(generateSampleData())
	s.dataTable = widget.NewTable("data-table", columns, tableDS, themeTableStyles())
	s.dataTable.OnRowClick(func(row int) tea.Cmd {
		s.showRowDetail(tableDS, row)
		return nil
	})
	s.dataTable.OnRowKeyPress(func(row int, msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "enter" {
			s.showRowDetail(tableDS, row)
		}
		return nil
	})
	tablePanel := widget.NewPanel("table-panel", widget.TCB)
	tablePanel.SetBorder(themeBorder())
	tablePanel.Add(s.dataTable, widget.TCBCenter)
	s.dashPanel.Add(tablePanel, widget.TCBCenter)

	// Scrollable detail area — pinned to bottom
	s.detailView = widget.NewScrollableText("detail-view", widget.DefaultScrollableTextStyles())
	s.detailView.SetPreferredHeight(5)
	s.detailView.SetContent("Select a row to view details (click or Enter)")
	s.dashPanel.Add(s.detailView, widget.TCBBottom)

	// dashPanel is added to tabContent in buildMainWindow
}

func (s *demoState) showRowDetail(ds *widget.SliceDataSource, row int) {
	t := ds.CellData(row, 0)
	level := ds.CellData(row, 1)
	source := ds.CellData(row, 2)
	message := ds.CellData(row, 3)
	detail := fmt.Sprintf("Time:    %s\nLevel:   %s\nSource:  %s\nMessage: %s", t, level, source, message)
	s.detailView.SetContent(detail)
	s.setStatus(fmt.Sprintf("Viewing entry from %s", t))
}

func (s *demoState) buildServers() {
	s.serverPanel = widget.NewPanel("server-panel", widget.TCB)
	s.serverPanel.SetBorder(themeBorder())
	s.serverPanel.SetTitle("Server Management")

	// Server list with checkboxes
	serverListPanel := widget.NewPanel("server-list-panel")
	serverListPanel.SetBorder(themeBorder())
	serverListPanel.SetTitle("Servers")
	serverListPanel.SetPreferredWidth(40)

	items := []widget.CheckboxItem{
		{Label: "--- Production ---", Value: "", IsGroup: true},
		{Label: "web-prod-01", Value: "web-prod-01", Checked: true},
		{Label: "web-prod-02", Value: "web-prod-02", Checked: true},
		{Label: "api-prod-01", Value: "api-prod-01", Checked: true},
		{Label: "--- Staging ---", Value: "", IsGroup: true},
		{Label: "web-stg-01", Value: "web-stg-01", Checked: false},
		{Label: "api-stg-01", Value: "api-stg-01", Checked: false},
		{Label: "--- Development ---", Value: "", IsGroup: true},
		{Label: "dev-01", Value: "dev-01", Checked: false},
	}
	s.serverList = widget.NewCheckboxList("server-list", items, themeCheckboxListStyles())
	s.serverList.OnChange(func(items []widget.CheckboxItem) tea.Cmd {
		selected := s.serverList.Selected()
		s.setStatus(fmt.Sprintf("Selected %d servers", len(selected)))
		return nil
	})
	serverListPanel.Add(s.serverList, widget.Next)
	s.serverPanel.Add(serverListPanel, widget.TCBTop)

	// Server details table
	detailCols := []widget.ColumnDef{
		{Title: "SERVER", Width: 16},
		{Title: "STATUS", Width: 8, Renderer: statusRenderer},
		{Title: "CPU", Width: 6},
		{Title: "MEM", Width: 8},
		{Title: "UPTIME", Width: 10},
		{Title: "IP", Width: 0},
	}
	s.serverData = [][]string{
		{"web-prod-01", "OK", "34%", "2.1 GB", "14d 6h", "10.0.1.10"},
		{"web-prod-02", "OK", "28%", "1.9 GB", "14d 6h", "10.0.1.11"},
		{"api-prod-01", "WARN", "78%", "3.8 GB", "7d 2h", "10.0.1.20"},
		{"web-stg-01", "OK", "12%", "0.8 GB", "3d 1h", "10.0.2.10"},
		{"api-stg-01", "FAIL", "0%", "0 GB", "0h", "10.0.2.20"},
		{"dev-01", "OK", "45%", "1.2 GB", "1d 3h", "10.0.3.10"},
	}
	s.serverDS = widget.NewSliceDataSource(s.serverData)
	s.serverTable = widget.NewTable("server-detail-table", detailCols, s.serverDS, themeTableStyles())
	// No preferred height — center slot stretches
	s.serverTable.OnRowClick(func(row int) tea.Cmd {
		name := s.serverDS.CellData(row, 0)
		status := s.serverDS.CellData(row, 1)
		s.setStatus(fmt.Sprintf("Selected: %s (%s)", name, status))
		return nil
	})
	s.serverPanel.Add(s.serverTable, widget.TCBCenter)

	// Action buttons in a row
	s.addServerBtn = widget.NewButton("add-server-btn", "+ Add Server", themeSuccessButtonStyles())
	s.addServerBtn.OnPress(func() tea.Cmd {
		return s.win.ShowPopup(s.buildAddServerPopup())
	})
	s.addServerBtn.BindDefaultActionToKey("ctrl+n", "Add server")

	s.deleteServerBtn = widget.NewButton("delete-server-btn", "- Delete", themeDangerButtonStyles())
	s.deleteServerBtn.OnPress(func() tea.Cmd {
		selected := s.serverList.Selected()
		if len(selected) == 0 {
			return nil
		}
		return s.win.ShowPopup(s.buildConfirmPopup(
			fmt.Sprintf("Delete %d selected server(s)?", len(selected)),
			func() tea.Cmd {
				s.setStatus(fmt.Sprintf("Deleted %d servers", len(selected)))
				return nil
			},
		))
	})

	s.refreshBtn = widget.NewButton("refresh-btn", "Refresh", themeButtonStyles())
	s.refreshBtn.OnPress(func() tea.Cmd {
		s.setStatus("Refreshing server data...")
		return nil
	})
	s.refreshBtn.BindDefaultActionToKey("ctrl+r", "Refresh")

	btnRow := widget.NewPanel("server-btn-row", widget.Horizontal)
	btnRow.SetSpacing(2)
	btnRow.Add(s.addServerBtn, widget.Next)
	btnRow.Add(s.deleteServerBtn, widget.Next)
	btnRow.Add(s.refreshBtn, widget.Next)
	s.serverPanel.Add(btnRow, widget.TCBBottom)

	// serverPanel is added to tabContent in buildMainWindow
}

func (s *demoState) buildSettings() {
	s.settingsPanel = widget.NewPanel("settings-panel")
	s.settingsPanel.SetBorder(themeBorder())
	s.settingsPanel.SetTitle("Application Settings")

	// Connection sub-panel
	connPanel := widget.NewPanel("conn-panel")
	connPanel.SetBorder(themeBorder())
	connPanel.SetTitle("Connection")
	connPanel.SetPreferredHeight(8)

	s.hostnameInput = widget.NewTextInput("hostname", 35)
	s.hostnameInput.WithPlaceholder("dashboard.example.com")
	s.hostnameInput.SetValue("dashboard.local")
	connPanel.Add(widget.NewField("hostname-field", "Hostname", s.hostnameInput, themeFieldStyles()), widget.Next)

	s.apiKeyInput = widget.NewTextInput("api-key", 35)
	s.apiKeyInput.WithPlaceholder("sk-xxxxxxxxxxxxxxxx")
	s.apiKeyInput.WithCharLimit(64)
	connPanel.Add(widget.NewField("api-key-field", "API Key", s.apiKeyInput, themeFieldStyles()), widget.Next)

	s.refreshInterval = widget.NewFormattedTextInput("refresh-interval", 8)
	s.refreshInterval.WithPlaceholder("30")
	s.refreshInterval.SetValue("30")
	s.refreshInterval.WithValidation(func(val string) error {
		if val == "" {
			return nil
		}
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if n < 5 || n > 3600 {
			return fmt.Errorf("must be 5-3600 seconds")
		}
		return nil
	})
	connPanel.Add(widget.NewField("interval-field", "Refresh (s)", s.refreshInterval, themeFieldStyles()), widget.Next)

	s.settingsPanel.Add(connPanel, widget.Next)

	// Preferences sub-panel
	togglePanel := widget.NewPanel("toggle-panel")
	togglePanel.SetBorder(themeBorder())
	togglePanel.SetTitle("Preferences")
	togglePanel.SetPreferredHeight(8)

	s.darkMode = widget.NewToggle("dark-mode", "Theme", true, widget.ToggleModeRadio, themeToggleStyles())
	s.darkMode.SetLabels("[Dark]", "[Light]")
	s.darkMode.OnChange(func(on bool) tea.Cmd {
		if on {
			s.setStatus("Switched to dark theme")
		} else {
			s.setStatus("Switched to light theme")
		}
		return nil
	})
	togglePanel.Add(widget.NewField("dark-mode-field", "Theme", s.darkMode, themeFieldStyles()), widget.Next)

	s.notifications = widget.NewToggle("notifications", "Alerts", true, widget.ToggleModeOnOff, themeToggleStyles())
	s.notifications.OnChange(func(on bool) tea.Cmd {
		if on {
			s.setStatus("Notifications enabled")
		} else {
			s.setStatus("Notifications disabled")
		}
		return nil
	})
	togglePanel.Add(widget.NewField("notif-field", "Alerts", s.notifications, themeFieldStyles()), widget.Next)

	s.autoRefresh = widget.NewToggle("auto-refresh", "Auto-Refresh", true, widget.ToggleModeOnOff, themeToggleStyles())
	s.autoRefresh.BindDefaultActionToKey("ctrl+a", "Toggle auto-refresh")
	togglePanel.Add(widget.NewField("auto-field", "Auto-Refresh", s.autoRefresh, themeFieldStyles()), widget.Next)

	s.sourceToggle = widget.NewToggle("source-toggle", "Source", false, widget.ToggleModeRadio, themeToggleStyles())
	s.sourceToggle.SetLabels("[Remote]", "[Local]")
	togglePanel.Add(widget.NewField("source-field", "Data Source", s.sourceToggle, themeFieldStyles()), widget.Next)

	s.settingsPanel.Add(togglePanel, widget.Next)

	// Buttons in a row
	s.saveBtn = widget.NewButton("save-settings", "Save Settings", themeSuccessButtonStyles())
	s.saveBtn.OnPress(func() tea.Cmd {
		s.setStatus(fmt.Sprintf("Settings saved at %s", time.Now().Format("15:04:05")))
		return nil
	})
	s.saveBtn.BindDefaultActionToKey("ctrl+s", "Save settings")

	s.resetBtn = widget.NewButton("reset-settings", "Reset", themeDangerButtonStyles())
	s.resetBtn.OnPress(func() tea.Cmd {
		s.hostnameInput.SetValue("dashboard.local")
		s.apiKeyInput.SetValue("")
		s.refreshInterval.SetValue("30")
		s.darkMode.SetOn(true)
		s.notifications.SetOn(true)
		s.autoRefresh.SetOn(true)
		s.sourceToggle.SetOn(false)
		s.setStatus("Settings reset to defaults")
		return nil
	})

	settingsBtnRow := widget.NewPanel("settings-btn-row", widget.Horizontal)
	settingsBtnRow.SetSpacing(2)
	settingsBtnRow.Add(s.saveBtn, widget.Next)
	settingsBtnRow.Add(s.resetBtn, widget.Next)
	s.settingsPanel.Add(settingsBtnRow, widget.Next)

	// settingsPanel is added to tabContent in buildMainWindow
}

func (s *demoState) addServerRow(name, host, port string) {
	s.serverData = append(s.serverData, []string{name, "OK", "0%", "0 GB", "0h", host + ":" + port})
	s.serverDS.SetData(s.serverData)
}

// Helper functions

func levelRenderer(ds widget.TableDataSource, row, col int, selected bool, width int, styles widget.TableStyles) string {
	value := ds.CellData(row, col)
	style := styles.Cell
	if selected {
		style = styles.Selected.Width(width)
	}
	switch value {
	case "ERROR":
		style = style.Foreground(colorDanger).Bold(true)
	case "WARN":
		style = style.Foreground(colorWarning)
	case "INFO":
		style = style.Foreground(colorPrimary)
	case "DEBUG":
		style = style.Foreground(colorDim)
	}
	return style.Render(value)
}

func statusRenderer(ds widget.TableDataSource, row, col int, selected bool, width int, styles widget.TableStyles) string {
	value := ds.CellData(row, col)
	style := styles.Cell
	if selected {
		style = styles.Selected.Width(width)
	}
	switch value {
	case "OK":
		style = style.Foreground(colorSecondary)
	case "WARN":
		style = style.Foreground(colorWarning)
	case "FAIL":
		style = style.Foreground(colorDanger).Bold(true)
	}
	return style.Render(value)
}

func generateSampleData() [][]string {
	return [][]string{
		{"14:23:01", "INFO", "web-prod-01", "Request served: GET /api/v1/users"},
		{"14:23:02", "DEBUG", "api-prod-01", "Cache hit for key: user:1234"},
		{"14:23:03", "WARN", "api-prod-01", "Slow query detected: 2.3s on users table"},
		{"14:23:04", "ERROR", "api-stg-01", "Connection refused: database unreachable"},
		{"14:23:05", "INFO", "web-prod-02", "Health check passed"},
		{"14:23:06", "INFO", "web-prod-01", "Request served: POST /api/v1/orders"},
		{"14:23:07", "WARN", "web-stg-01", "TLS certificate expires in 7 days"},
		{"14:23:08", "ERROR", "api-stg-01", "Connection refused: database unreachable"},
		{"14:23:09", "INFO", "dev-01", "Build completed successfully"},
		{"14:23:10", "DEBUG", "api-prod-01", "Background task completed"},
		{"14:23:11", "INFO", "web-prod-01", "Request served: GET /api/v1/products"},
		{"14:23:12", "WARN", "api-prod-01", "Memory usage at 78%"},
		{"14:23:13", "INFO", "web-prod-02", "Deployed version v2.4.1"},
		{"14:23:14", "ERROR", "dev-01", "Test suite failed: 3 assertions"},
		{"14:23:15", "INFO", "web-prod-01", strings.Repeat("Long message ", 5)},
	}
}
