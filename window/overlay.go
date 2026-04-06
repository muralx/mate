package window

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// OverlayOffset stores the screen offset where the popup content is rendered.
// Used by the Stack to adjust mouse coordinates for popup hit testing.
type OverlayOffset struct {
	X, Y int // content start position (inside the border)
}

// RenderOverlay renders a popup centered on a dimmed background.
// Returns the rendered string and the content offset for mouse coordinate adjustment.
func RenderOverlay(content, title string, width, height int) (string, OverlayOffset) {
	// Create a dimmed empty background
	dimStyle := lipgloss.NewStyle().Faint(true)
	baseLines := make([]string, height)
	emptyLine := dimStyle.Render(strings.Repeat(" ", width))
	for i := range baseLines {
		baseLines[i] = emptyLine
	}

	// Build popup box
	popupLines := strings.Split(content, "\n")
	popupWidth := 0
	for _, l := range popupLines {
		if w := lipgloss.Width(l); w > popupWidth {
			popupWidth = w
		}
	}
	popupWidth += 4 // padding

	if popupWidth > width-4 {
		popupWidth = width - 4
	}
	popupHeight := len(popupLines) + 2 // border
	if title != "" {
		popupHeight++ // title line inside border via SetString
	}

	// Build bordered popup
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(0, 1).
		Width(popupWidth)

	renderContent := content
	if title != "" {
		renderContent = title + "\n" + content
	}

	bordered := borderStyle.Render(renderContent)
	borderedLines := strings.Split(bordered, "\n")

	// Center the popup
	startY := (height - popupHeight) / 2
	if startY < 1 {
		startY = 1
	}
	startX := (width - popupWidth - 2) / 2
	if startX < 0 {
		startX = 0
	}

	// Overlay popup lines onto dimmed base
	for i, line := range borderedLines {
		y := startY + i
		if y >= 0 && y < len(baseLines) {
			padded := strings.Repeat(" ", startX) + line
			baseLines[y] = padded
		}
	}

	// Content offset: startX + border(1) + padding(1) for X, startY + border(1) for Y
	// If title present, add 1 for the title line rendered inside the border
	offsetX := startX + 2 // border + padding
	offsetY := startY + 1 // border
	if title != "" {
		offsetY++ // title line inside the border
	}

	return strings.Join(baseLines[:height], "\n"), OverlayOffset{X: offsetX, Y: offsetY}
}
