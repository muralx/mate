package widget

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabBarStyles defines the styles used by a TabBar in different states.
type TabBarStyles struct {
	Active   lipgloss.Style // selected tab
	Inactive lipgloss.Style // non-selected tab
	Focused  lipgloss.Style // focused inactive tab (keyboard cursor on it but not the selected tab)
}

// DefaultTabBarStyles returns a TabBarStyles with sensible defaults.
func DefaultTabBarStyles() TabBarStyles {
	return TabBarStyles{
		Active:   lipgloss.NewStyle().Background(lipgloss.Color("#2a2a3e")).Foreground(lipgloss.Color("#e0e0e0")).Bold(true).Padding(0, 2),
		Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 2),
		Focused:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Bold(true).Padding(0, 2),
	}
}

// TabBar is a focusable horizontal tab selector component.
// It has two distinct states:
//   - cursor: which tab the keyboard highlight is on (moves with left/right)
//   - active: which tab's content is currently displayed (changes on space/enter)
type TabBar struct {
	FocusableComponent
	labels      []string
	cursor      int // which tab has the visual cursor
	active      int // which tab is selected (content displayed)
	styles      TabBarStyles
	onChange    func(int) tea.Cmd
	tabBindings map[int]key.Binding // per-tab accelerator bindings, keyed by tab index
}

// NewTabBar creates a new TabBar with the given ID, labels, and styles.
func NewTabBar(id string, labels []string, styles TabBarStyles) *TabBar {
	tb := &TabBar{
		labels: labels,
		styles: styles,
	}
	tb.FocusableComponent = NewFocusableComponent(id)
	return tb
}

// OnChange sets the callback invoked when the active tab changes (on space/enter).
func (tb *TabBar) OnChange(fn func(int) tea.Cmd) { tb.onChange = fn }

// ActiveTab returns the index of the currently active (selected) tab.
func (tb *TabBar) ActiveTab() int { return tb.active }

// SetActiveTab sets the active tab index and moves the cursor to match.
func (tb *TabBar) SetActiveTab(i int) {
	tb.active = i
	tb.cursor = i
}

// CursorTab returns the index the keyboard cursor is on.
func (tb *TabBar) CursorTab() int { return tb.cursor }

// SetFocused overrides to reset cursor to active tab when gaining focus.
func (tb *TabBar) SetFocused(v bool) tea.Cmd {
	cmd := tb.BaseComponent.SetFocused(v)
	if v {
		tb.cursor = tb.active // cursor starts on the active tab
	}
	return cmd
}

// Update handles key input.
// Left/right moves the cursor. Space/enter activates the tab under the cursor.
func (tb *TabBar) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !tb.Active() {
		return nil, false
	}
	switch msg.String() {
	case "left", "h":
		if tb.cursor > 0 {
			tb.cursor--
		}
		return nil, true
	case "right", "l":
		if tb.cursor < len(tb.labels)-1 {
			tb.cursor++
		}
		return nil, true
	case " ", "enter":
		if tb.cursor != tb.active {
			tb.active = tb.cursor
			if tb.onChange != nil {
				return tb.onChange(tb.active), true
			}
		}
		return nil, true
	}
	if tb.onKeyPress != nil {
		if cmd := tb.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// BindDefaultActionToKey registers a global key binding that triggers the tab bar's
// default action (activate the tab under cursor).
func (tb *TabBar) BindDefaultActionToKey(keys string, description ...string) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	tb.RegisterKeyBinding(keys, desc, func() tea.Cmd {
		if tb.cursor != tb.active {
			tb.active = tb.cursor
			if tb.onChange != nil {
				return tb.onChange(tb.active)
			}
		}
		return nil
	})
}

// SetTabKeyBinding binds a keyboard shortcut to activate a specific tab by index.
// The keys parameter is a key combo string (e.g. "ctrl+d").
// The optional description is used for help text; if omitted, the tab's label is used.
// Panics if index is out of range.
// If a binding already exists for this index, it is replaced.
func (tb *TabBar) SetTabKeyBinding(index int, keys string, description ...string) {
	if index < 0 || index >= len(tb.labels) {
		panic(fmt.Sprintf("SetTabKeyBinding: index %d out of range [0, %d)", index, len(tb.labels)))
	}
	// Remove previous binding for this index if any
	if prev, ok := tb.tabBindings[index]; ok {
		tb.RemoveKeyBinding(prev)
	}
	if tb.tabBindings == nil {
		tb.tabBindings = make(map[int]key.Binding)
	}
	desc := tb.labels[index]
	if len(description) > 0 {
		desc = description[0]
	}
	binding := key.NewBinding(key.WithKeys(keys), key.WithHelp(keys, desc))
	tb.tabBindings[index] = binding
	tb.RegisterKeyBinding(keys, desc, func() tea.Cmd {
		if tb.active == index {
			return nil
		}
		tb.active = index
		tb.cursor = index
		if tb.onChange != nil {
			return tb.onChange(index)
		}
		return nil
	})
}

// View renders the tabs horizontally joined with the appropriate styles.
//
// Rendering rules:
//   - Active tab (content displayed): Active style
//   - Cursor tab (when focused and cursor != active): Focused style (yellow, highlighted)
//   - Active + cursor (focused, cursor on active tab): Active style with underline
//   - Other tabs: Inactive style
func (tb *TabBar) View() string {
	var parts []string
	for i, label := range tb.labels {
		switch {
		case i == tb.active && i == tb.cursor && tb.Focused():
			// Cursor is on the active tab while focused: underline to show focus
			parts = append(parts, tb.styles.Active.Underline(true).Render(label))
		case i == tb.cursor && tb.Focused():
			// Cursor is on an inactive tab: highlight it (user can press space to select)
			parts = append(parts, tb.styles.Focused.Render(label))
		case i == tb.active:
			// Active tab (selected), no cursor on it
			parts = append(parts, tb.styles.Active.Render(label))
		default:
			// Inactive tab
			parts = append(parts, tb.styles.Inactive.Render(label))
		}
	}
	result := strings.Join(parts, "")
	return tb.RenderInSize(result)
}

// HandleEvent handles mouse clicks by activating the clicked tab.
// The click X is relative to the tab bar's position and mapped to a tab
// using the rendered width of each label (including style padding).
func (tb *TabBar) HandleEvent(event Event) (tea.Cmd, bool) {
	if click, ok := event.(MouseClickEvent); ok {
		if !tb.Active() {
			return nil, false
		}
		px, _ := tb.Position()
		relX := click.X - px

		// Determine which tab was clicked by accumulating rendered widths
		offset := 0
		for i, label := range tb.labels {
			// Render with inactive style to get consistent width (all styles use same padding)
			w := lipgloss.Width(tb.styles.Inactive.Render(label))
			if relX >= offset && relX < offset+w {
				if tb.active == i {
					return nil, true // already active, consumed but no callback
				}
				tb.active = i
				tb.cursor = i
				if tb.onChange != nil {
					return tb.onChange(i), true
				}
				return nil, true
			}
			offset += w
		}
		return nil, false // click outside any tab
	}
	return tb.BaseComponent.HandleEvent(event)
}
