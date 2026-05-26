package widget

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var _ Leaf = (*ScrollableText)(nil)

func makeScrollableText(content string) *ScrollableText {
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetSize(40, 5)
	st.SetContent(content)
	return st
}

func TestScrollableText_Defaults(t *testing.T) {
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	if st.ID() != "st" {
		t.Errorf("ID = %q, want %q", st.ID(), "st")
	}
	if !st.Focusable() {
		t.Error("should be focusable")
	}
	if st.Content() != "" {
		t.Errorf("content should be empty, got %q", st.Content())
	}
}

func TestScrollableText_SetContent(t *testing.T) {
	st := makeScrollableText("hello\nworld")
	if st.Content() != "hello\nworld" {
		t.Errorf("content = %q", st.Content())
	}
}

func TestScrollableText_View_FitsInViewport(t *testing.T) {
	st := makeScrollableText("line1\nline2\nline3")
	view := st.View()
	if !strings.Contains(view, "line1") || !strings.Contains(view, "line2") || !strings.Contains(view, "line3") {
		t.Errorf("all lines should be visible, got:\n%s", view)
	}
}

func TestScrollableText_View_Empty(t *testing.T) {
	st := makeScrollableText("")
	view := st.View()
	// Should render an empty area with the set dimensions
	h := lipgloss.Height(view)
	if h < 5 {
		t.Errorf("empty view height = %d, want >= 5", h)
	}
}

func TestScrollableText_View_Scrolled(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strings.Repeat("x", 5)
	}
	content := strings.Join(lines, "\n")

	st := makeScrollableText(content)
	// viewport = 5, scroll down
	st.Update(tea.KeyMsg{Type: tea.KeyDown})
	st.Update(tea.KeyMsg{Type: tea.KeyDown})

	// offset should be 2
	view := st.View()
	viewLines := strings.Split(view, "\n")
	if len(viewLines) < 5 {
		t.Errorf("expected at least 5 visible lines, got %d", len(viewLines))
	}
}

func TestScrollableText_Update_Down(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !consumed {
		t.Error("down should be consumed")
	}
}

func TestScrollableText_Update_Up(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(3)

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyUp})
	if !consumed {
		t.Error("up should be consumed")
	}
}

func TestScrollableText_Update_Up_AtTop(t *testing.T) {
	st := makeScrollableText("a\nb\nc")

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyUp})
	if !consumed {
		t.Error("up should be consumed even at top")
	}
}

func TestScrollableText_Update_PgDown(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no")

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if !consumed {
		t.Error("pgdown should be consumed")
	}
}

func TestScrollableText_Update_PgUp(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no")
	st.ScrollTo(10)

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if !consumed {
		t.Error("pgup should be consumed")
	}
}

func TestScrollableText_Update_Home(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(5)

	st.Update(tea.KeyMsg{Type: tea.KeyHome})
	view := st.View()
	if !strings.Contains(view, "a") {
		t.Error("home should scroll to top")
	}
}

func TestScrollableText_Update_End(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")

	st.Update(tea.KeyMsg{Type: tea.KeyEnd})
	view := st.View()
	if !strings.Contains(view, "j") {
		t.Error("end should scroll to bottom")
	}
}

func TestScrollableText_Update_Inactive_Ignored(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf")
	st.SetEnabled(false)

	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyDown})
	if consumed {
		t.Error("inactive scrollable text should not consume keys")
	}
}

func TestScrollableText_ScrollTo(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(3)
	view := st.View()
	if !strings.Contains(view, "d") {
		t.Error("scroll to 3 should show line 'd'")
	}
}

func TestScrollableText_ScrollTo_Clamped(t *testing.T) {
	st := makeScrollableText("a\nb\nc")
	st.ScrollTo(100) // should clamp
	// With 3 lines and viewport=5, max offset is 0
	view := st.View()
	if !strings.Contains(view, "a") {
		t.Error("clamped scroll should still show first line")
	}
}

func TestScrollableText_ScrollTop(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(5)
	st.ScrollTop()
	view := st.View()
	if !strings.Contains(view, "a") {
		t.Error("ScrollTop should show first line")
	}
}

func TestScrollableText_SetWrap_False(t *testing.T) {
	longLine := strings.Repeat("x", 100)
	st := makeScrollableText(longLine)
	st.SetWrap(false)

	view := st.View()
	// View should be truncated to width
	w := lipgloss.Width(view)
	if w > 40 {
		t.Errorf("truncated view width = %d, want <= 40", w)
	}
}

