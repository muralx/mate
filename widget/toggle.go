package widget

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToggleMode controls how the toggle is rendered.
type ToggleMode int

const (
	// ToggleModeOnOff renders one value at a time: "Feature:[on]" or "Feature:[off]".
	ToggleModeOnOff ToggleMode = iota
	// ToggleModeRadio renders both labels side by side: "[Live] [Cache]".
	ToggleModeRadio
)

// ToggleStyles defines the styles used by a Toggle in different states.
type ToggleStyles struct {
	Label       lipgloss.Style // label prefix
	OnActive    lipgloss.Style // on value, unfocused (green)
	OnFocused   lipgloss.Style // on value, focused (yellow)
	OffActive   lipgloss.Style // off value, unfocused (dim for OnOff, orange for Radio)
	OffFocused  lipgloss.Style // off value, focused (yellow)
	OffInactive lipgloss.Style // off value, not selected (dim) — used in Radio mode
}

// DefaultToggleStyles returns a ToggleStyles with sensible defaults.
func DefaultToggleStyles() ToggleStyles {
	return ToggleStyles{
		Label:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		OnActive:    lipgloss.NewStyle().Foreground(lipgloss.Color("#81c784")),
		OnFocused:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Bold(true),
		OffActive:   lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		OffFocused:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ffeb3b")).Bold(true),
		OffInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
	}
}

// Toggle is a boolean on/off component with two rendering modes.
type Toggle struct {
	FocusableComponent
	label    string // prefix, e.g., "Feature" → rendered as "Feature:"
	onLabel  string // e.g., "[on]" or "[Live]"
	offLabel string // e.g., "[off]" or "[Cache]"
	mode     ToggleMode
	on       bool
	styles   ToggleStyles
	onChange func(bool) tea.Cmd
}

// NewToggle creates a new Toggle with the given ID, label, initial state, mode, and styles.
func NewToggle(id, label string, initial bool, mode ToggleMode, styles ToggleStyles) *Toggle {
	t := &Toggle{
		label:    label,
		onLabel:  "[on]",
		offLabel: "[off]",
		mode:     mode,
		on:       initial,
		styles:   styles,
	}
	t.FocusableComponent = NewFocusableComponent(id)
	return t
}

// SetLabels sets the on/off display labels.
func (t *Toggle) SetLabels(onLabel, offLabel string) {
	t.onLabel = onLabel
	t.offLabel = offLabel
}

// OnChange sets the callback invoked when the toggle state changes.
func (t *Toggle) OnChange(fn func(bool) tea.Cmd) { t.onChange = fn }

// On returns the current toggle state.
func (t *Toggle) On() bool { return t.on }

// SetOn sets the toggle state.
func (t *Toggle) SetOn(v bool) { t.on = v }

// Update handles key input. Space and enter toggle the state.
func (t *Toggle) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !t.Active() {
		return nil, false
	}
	switch msg.String() {
	case " ", "enter":
		t.on = !t.on
		if t.onChange != nil {
			return t.onChange(t.on), true
		}
		return nil, true
	}
	if t.onKeyPress != nil {
		if cmd := t.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// BindDefaultActionToKey registers a global key binding that triggers the toggle's
// default action (toggle state).
func (t *Toggle) BindDefaultActionToKey(keys string, description ...string) {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	t.RegisterKeyBinding(keys, desc, func() tea.Cmd {
		t.on = !t.on
		if t.onChange != nil {
			return t.onChange(t.on)
		}
		return nil
	})
}

// View renders the toggle with the appropriate style for its state.
func (t *Toggle) View() string {
	var rendered string
	if !t.Active() {
		rendered = t.viewInactive()
	} else {
		switch t.mode {
		case ToggleModeRadio:
			rendered = t.viewRadio()
		default:
			rendered = t.viewOnOff()
		}
	}
	return t.RenderInSize(rendered)
}

// viewOnOff renders "Label:[value]" with padding to keep width stable.
func (t *Toggle) viewOnOff() string {
	var prefix string
	if t.label != "" {
		prefix = t.styles.Label.Render(t.label + ":")
	}

	// Pad value text to the longer label length for width stability.
	maxLen := len(t.onLabel)
	if len(t.offLabel) > maxLen {
		maxLen = len(t.offLabel)
	}

	var value string
	if t.on {
		padded := t.onLabel + strings.Repeat(" ", maxLen-len(t.onLabel))
		if t.Focused() {
			value = t.styles.OnFocused.Render(padded)
		} else {
			value = t.styles.OnActive.Render(padded)
		}
	} else {
		padded := t.offLabel + strings.Repeat(" ", maxLen-len(t.offLabel))
		if t.Focused() {
			value = t.styles.OffFocused.Render(padded)
		} else {
			value = t.styles.OffActive.Render(padded)
		}
	}

	return prefix + value
}

// viewRadio renders both labels side by side with the active one highlighted.
func (t *Toggle) viewRadio() string {
	var prefix string
	if t.label != "" {
		prefix = t.styles.Label.Render(t.label + ":")
	}

	var onPart, offPart string
	if t.on {
		if t.Focused() {
			onPart = t.styles.OnFocused.Render(t.onLabel)
		} else {
			onPart = t.styles.OnActive.Render(t.onLabel)
		}
		offPart = t.styles.OffInactive.Render(t.offLabel)
	} else {
		onPart = t.styles.OffInactive.Render(t.onLabel)
		if t.Focused() {
			offPart = t.styles.OffFocused.Render(t.offLabel)
		} else {
			offPart = t.styles.OffActive.Render(t.offLabel)
		}
	}

	return prefix + onPart + " " + offPart
}

// viewInactive renders the toggle in a faint style.
func (t *Toggle) viewInactive() string {
	var prefix string
	if t.label != "" {
		prefix = t.label + ":"
	}

	var text string
	switch t.mode {
	case ToggleModeRadio:
		text = prefix + t.onLabel + " " + t.offLabel
	default:
		// Pad to the longer label for consistent width.
		maxLen := len(t.onLabel)
		if len(t.offLabel) > maxLen {
			maxLen = len(t.offLabel)
		}
		if t.on {
			text = prefix + t.onLabel + strings.Repeat(" ", maxLen-len(t.onLabel))
		} else {
			text = prefix + t.offLabel + strings.Repeat(" ", maxLen-len(t.offLabel))
		}
	}

	return lipgloss.NewStyle().Faint(true).Render(text)
}

// HandleEvent handles high-level events. MouseClick toggles state.
func (t *Toggle) HandleEvent(event Event) (tea.Cmd, bool) {
	if _, ok := event.(MouseClickEvent); ok {
		if !t.Active() {
			return nil, false
		}
		t.on = !t.on
		if t.onChange != nil {
			cmd := t.onChange(t.on)
			return cmd, true
		}
		return nil, true
	}
	return t.BaseComponent.HandleEvent(event)
}
