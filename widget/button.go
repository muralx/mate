package widget

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ButtonStyles defines the styles used by a Button in different states.
type ButtonStyles struct {
	Normal  lipgloss.Style
	Focused lipgloss.Style
}

// DefaultButtonStyles returns a ButtonStyles with sensible defaults.
func DefaultButtonStyles() ButtonStyles {
	return ButtonStyles{
		Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ef5350")).Bold(true),
		Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Background(lipgloss.Color("#3a3a5e")).Bold(true),
	}
}

// DefaultPopupButtonStyles returns a ButtonStyles with defaults for popup buttons.
func DefaultPopupButtonStyles() ButtonStyles {
	return ButtonStyles{
		Normal:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Background(lipgloss.Color("#3a3a5e")).Bold(true),
	}
}

// Button is a focusable action button component.
type Button struct {
	FocusableComponent
	label   string
	styles  ButtonStyles
	onPress func() tea.Cmd
}

// NewButton creates a new Button with the given ID, label, and styles.
func NewButton(id, label string, styles ButtonStyles) *Button {
	b := &Button{
		label:  label,
		styles: styles,
	}
	b.FocusableComponent = NewFocusableComponent(id)
	return b
}

// OnPress sets the callback invoked when the button is pressed.
func (b *Button) OnPress(fn func() tea.Cmd) { b.onPress = fn }

// Update handles key input. Space and enter trigger the onPress callback.
func (b *Button) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !b.Active() {
		return nil, false
	}
	switch msg.String() {
	case " ", "enter":
		if b.onPress != nil {
			return b.onPress(), true
		}
		return nil, true
	}
	if b.onKeyPress != nil {
		if cmd := b.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// View renders the button label with the appropriate style for its state.
func (b *Button) View() string {
	text := "[ " + b.label + " ]"
	var rendered string
	if !b.Active() {
		rendered = lipgloss.NewStyle().Faint(true).Render(text)
	} else if b.Focused() {
		rendered = b.styles.Focused.Render(text)
	} else {
		rendered = b.styles.Normal.Render(text)
	}
	return b.RenderInSize(rendered)
}

// HandleEvent handles high-level events. MouseClick fires onPress.
func (b *Button) HandleEvent(event Event) (tea.Cmd, bool) {
	if _, ok := event.(MouseClickEvent); ok {
		if !b.Active() {
			return nil, false
		}
		if b.onPress != nil {
			return b.onPress(), true
		}
		return nil, true
	}
	return b.FocusableComponent.HandleEvent(event)
}

// BindDefaultActionToKey registers a global key binding that triggers the button's
// default action (onPress).
func (b *Button) BindDefaultActionToKey(keys string, description ...string) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	b.RegisterKeyBinding(keys, desc, func() tea.Cmd {
		if b.onPress != nil {
			return b.onPress()
		}
		return nil
	})
}
