package widget

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Panel is the universal container. It has a layout (Vertical, Horizontal, or TCB),
// an optional border, an optional title, and a unified Add(child, Position) method.
type Panel struct {
	BaseContainer
	layout     Layout
	border     BorderConfig
	title      string
	titleStyle lipgloss.Style
	spacing    int // horizontal padding between children (Horizontal layout only)

	// TCB slots (only used when layout == TCB)
	top         Component
	center      Component
	bottom      Component
	tcbNextSlot Position
}

// NewPanel creates a Panel. Layout defaults to Vertical.
// Optional layout parameter: NewPanel("id") or NewPanel("id", TCB)
func NewPanel(id string, layout ...Layout) *Panel {
	l := Vertical
	if len(layout) > 0 {
		l = layout[0]
	}
	p := &Panel{
		layout:      l,
		titleStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0")).Bold(true),
		tcbNextSlot: TCBTop,
	}
	p.BaseContainer = *NewBaseContainer(id, p)
	return p
}

// SetBorder sets the border configuration.
func (p *Panel) SetBorder(config BorderConfig) { p.border = config }

// SetTitle sets the panel's title text (rendered inside the border, above children).
func (p *Panel) SetTitle(title string) { p.title = title }

// SetTitleStyle sets the style for the title text.
func (p *Panel) SetTitleStyle(style lipgloss.Style) { p.titleStyle = style }

// SetSpacing sets spacing between children (Horizontal layout only).
func (p *Panel) SetSpacing(spacing int) { p.spacing = spacing }

// Add places a child component at the given position.
// For Vertical/Horizontal: use Next to append sequentially.
// For TCB: use Next (fills Top->Center->Bottom) or explicit TCBTop/TCBCenter/TCBBottom.
func (p *Panel) Add(child Component, position Position) {
	switch p.layout {
	case Vertical, Horizontal:
		if position != Next {
			panic(fmt.Sprintf("position %d is not valid for layout %d; use Next", position, p.layout))
		}
		p.AddChild(child)

	case TCB:
		slot := position
		if slot == Next {
			slot = p.tcbNextSlot
		}
		switch slot {
		case TCBTop:
			p.top = child
			if position == Next {
				p.tcbNextSlot = TCBCenter
			}
		case TCBCenter:
			p.center = child
			if position == Next {
				p.tcbNextSlot = TCBBottom
			}
		case TCBBottom:
			p.bottom = child
		default:
			panic(fmt.Sprintf("invalid TCB position: %d", slot))
		}
		p.AddChild(child) // still add for focus management
	}
}

// View renders the panel with its layout and optional border.
func (p *Panel) View() string {
	if !p.Visible() {
		return ""
	}

	w, h := p.Size()
	px, py := p.Position()

	// Chrome from border
	chromeW := p.border.ChromeWidth()
	chromeH := p.border.ChromeHeight()

	contentW := w - chromeW
	if contentW < 0 {
		contentW = 0
	}
	contentH := h - chromeH
	if contentH < 0 {
		contentH = 0
	}

	contentX := px
	contentY := py
	if p.border.HasBorder() {
		contentX += 1 + p.border.Padding // border char + padding
		contentY += 1                    // border top
	}

	// Title
	var parts []string
	titleH := 0
	if p.title != "" {
		parts = append(parts, p.titleStyle.Render(p.title))
		titleH = 1
		contentY++
		contentH -= titleH
		if contentH < 0 {
			contentH = 0
		}
	}

	// Layout children
	switch p.layout {
	case Vertical:
		visible := p.visibleSequentialChildren()
		LayoutVertical(visible, contentX, contentY, contentW, contentH)
		for _, child := range visible {
			parts = append(parts, child.View())
		}

	case Horizontal:
		visible := p.visibleSequentialChildren()
		LayoutHorizontal(visible, contentX, contentY, contentW, contentH, p.spacing)
		var childParts []string
		for i, child := range visible {
			if i > 0 && p.spacing > 0 {
				childParts = append(childParts, strings.Repeat(" ", p.spacing))
			}
			childParts = append(childParts, child.View())
		}
		if len(childParts) > 0 {
			parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Top, childParts...))
		}

	case TCB:
		var top, center, bottom Component
		if p.top != nil && p.top.Visible() {
			top = p.top
		}
		if p.center != nil && p.center.Visible() {
			center = p.center
		}
		if p.bottom != nil && p.bottom.Visible() {
			bottom = p.bottom
		}
		LayoutTCB(top, center, bottom, contentX, contentY, contentW, contentH)
		if top != nil {
			parts = append(parts, top.View())
		}
		if center != nil {
			parts = append(parts, center.View())
		}
		if bottom != nil {
			parts = append(parts, bottom.View())
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Apply border if set
	if p.border.HasBorder() {
		borderStyle := p.border.Style(p.InnerFocused())
		if w > 0 {
			borderStyle = borderStyle.Width(w - 2) // border chars only
		}
		if h > 0 {
			borderStyle = borderStyle.Height(h - 2)
		}
		if !p.Active() {
			return lipgloss.NewStyle().Faint(true).Render(borderStyle.Render(content))
		}
		return borderStyle.Render(content)
	}

	// No border — still enforce allocated size
	s := lipgloss.NewStyle()
	if w > 0 {
		s = s.Width(w)
	}
	if h > 0 {
		s = s.Height(h)
	}
	if !p.Active() {
		return lipgloss.NewStyle().Faint(true).Render(s.Render(content))
	}
	return s.Render(content)
}

// visibleSequentialChildren returns visible children for Vertical/Horizontal layouts.
func (p *Panel) visibleSequentialChildren() []Component {
	var visible []Component
	for _, child := range p.Children() {
		if child.Visible() {
			visible = append(visible, child)
		}
	}
	return visible
}