func TestScrollableText_OnKeyPress_Fallthrough(t *testing.T) {
	st := makeScrollableText("test")
	called := false
	st.OnKeyPress(func(msg tea.KeyMsg) tea.Cmd {
		if msg.String() == "enter" {
			called = true
			return tea.Quit
		}
		return nil
	})

	cmd, consumed := st.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !consumed {
		t.Error("enter should be consumed via onKeyPress")
	}
	if !called {
		t.Error("onKeyPress should have been called")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func TestScrollableText_InsidePanel(t *testing.T) {
	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())
	panel.SetSize(60, 10)
	panel.SetPosition(0, 0)

	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetContent("hello\nworld")
	panel.Add(st, Next)

	view := panel.View()
	if !strings.Contains(view, "hello") {
		t.Error("scrollable text should be visible inside panel")
	}
	if st.Parent() != panel {
		t.Error("parent should be panel")
	}
}

func TestScrollableText_PositionAndSize(t *testing.T) {
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetSize(40, 5)
	st.SetPosition(10, 20)
	st.SetContent("test")

	x, y := st.Position()
	if x != 10 || y != 20 {
		t.Errorf("position = (%d,%d), want (10,20)", x, y)
	}

	w, h := st.Size()
	if w != 40 || h != 5 {
		t.Errorf("size = (%d,%d), want (40,5)", w, h)
	}
}

func TestScrollableText_ActivePropagation(t *testing.T) {
	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetSize(40, 5)
	st.SetContent("a\nb\nc\nd\ne\nf\ng")
	panel.Add(st, Next)

	if !st.Active() {
		t.Error("should be active when parent enabled")
	}
	panel.SetEnabled(false)
	if st.Active() {
		t.Error("should be inactive when parent disabled")
	}

	// Inactive should not consume keys
	_, consumed := st.Update(tea.KeyMsg{Type: tea.KeyDown})
	if consumed {
		t.Error("inactive scrollable text should not consume keys")
	}
}

func TestScrollableText_ContentReset(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(5)

	// Setting new content should reset scroll
	st.SetContent("new content")
	view := st.View()
	if !strings.Contains(view, "new content") {
		t.Error("new content should be visible after SetContent")
	}
}

func TestScrollableText_WrapLongLines(t *testing.T) {
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetSize(10, 20) // narrow width
	st.SetContent(strings.Repeat("x", 25))

	view := st.View()
	lines := strings.Split(view, "\n")
	// 25 chars at width 10 should wrap to at least 3 lines of content
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty < 2 {
		t.Errorf("wrapped content should span multiple lines, got %d non-empty lines", nonEmpty)
	}
}

func TestScrollableText_HandleEvent_MouseScroll_Down(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	st := makeScrollableText(strings.Join(lines, "\n"))

	_, consumed := st.HandleEvent(MouseScrollEvent{Direction: 1})
	if !consumed {
		t.Error("scroll down should be consumed")
	}
	// Scrolled by 3 lines (direction * 3)
	if st.offset != 3 {
		t.Errorf("offset = %d, want 3 after scroll down", st.offset)
	}
}

func TestScrollableText_HandleEvent_MouseScroll_Up(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	st := makeScrollableText(strings.Join(lines, "\n"))
	st.ScrollTo(10)

	_, consumed := st.HandleEvent(MouseScrollEvent{Direction: -1})
	if !consumed {
		t.Error("scroll up should be consumed")
	}
	// Scrolled up by 3 lines
	if st.offset != 7 {
		t.Errorf("offset = %d, want 7 after scroll up from 10", st.offset)
	}
}

func TestScrollableText_HandleEvent_MouseScroll_Inactive(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.SetEnabled(false)

	_, consumed := st.HandleEvent(MouseScrollEvent{Direction: 1})
	if consumed {
		t.Error("scroll on inactive scrollabletext should not be consumed")
	}
}

func TestScrollableText_ScrollTo_Negative(t *testing.T) {
	st := makeScrollableText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	st.ScrollTo(5)
	st.ScrollTo(-5)
	if st.offset != 0 {
		t.Errorf("offset = %d, want 0 (negative clamped to 0)", st.offset)
	}
}

func TestScrollableText_View_Inactive(t *testing.T) {
	st := makeScrollableText("hello\nworld")
	st.SetEnabled(false)
	output := st.View()
	if !strings.Contains(output, "hello") {
		t.Errorf("inactive view should still show content, got %q", output)
	}
}

func TestScrollableText_View_Inactive_StripsANSI(t *testing.T) {
	st := NewScrollableText("st", DefaultScrollableTextStyles())
	st.SetSize(40, 3)
	// Content contains a hard-coded ANSI color sequence.
	st.SetContent("\x1b[31mred text\x1b[0m")
	st.SetEnabled(false)
	out := st.View()
	if strings.Contains(out, "\x1b[31m") {
		t.Errorf("inactive view should not contain inner ANSI color codes, got %q", out)
	}
	if !strings.Contains(out, "red text") {
		t.Errorf("inactive view should still contain the visible text, got %q", out)
	}
}

func TestScrollableText_FocusedStyle(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	styles := ScrollableTextStyles{
		Normal:  lipgloss.NewStyle(),
		Focused: lipgloss.NewStyle().Background(lipgloss.Color("#ff0000")),
	}
	st := NewScrollableText("st", styles)
	st.SetSize(20, 3)
	st.SetContent("test")

	// Unfocused — uses Normal style
	st.SetFocused(false)
	view1 := st.View()

	// Focused — uses Focused style (has background color)
	st.SetFocused(true)
	view2 := st.View()

	// They should differ when styles differ
	if view1 == view2 {
		t.Error("focused and unfocused views should differ when styles differ")
	}
}
