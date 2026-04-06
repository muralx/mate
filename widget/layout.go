package widget

import "github.com/charmbracelet/lipgloss"

// Layout represents the layout strategy for a container.
type Layout int

const (
	Vertical   Layout = iota // stack children top-to-bottom
	Horizontal               // stack children left-to-right
	TCB                      // Top-Center-Bottom: center flexes
)

// Position specifies where a child is placed in the layout.
type Position int

const (
	Next      Position = iota // sequential: append for V/H, fill next slot for TCB
	TCBTop                    // TCB only: top slot
	TCBCenter                 // TCB only: center slot
	TCBBottom                 // TCB only: bottom slot
)

// LayoutVertical arranges children top-to-bottom within available space.
// Each child gets availW (unless it has a preferred width).
// Each child gets its preferred height, or its natural rendered height if no preference.
// No flexing — remaining space below the last child is unused.
func LayoutVertical(children []Component, x, y, availW, availH int) {
	yOffset := y
	for _, child := range children {
		cw := child.PreferredWidth()
		if cw == 0 {
			cw = availW
		}
		ch := child.PreferredHeight()
		if ch == 0 {
			// Measure natural height
			child.SetSize(cw, 0)
			rendered := child.View()
			ch = lipgloss.Height(rendered)
		}
		child.SetSize(cw, ch)
		child.SetPosition(x, yOffset)
		yOffset += ch
	}
}

// LayoutHorizontal arranges children left-to-right within available space.
// Each child gets availH (unless it has a preferred height).
// Each child gets its preferred width, or its natural rendered width if no preference.
// No flexing — remaining space to the right is unused.
func LayoutHorizontal(children []Component, x, y, availW, availH, spacing int) {
	xOffset := x
	for i, child := range children {
		ch := child.PreferredHeight()
		if ch == 0 {
			ch = availH
		}
		cw := child.PreferredWidth()
		if cw == 0 {
			// Measure natural width
			child.SetSize(0, ch)
			rendered := child.View()
			cw = lipgloss.Width(rendered)
		}
		child.SetSize(cw, ch)
		child.SetPosition(xOffset, y)
		xOffset += cw
		if i < len(children)-1 {
			xOffset += spacing
		}
	}
}

// LayoutTCB arranges three slots: top, center, bottom.
// Top and Bottom get their preferred/natural height.
// Center gets ALL remaining height — this is the only layout that stretches.
// All slots get availW (unless they have a preferred width).
// Nil slots use no space.
func LayoutTCB(top, center, bottom Component, x, y, availW, availH int) {
	topH := 0
	bottomH := 0

	if top != nil {
		tw := top.PreferredWidth()
		if tw == 0 {
			tw = availW
		}
		topH = top.PreferredHeight()
		if topH == 0 {
			top.SetSize(tw, 0)
			rendered := top.View()
			topH = lipgloss.Height(rendered)
		}
		top.SetSize(tw, topH)
		top.SetPosition(x, y)
	}

	if bottom != nil {
		bw := bottom.PreferredWidth()
		if bw == 0 {
			bw = availW
		}
		bottomH = bottom.PreferredHeight()
		if bottomH == 0 {
			bottom.SetSize(bw, 0)
			rendered := bottom.View()
			bottomH = lipgloss.Height(rendered)
		}
		bottom.SetSize(bw, bottomH)
	}

	centerH := availH - topH - bottomH
	if centerH < 0 {
		centerH = 0
	}

	if center != nil {
		cw := center.PreferredWidth()
		if cw == 0 {
			cw = availW
		}
		center.SetSize(cw, centerH)
		center.SetPosition(x, y+topH)
	}

	if bottom != nil {
		bottom.SetPosition(x, y+topH+centerH)
	}
}
