package widget

import "github.com/charmbracelet/lipgloss"

// CardStyles defines the styles used by a Card component.
type CardStyles struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Value  lipgloss.Style
	Alert  lipgloss.Style // for values that exceed thresholds
}

// DefaultCardStyles returns a CardStyles with sensible defaults.
func DefaultCardStyles() CardStyles {
	return CardStyles{
		Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#555555")).Padding(0, 1),
		Title:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Value:  lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0")).Bold(true),
		Alert:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ef5350")).Bold(true),
	}
}

// Card is a non-focusable bordered display component showing a title and value.
type Card struct {
	BaseComponent
	title  string
	value  string
	alert  bool
	styles CardStyles
}

// NewCard creates a new Card with the given ID, title, value, and styles.
func NewCard(id, title, value string, styles CardStyles) *Card {
	c := &Card{
		title:  title,
		value:  value,
		styles: styles,
	}
	c.BaseComponent = *NewBaseComponent(id)
	return c
}

// SetValue updates the card's displayed value.
func (c *Card) SetValue(value string) { c.value = value }

// Value returns the card's current value.
func (c *Card) Value() string { return c.value }

// SetAlert sets whether the card value should be rendered with the alert style.
func (c *Card) SetAlert(alert bool) { c.alert = alert }

// View renders a bordered box with title on the first line and value on the second.
func (c *Card) View() string {
	titleStr := c.styles.Title.Render(c.title)

	var valueStr string
	if c.alert {
		valueStr = c.styles.Alert.Render(c.value)
	} else {
		valueStr = c.styles.Value.Render(c.value)
	}

	content := titleStr + "\n" + valueStr

	var rendered string
	if !c.Active() {
		rendered = lipgloss.NewStyle().Faint(true).Render(
			c.styles.Border.Render(content),
		)
	} else {
		rendered = c.styles.Border.Render(content)
	}

	return c.RenderInSize(rendered)
}
