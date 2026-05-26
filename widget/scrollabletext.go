package widget

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ScrollableTextStyles defines the styles used by ScrollableText.
type ScrollableTextStyles struct {
	Normal  lipgloss.Style // unfocused appearance
	Focused lipgloss.Style // focused appearance
}

// DefaultScrollableTextStyles returns sensible defaults for ScrollableText.
func DefaultScrollableTextStyles() ScrollableTextStyles {
	return ScrollableTextStyles{
		Normal:  lipgloss.NewStyle(),
		Focused: lipgloss.NewStyle(),
	}
}

// ScrollableText is a focusable, scrollable, read-only text display area.
// It renders text within its SetSize bounds and scrolls with keyboard input
// when focused. Content may contain ANSI color codes.
type ScrollableText struct {
	FocusableComponent
	content string
	lines   []string // cached split lines
	offset  int      // scroll offset (first visible line)
	wrap    bool     // whether to wrap long lines
	styles  ScrollableTextStyles
}

// NewScrollableText creates a new ScrollableText with the given ID and styles.
func NewScrollableText(id string, styles ScrollableTextStyles) *ScrollableText {
	st := &ScrollableText{
		styles: styles,
		wrap:   true, // wrap by default
	}
	st.FocusableComponent = NewFocusableComponent(id)
	return st
}

// SetContent sets the full text content (may contain newlines and ANSI styling).
func (st *ScrollableText) SetContent(text string) {
	st.content = text
	st.lines = nil // invalidate cache
	st.offset = 0
}

// Content returns the current content.
func (st *ScrollableText) Content() string { return st.content }

// SetWrap sets whether long lines are wrapped to fit the viewport width.
// When false, lines are truncated. Default is true.
func (st *ScrollableText) SetWrap(wrap bool) {
	st.wrap = wrap
	st.lines = nil // invalidate cache
}

// ScrollTo scrolls to a specific line (clamped to valid range).
func (st *ScrollableText) ScrollTo(line int) {
	total := len(st.getLines())
	vp := st.viewportHeight()
	maxOffset := total - vp
	if maxOffset < 0 {
		maxOffset = 0
	}
	if line < 0 {
		line = 0
	}
	if line > maxOffset {
		line = maxOffset
	}
	st.offset = line
}

// ScrollTop scrolls to the top.
func (st *ScrollableText) ScrollTop() {
	st.offset = 0
}

// viewportHeight returns how many lines fit in the viewport.
func (st *ScrollableText) viewportHeight() int {
	if st.height < 1 {
		return 1
	}
	return st.height
}

// getLines returns the content split into display lines, applying wrapping
// or truncation based on the wrap setting.
func (st *ScrollableText) getLines() []string {
	if st.lines != nil {
		return st.lines
	}

	if st.content == "" {
		st.lines = []string{}
		return st.lines
	}

	rawLines := strings.Split(st.content, "\n")

	if !st.wrap || st.width <= 0 {
		st.lines = rawLines
		return st.lines
	}

	// Wrap long lines. ansi.Wrap is ANSI-safe and word-aware on the
	// supplied breakpoints (spaces, tabs, dashes); it emits a string
	// with newlines inserted at width boundaries that we split back
	// into display lines.
	//
	// (The previous Truncate-then-byte-advance loop split mid-CSI when
	// Truncate appended a reset that wasn't present in the input,
	// producing visible escape-sequence fragments like ";2;206;145;120m"
	// in styled markdown.)
	var wrapped []string
	for _, line := range rawLines {
		if lipgloss.Width(line) <= st.width {
			wrapped = append(wrapped, line)
			continue
		}
		w := ansi.Wrap(line, st.width, " \t-")
		wrapped = append(wrapped, strings.Split(w, "\n")...)
	}
	st.lines = wrapped
	return st.lines
}

// scrollBy shifts the viewport offset by delta lines (negative=up, positive=down),
// clamped to valid range.
func (st *ScrollableText) scrollBy(delta int) {
	total := len(st.getLines())
	vp := st.viewportHeight()
	maxOffset := total - vp
	if maxOffset < 0 {
		maxOffset = 0
	}
	st.offset += delta
	if st.offset < 0 {
		st.offset = 0
	}
	if st.offset > maxOffset {
		st.offset = maxOffset
	}
}

// HandleEvent handles mouse scroll events.
func (st *ScrollableText) HandleEvent(event Event) (tea.Cmd, bool) {
	if scroll, ok := event.(MouseScrollEvent); ok {
		if !st.Active() {
			return nil, false
		}
		st.scrollBy(scroll.Direction * 3)
		return nil, true
	}
	return st.BaseComponent.HandleEvent(event)
}

// Update handles keyboard scrolling when focused.
func (st *ScrollableText) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !st.Active() {
		return nil, false
	}

	vp := st.viewportHeight()
	switch msg.String() {
	case "up", "k":
		st.scrollBy(-1)
		return nil, true
	case "down", "j":
		st.scrollBy(1)
		return nil, true
	case "pgup":
		st.scrollBy(-vp)
		return nil, true
	case "pgdown":
		st.scrollBy(vp)
		return nil, true
	case "home":
		st.offset = 0
		return nil, true
	case "end":
		st.scrollBy(len(st.getLines()))
		return nil, true
	}

	if st.onKeyPress != nil {
		if cmd := st.onKeyPress(msg); cmd != nil {
			return cmd, true
		}
	}
	return nil, false
}

// View renders the visible portion of the text content.
func (st *ScrollableText) View() string {
	lines := st.getLines()
	vp := st.viewportHeight()

	style := st.styles.Normal
	if st.Focused() {
		style = st.styles.Focused
	}

	if len(lines) == 0 {
		return style.Width(st.width).Height(st.height).Render("")
	}

	// Clamp offset
	maxOffset := len(lines) - vp
	if maxOffset < 0 {
		maxOffset = 0
	}
	if st.offset > maxOffset {
		st.offset = maxOffset
	}

	end := st.offset + vp
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[st.offset:end]

	// Truncate lines if not wrapping
	if !st.wrap && st.width > 0 {
		truncated := make([]string, len(visible))
		for i, line := range visible {
			truncated[i] = ansi.Truncate(line, st.width, "")
		}
		visible = truncated
	}

	content := strings.Join(visible, "\n")

	if !st.Active() {
		return lipgloss.NewStyle().Faint(true).Width(st.width).Height(st.height).Render(content)
	}

	return style.Width(st.width).Height(st.height).Render(content)
}
