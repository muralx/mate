package widget

import "github.com/charmbracelet/lipgloss"

// BorderType defines the visual style of a panel's border.
type BorderType int

const (
	NoBorder      BorderType = iota // no border (default)
	RoundedBorder                   // rounded corners
)

// BorderConfig defines the border appearance for a Panel.
type BorderConfig struct {
	Type        BorderType
	Color       lipgloss.Color // border color when no inner focus
	ActiveColor lipgloss.Color // border color when inner focused
	Padding     int            // horizontal padding inside border (columns per side)
}

// SingleLineBorder creates a BorderConfig with rounded border style.
func SingleLineBorder(color, activeColor string) BorderConfig {
	return BorderConfig{
		Type:        RoundedBorder,
		Color:       lipgloss.Color(color),
		ActiveColor: lipgloss.Color(activeColor),
		Padding:     1,
	}
}

// DefaultBorder returns the default border configuration.
func DefaultBorder() BorderConfig {
	return SingleLineBorder("#0f3460", "#4fc3f7")
}

// HasBorder returns true if the border type is not NoBorder.
func (bc BorderConfig) HasBorder() bool {
	return bc.Type != NoBorder
}

// ChromeWidth returns the total horizontal chrome (border + padding) in columns.
func (bc BorderConfig) ChromeWidth() int {
	if !bc.HasBorder() {
		return 0
	}
	return 2 + bc.Padding*2 // 2 for border left+right + padding per side
}

// ChromeHeight returns the total vertical chrome (border) in rows.
func (bc BorderConfig) ChromeHeight() int {
	if !bc.HasBorder() {
		return 0
	}
	return 2 // border top+bottom
}

// Style returns the lipgloss style for the border in the given focus state.
func (bc BorderConfig) Style(focused bool) lipgloss.Style {
	if !bc.HasBorder() {
		return lipgloss.NewStyle()
	}
	color := bc.Color
	if focused {
		color = bc.ActiveColor
	}
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color)
	if bc.Padding > 0 {
		s = s.Padding(0, bc.Padding)
	}
	return s
}
