package widget

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CheckboxListStyles defines the styles used by a CheckboxList.
type CheckboxListStyles struct {
	Cursor    lipgloss.Style // "> " marker on current item
	Checked   lipgloss.Style // "[x] " checked indicator
	Unchecked lipgloss.Style // "[ ] " unchecked indicator
	Item      lipgloss.Style // normal item label
	Group     lipgloss.Style // group item label
	Dim       lipgloss.Style // non-cursor items prefix ("  ")
}

// DefaultCheckboxListStyles returns a CheckboxListStyles with sensible defaults.
func DefaultCheckboxListStyles() CheckboxListStyles {
	return CheckboxListStyles{
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Bold(true),
		Checked:   lipgloss.NewStyle().Foreground(lipgloss.Color("#66bb6a")),
		Unchecked: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Item:      lipgloss.NewStyle(),
		Group:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	}
}

// CheckboxItem represents a single item in a CheckboxList.
type CheckboxItem struct {
	Label   string
	Value   string
	Checked bool
	IsGroup bool
}

// CheckboxList is a focusable list of checkable items.
type CheckboxList struct {
	FocusableComponent
	items    []CheckboxItem
	cursor   int
	styles   CheckboxListStyles
	onChange func([]CheckboxItem) tea.Cmd
}

// NewCheckboxList creates a new CheckboxList with the given ID, items, and styles.
func NewCheckboxList(id string, items []CheckboxItem, styles CheckboxListStyles) *CheckboxList {
	cl := &CheckboxList{
		items:  items,
		styles: styles,
	}
	cl.FocusableComponent = NewFocusableComponent(id)
	return cl
}

// OnChange sets the callback invoked when an item is toggled.
func (cl *CheckboxList) OnChange(fn func([]CheckboxItem) tea.Cmd) { cl.onChange = fn }

// Items returns the current items.
func (cl *CheckboxList) Items() []CheckboxItem { return cl.items }

// Cursor returns the current cursor position.
func (cl *CheckboxList) Cursor() int { return cl.cursor }

// Selected returns the values of all checked items.
func (cl *CheckboxList) Selected() []string {
	var sel []string
	for _, item := range cl.items {
		if item.Checked {
			sel = append(sel, item.Value)
		}
	}
	return sel
}

// Update handles key input for navigation and toggling.
func (cl *CheckboxList) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !cl.Active() {
		return nil, false
	}
	switch msg.String() {
	case "up", "k":
		if cl.cursor > 0 {
			cl.cursor--
		}
		return nil, true
	case "down", "j":
		if cl.cursor < len(cl.items)-1 {
			cl.cursor++
		}
		return nil, true
	case " ":
		if cl.cursor >= 0 && cl.cursor < len(cl.items) {
			cl.items[cl.cursor].Checked = !cl.items[cl.cursor].Checked
			if cl.onChange != nil {
				return cl.onChange(cl.items), true
			}
		}
		return nil, true
	}
	if cl.onKeyPress != nil {
		if cmd := cl.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// BindDefaultActionToKey registers a global key binding that triggers the checkbox's
// default action (toggle the item at cursor).
func (cl *CheckboxList) BindDefaultActionToKey(keys string, description ...string) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	cl.RegisterKeyBinding(keys, desc, func() tea.Cmd {
		if cl.cursor >= 0 && cl.cursor < len(cl.items) {
			cl.items[cl.cursor].Checked = !cl.items[cl.cursor].Checked
			if cl.onChange != nil {
				return cl.onChange(cl.items)
			}
		}
		return nil
	})
}

// View renders the checkbox list.
func (cl *CheckboxList) View() string {
	var lines []string
	for i, item := range cl.items {
		var prefix string
		if cl.Focused() && i == cl.cursor {
			prefix = cl.styles.Cursor.Render("> ")
		} else {
			prefix = cl.styles.Dim.Render("  ")
		}

		var checkbox string
		if item.Checked {
			checkbox = cl.styles.Checked.Render("[x] ")
		} else {
			checkbox = cl.styles.Unchecked.Render("[ ] ")
		}

		var label string
		if item.IsGroup {
			label = cl.styles.Group.Render(item.Label)
		} else {
			label = cl.styles.Item.Render(item.Label)
		}

		lines = append(lines, fmt.Sprintf("%s%s%s", prefix, checkbox, label))
	}

	output := strings.Join(lines, "\n")
	if !cl.Active() {
		output = lipgloss.NewStyle().Faint(true).Render(output)
	}
	return cl.RenderInSize(output)
}

// HandleEvent handles high-level events. MouseClick toggles cursor item.
func (cl *CheckboxList) HandleEvent(event Event) (tea.Cmd, bool) {
	if click, ok := event.(MouseClickEvent); ok {
		if !cl.Active() {
			return nil, false
		}
		_, cy := cl.Position()
		idx := click.Y - cy
		if idx >= 0 && idx < len(cl.items) {
			cl.cursor = idx
			cl.items[idx].Checked = !cl.items[idx].Checked
			if cl.onChange != nil {
				return cl.onChange(cl.items), true
			}
			return nil, true
		}
		return nil, false
	}
	return cl.BaseComponent.HandleEvent(event)
}
