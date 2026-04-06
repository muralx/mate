package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

var (
	colorPrimary = lipgloss.Color("#4fc3f7")
	colorDim     = lipgloss.Color("#888888")
	colorText    = lipgloss.Color("#e0e0e0")
	colorAccent  = lipgloss.Color("#ce93d8")
	colorGreen   = lipgloss.Color("#81c784")
	colorYellow  = lipgloss.Color("#ffeb3b")
)

func main() {
	win := window.NewWindow("main")

	// --- Outer TabComponent ---
	tabs := widget.NewTabComponent("tabs", widget.TabBarStyles{
		Active:   lipgloss.NewStyle().Background(colorPrimary).Foreground(lipgloss.Color("#000")).Bold(true).Padding(0, 2),
		Inactive: lipgloss.NewStyle().Foreground(colorDim).Padding(0, 2),
		Focused:  lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Padding(0, 2),
	})

	// Tab 1: Simple content with TCB panel
	tab1 := widget.NewPanel("tab1", widget.TCB)
	tab1.SetBorder(widget.SingleLineBorder(string(colorDim), string(colorPrimary)))
	tab1.SetTitle("Overview")

	header := widget.NewPanel("header-row", widget.Horizontal)
	header.SetSpacing(2)
	for _, item := range []struct{ title, value string }{
		{"Users", "1,234"}, {"Sessions", "567"}, {"Errors", "12"}, {"Uptime", "99.9%"},
	} {
		card := widget.NewCard(item.title+"-card", item.title, item.value, widget.CardStyles{
			Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorDim).Padding(0, 1),
			Title:  lipgloss.NewStyle().Foreground(colorDim),
			Value:  lipgloss.NewStyle().Foreground(colorText).Bold(true),
			Alert:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ef5350")).Bold(true),
		})
		card.SetPreferredWidth(18)
		card.SetPreferredHeight(4)
		header.Add(card, widget.Next)
	}
	header.SetPreferredHeight(4)

	mainContent := widget.NewText("main-content", "Main content area — this fills the remaining space", lipgloss.NewStyle().Foreground(colorText))

	footer := widget.NewText("footer", "Footer: showing overview data", lipgloss.NewStyle().Foreground(colorDim))
	footer.SetPreferredHeight(1)

	tab1.Add(header, widget.TCBTop)
	tab1.Add(mainContent, widget.TCBCenter)
	tab1.Add(footer, widget.TCBBottom)

	// Tab 2: Nested TabComponent
	innerTabs := widget.NewTabComponent("inner-tabs", widget.TabBarStyles{
		Active:   lipgloss.NewStyle().Background(colorAccent).Foreground(lipgloss.Color("#000")).Bold(true).Padding(0, 2),
		Inactive: lipgloss.NewStyle().Foreground(colorDim).Padding(0, 2),
		Focused:  lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Padding(0, 2),
	})

	innerP1 := widget.NewPanel("inner-p1")
	innerP1.SetBorder(widget.SingleLineBorder(string(colorDim), string(colorAccent)))
	innerP1.SetTitle("Connection Settings")
	innerP1.Add(widget.NewField("host-field", "Hostname", widget.NewTextInput("host", 30), widget.DefaultFieldStyles()), widget.Next)
	innerP1.Add(widget.NewField("port-field", "Port", widget.NewTextInput("port", 8), widget.DefaultFieldStyles()), widget.Next)

	innerP2 := widget.NewPanel("inner-p2")
	innerP2.SetBorder(widget.SingleLineBorder(string(colorDim), string(colorAccent)))
	innerP2.SetTitle("Authentication")
	innerP2.Add(widget.NewField("user-field", "Username", widget.NewTextInput("user", 30), widget.DefaultFieldStyles()), widget.Next)
	innerP2.Add(widget.NewField("pass-field", "Password", widget.NewTextInput("pass", 30), widget.DefaultFieldStyles()), widget.Next)

	innerP3 := widget.NewPanel("inner-p3")
	innerP3.SetBorder(widget.SingleLineBorder(string(colorDim), string(colorAccent)))
	innerP3.SetTitle("Advanced")
	innerP3.Add(widget.NewField("timeout-field", "Timeout", widget.NewTextInput("timeout", 8), widget.DefaultFieldStyles()), widget.Next)

	innerTabs.AddTab("Connection", innerP1)
	innerTabs.AddTab("Auth", innerP2)
	innerTabs.AddTab("Advanced", innerP3)
	innerTabs.SetTabKeyBinding(0, "ctrl+1")
	innerTabs.SetTabKeyBinding(1, "ctrl+2")
	innerTabs.SetTabKeyBinding(2, "ctrl+3")

	// Tab 3: Simple vertical panel
	tab3 := widget.NewPanel("tab3")
	tab3.SetBorder(widget.SingleLineBorder(string(colorDim), string(colorGreen)))
	tab3.SetTitle("About")
	tab3.Add(widget.NewText("about-1", "Mate — A component framework for terminal UIs", lipgloss.NewStyle().Foreground(colorText).Bold(true)), widget.Next)
	tab3.Add(widget.NewText("about-2", "Built on top of Bubble Tea", lipgloss.NewStyle().Foreground(colorDim)), widget.Next)
	tab3.Add(widget.NewText("about-3", "", lipgloss.NewStyle()), widget.Next)
	tab3.Add(widget.NewText("about-4", "This demo shows nested TabComponents, TCB layout,", lipgloss.NewStyle().Foreground(colorText)), widget.Next)
	tab3.Add(widget.NewText("about-5", "horizontal panels, cards, and text inputs.", lipgloss.NewStyle().Foreground(colorText)), widget.Next)

	tabs.AddTab("Overview", tab1)
	tabs.AddTab("Settings", innerTabs)
	tabs.AddTab("About", tab3)
	tabs.SetTabKeyBinding(0, "ctrl+d")
	tabs.SetTabKeyBinding(1, "ctrl+e")
	tabs.SetTabKeyBinding(2, "ctrl+g")

	// Status bar
	statusBar := widget.NewText("status", "Tab: focus | ctrl+d: Overview | ctrl+e: Settings | ctrl+g: About | ctrl+q: Quit", lipgloss.NewStyle().Foreground(colorDim).Background(lipgloss.Color("#2a2a3e")))
	statusBar.SetPreferredHeight(1)

	win.Add(tabs, widget.TCBCenter)
	win.Add(statusBar, widget.TCBBottom)

	win.RegisterKeyBinding("ctrl+q", "Quit", func() tea.Cmd { return tea.Quit })
	win.RegisterKeyBinding("ctrl+c", "", func() tea.Cmd { return tea.Quit })

	app := window.NewApp(win)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
