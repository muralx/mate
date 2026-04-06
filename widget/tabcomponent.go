package widget

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabComponent is a container that manages tab switching.
// It uses TCB layout internally: the TabBar header at the top,
// the active tab's content panel at the center (fills remaining space).
//
// Usage:
//
//	tabs := widget.NewTabComponent("tabs", themeTabBarStyles())
//	tabs.AddTab("Dashboard", dashPanel)
//	tabs.AddTab("Servers", serverPanel)
//	tabs.AddTab("Settings", settingsPanel)
//	tabs.SetTabKeyBinding(0, "ctrl+d")
//	win.Add(tabs, widget.TCBCenter)
type TabComponent struct {
	BaseContainer
	bar      *TabBar
	panels   []Component
	active   int
	onChange func(int) tea.Cmd
}

// NewTabComponent creates a TabComponent with the given ID and tab bar styles.
func NewTabComponent(id string, styles TabBarStyles) *TabComponent {
	tc := &TabComponent{}
	tc.BaseContainer = *NewBaseContainer(id, tc)
	tc.bar = NewTabBar(id+"-bar", nil, styles)
	tc.bar.OnChange(func(index int) tea.Cmd {
		tc.activate(index)
		if tc.onChange != nil {
			return tc.onChange(index)
		}
		return nil
	})
	tc.AddChild(tc.bar)
	return tc
}

// AddTab adds a tab with the given label and content panel.
func (tc *TabComponent) AddTab(label string, content Component) {
	tc.bar.labels = append(tc.bar.labels, label)
	tc.panels = append(tc.panels, content)
	tc.AddChild(content)

	// First tab added becomes active
	if len(tc.panels) == 1 {
		tc.activate(0)
	} else {
		content.SetVisible(false)
	}
}

// SetTabKeyBinding binds a keyboard shortcut to activate a specific tab.
func (tc *TabComponent) SetTabKeyBinding(index int, keys string, description ...string) {
	tc.bar.SetTabKeyBinding(index, keys, description...)
}

// OnChange sets a callback invoked when the active tab changes.
func (tc *TabComponent) OnChange(fn func(int) tea.Cmd) { tc.onChange = fn }

// ActiveTab returns the index of the currently active tab.
func (tc *TabComponent) ActiveTab() int { return tc.active }

// SetActiveTab switches to the tab at the given index.
func (tc *TabComponent) SetActiveTab(index int) {
	tc.activate(index)
	tc.bar.SetActiveTab(index)
}

// TabBar returns the underlying TabBar leaf component.
func (tc *TabComponent) TabBar() *TabBar { return tc.bar }

// activate switches to the tab at the given index.
func (tc *TabComponent) activate(index int) {
	if index < 0 || index >= len(tc.panels) {
		return
	}
	for i, p := range tc.panels {
		p.SetVisible(i == index)
	}
	tc.active = index
}

// View renders the tab header at the top and the active tab's content
// filling the remaining height.
func (tc *TabComponent) View() string {
	if !tc.Visible() {
		return ""
	}

	w, h := tc.Size()

	// Layout: TabBar at top, active panel fills remaining
	tc.bar.SetPreferredHeight(1)
	var activePanel Component
	for _, p := range tc.panels {
		if p.Visible() {
			activePanel = p
			break
		}
	}

	px, py := tc.Position()
	LayoutTCB(tc.bar, activePanel, nil, px, py, w, h)

	// Render
	barView := tc.bar.View()
	if activePanel != nil {
		panelView := activePanel.View()
		return lipgloss.JoinVertical(lipgloss.Left, barView, panelView)
	}
	return tc.RenderInSize(barView)
}
