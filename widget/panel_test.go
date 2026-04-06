package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Compile-time interface check
var _ Container = (*Panel)(nil)

// --- Construction ---

func TestPanel_DefaultLayout_IsVertical(t *testing.T) {
	p := NewPanel("p")
	if p.layout != Vertical {
		t.Errorf("default layout = %d, want Vertical (%d)", p.layout, Vertical)
	}
}

func TestPanel_ExplicitTCBLayout(t *testing.T) {
	p := NewPanel("p", TCB)
	if p.layout != TCB {
		t.Errorf("layout = %d, want TCB (%d)", p.layout, TCB)
	}
}

func TestPanel_ExplicitHorizontalLayout(t *testing.T) {
	p := NewPanel("p", Horizontal)
	if p.layout != Horizontal {
		t.Errorf("layout = %d, want Horizontal (%d)", p.layout, Horizontal)
	}
}

func TestPanel_Defaults(t *testing.T) {
	p := NewPanel("p")
	if p.ID() != "p" {
		t.Errorf("ID = %q, want %q", p.ID(), "p")
	}
	if !p.Visible() {
		t.Error("should be visible by default")
	}
	if !p.Enabled() {
		t.Error("should be enabled by default")
	}
}

func TestPanel_Interface(t *testing.T) {
	var _ Container = (*Panel)(nil)
}

// --- Vertical layout ---

func TestPanel_Vertical_AddNext_StacksChildren(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 20)
	p.SetPosition(0, 0)

	btn1 := NewButton("b1", "A", DefaultButtonStyles())
	btn2 := NewButton("b2", "B", DefaultButtonStyles())
	p.Add(btn1, Next)
	p.Add(btn2, Next)

	p.View()

	_, y1 := btn1.Position()
	_, y2 := btn2.Position()
	if y2 <= y1 {
		t.Errorf("btn2 y=%d should be below btn1 y=%d", y2, y1)
	}
}

func TestPanel_Vertical_ChildrenGetContentWidth(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 20)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.View()

	w, _ := btn.Size()
	// No border, so content width == panel width
	if w != 40 {
		t.Errorf("child width = %d, want 40 (panel width, no border)", w)
	}
}

func TestPanel_Vertical_PreferredSizesHonored(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 20)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	btn.SetPreferredWidth(20)
	btn.SetPreferredHeight(3)
	p.Add(btn, Next)

	p.View()

	w, h := btn.Size()
	if w != 20 {
		t.Errorf("child width = %d, want 20 (preferred)", w)
	}
	if h != 3 {
		t.Errorf("child height = %d, want 3 (preferred)", h)
	}
}

func TestPanel_Vertical_TitleReducesContentHeight(t *testing.T) {
	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetSize(40, 20)
	p.SetTitle("Title")

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.View()

	// With border: chromeH = 2, contentH = 18
	// With title: contentH = 17
	// Title takes 1 line of the content area
	_, y := btn.Position()
	// btn should be positioned after title (border top + title)
	if y < 2 {
		t.Errorf("child y=%d should be offset past border top + title", y)
	}
}

func TestPanel_Vertical_WithBorder_ContentWidthReduced(t *testing.T) {
	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetSize(40, 20)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.View()

	w, _ := btn.Size()
	// DefaultBorder: chromeW = 2 + 1*2 = 4, so content width = 36
	if w != 36 {
		t.Errorf("child width = %d, want 36 (panel 40 - chrome 4)", w)
	}
}

// --- Horizontal layout ---

func TestPanel_Horizontal_AddNext_PlacesLeftToRight(t *testing.T) {
	p := NewPanel("p", Horizontal)
	p.SetSize(60, 10)
	p.SetPosition(0, 0)

	btn1 := NewButton("b1", "A", DefaultButtonStyles())
	btn2 := NewButton("b2", "B", DefaultButtonStyles())
	p.Add(btn1, Next)
	p.Add(btn2, Next)

	p.View()

	x1, _ := btn1.Position()
	x2, _ := btn2.Position()
	if x2 <= x1 {
		t.Errorf("btn2 x=%d should be right of btn1 x=%d", x2, x1)
	}
}

func TestPanel_Horizontal_Spacing(t *testing.T) {
	p := NewPanel("p", Horizontal)
	p.SetSize(60, 10)
	p.SetPosition(0, 0)
	p.SetSpacing(3)

	btn1 := NewButton("b1", "A", DefaultButtonStyles())
	btn2 := NewButton("b2", "B", DefaultButtonStyles())
	p.Add(btn1, Next)
	p.Add(btn2, Next)

	p.View()

	x1, _ := btn1.Position()
	w1, _ := btn1.Size()
	x2, _ := btn2.Position()
	gap := x2 - (x1 + w1)
	if gap != 3 {
		t.Errorf("gap between children = %d, want 3", gap)
	}
}

func TestPanel_Horizontal_ChildrenGetContentHeight(t *testing.T) {
	p := NewPanel("p", Horizontal)
	p.SetSize(60, 10)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.View()

	_, h := btn.Size()
	// No border, so content height == panel height
	if h != 10 {
		t.Errorf("child height = %d, want 10 (panel height, no border)", h)
	}
}

// --- TCB layout ---

func TestPanel_TCB_AddNext_FillsTopCenterBottom(t *testing.T) {
	p := NewPanel("p", TCB)
	p.SetSize(40, 30)
	p.SetPosition(0, 0)

	topBtn := NewButton("top", "Top", DefaultButtonStyles())
	centerBtn := NewButton("center", "Center", DefaultButtonStyles())
	bottomBtn := NewButton("bottom", "Bottom", DefaultButtonStyles())

	p.Add(topBtn, Next)
	p.Add(centerBtn, Next)
	p.Add(bottomBtn, Next)

	if p.top != topBtn {
		t.Error("first Next should fill top slot")
	}
	if p.center != centerBtn {
		t.Error("second Next should fill center slot")
	}
	if p.bottom != bottomBtn {
		t.Error("third Next should fill bottom slot")
	}

	p.View()

	_, yTop := topBtn.Position()
	_, yCenter := centerBtn.Position()
	_, yBottom := bottomBtn.Position()

	if yCenter <= yTop {
		t.Errorf("center y=%d should be below top y=%d", yCenter, yTop)
	}
	if yBottom <= yCenter {
		t.Errorf("bottom y=%d should be below center y=%d", yBottom, yCenter)
	}
}

func TestPanel_TCB_ExplicitPosition(t *testing.T) {
	p := NewPanel("p", TCB)
	p.SetSize(40, 30)

	centerBtn := NewButton("center", "Center", DefaultButtonStyles())
	p.Add(centerBtn, TCBCenter)

	if p.center != centerBtn {
		t.Error("explicit TCBCenter should fill center slot")
	}
	if p.top != nil {
		t.Error("top should be nil")
	}
	if p.bottom != nil {
		t.Error("bottom should be nil")
	}
}

func TestPanel_TCB_CenterGetsRemainingHeight(t *testing.T) {
	p := NewPanel("p", TCB)
	p.SetSize(40, 30)
	p.SetPosition(0, 0)

	topBtn := NewButton("top", "Top", DefaultButtonStyles())
	topBtn.SetPreferredHeight(3)
	centerBtn := NewButton("center", "Center", DefaultButtonStyles())
	bottomBtn := NewButton("bottom", "Bottom", DefaultButtonStyles())
	bottomBtn.SetPreferredHeight(2)

	p.Add(topBtn, Next)
	p.Add(centerBtn, Next)
	p.Add(bottomBtn, Next)

	p.View()

	_, topH := topBtn.Size()
	_, centerH := centerBtn.Size()
	_, bottomH := bottomBtn.Size()

	expected := 30 - topH - bottomH
	if centerH != expected {
		t.Errorf("center height = %d, want %d (total 30 - top %d - bottom %d)", centerH, expected, topH, bottomH)
	}
}

func TestPanel_TCB_NilSlotsUseNoSpace(t *testing.T) {
	p := NewPanel("p", TCB)
	p.SetSize(40, 30)
	p.SetPosition(0, 0)

	centerBtn := NewButton("center", "Center", DefaultButtonStyles())
	p.Add(centerBtn, TCBCenter)

	p.View()

	_, centerH := centerBtn.Size()
	// With no top/bottom, center gets full height
	if centerH != 30 {
		t.Errorf("center height = %d, want 30 (full height, no top/bottom)", centerH)
	}
}

// --- Border ---

func TestPanel_NoBorder_NoChrome(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 20)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.View()

	w, _ := btn.Size()
	if w != 40 {
		t.Errorf("child width = %d, want 40 (no border chrome)", w)
	}
}

func TestPanel_SetBorder_AddsChrome(t *testing.T) {
	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetSize(40, 20)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	output := stripansi.Strip(p.View())
	if !strings.Contains(output, "OK") {
		t.Errorf("expected button in output, got:\n%s", output)
	}
	// Rounded border chars
	if !strings.Contains(output, "\u256d") || !strings.Contains(output, "\u256f") {
		t.Errorf("expected rounded border chars, got:\n%s", output)
	}
}

func TestPanel_Border_ActiveInactiveColors(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetSize(40, 10)

	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	normalOutput := p.View()

	btn.SetFocused(true)
	activeOutput := p.View()

	if normalOutput == activeOutput {
		t.Error("border style should change when a child is focused")
	}
}

// --- Invalid position panics ---

func TestPanel_Vertical_InvalidPosition_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for TCBTop on Vertical layout")
		}
	}()

	p := NewPanel("p")
	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, TCBTop)
}

func TestPanel_Horizontal_InvalidPosition_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for TCBCenter on Horizontal layout")
		}
	}()

	p := NewPanel("p", Horizontal)
	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, TCBCenter)
}

// --- Visibility ---

func TestPanel_InvisibleChildrenSkipped(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 20)

	btn1 := NewButton("b1", "Visible", DefaultButtonStyles())
	btn2 := NewButton("b2", "Hidden", DefaultButtonStyles())
	btn2.SetVisible(false)
	p.Add(btn1, Next)
	p.Add(btn2, Next)

	output := stripansi.Strip(p.View())
	if !strings.Contains(output, "Visible") {
		t.Errorf("visible child should appear, got:\n%s", output)
	}
	if strings.Contains(output, "Hidden") {
		t.Errorf("invisible child should be skipped, got:\n%s", output)
	}
}

func TestPanel_InvisiblePanel_ReturnsEmpty(t *testing.T) {
	p := NewPanel("p")
	p.SetVisible(false)
	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	output := p.View()
	if output != "" {
		t.Errorf("invisible panel should return empty string, got: %q", output)
	}
}

// --- Focus / Active propagation ---

func TestPanel_AddChild_SetsParent(t *testing.T) {
	p := NewPanel("p")
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	if btn.Parent() != p {
		t.Error("child parent should be the panel")
	}
}

func TestPanel_InnerFocused_NoFocus(t *testing.T) {
	p := NewPanel("p")
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	if p.InnerFocused() {
		t.Error("InnerFocused should be false when no child is focused")
	}
}

func TestPanel_InnerFocused_ChildFocused(t *testing.T) {
	p := NewPanel("p")
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	p.Add(btn, Next)
	btn.SetFocused(true)

	if !p.InnerFocused() {
		t.Error("InnerFocused should be true when a child is focused")
	}
}

func TestPanel_Active_Propagation(t *testing.T) {
	p := NewPanel("p")
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	if !btn.Active() {
		t.Error("child should be active when panel is enabled")
	}

	p.SetEnabled(false)
	if btn.Active() {
		t.Error("child should be inactive when panel is disabled")
	}
}

func TestPanel_Inactive_FaintRendering(t *testing.T) {
	p := NewPanel("p")
	p.SetSize(40, 10)
	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	p.SetEnabled(false)
	output := p.View()
	// Should still render (faint), not empty
	if output == "" {
		t.Error("inactive panel should still render content (faint)")
	}
}

// --- Title ---

func TestPanel_View_Title(t *testing.T) {
	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetTitle("My Panel")
	btn := NewButton("btn", "OK", DefaultButtonStyles())
	p.Add(btn, Next)

	output := stripansi.Strip(p.View())
	if !strings.Contains(output, "My Panel") {
		t.Errorf("expected title in output, got:\n%s", output)
	}
}

// --- TCB visibility ---

func TestPanel_TCB_InvisibleSlotSkipped(t *testing.T) {
	p := NewPanel("p", TCB)
	p.SetSize(40, 30)

	topBtn := NewButton("top", "Top", DefaultButtonStyles())
	topBtn.SetVisible(false)
	centerBtn := NewButton("center", "Center", DefaultButtonStyles())

	p.Add(topBtn, Next)
	p.Add(centerBtn, Next)

	p.View()

	// Center should get full height since top is invisible
	_, centerH := centerBtn.Size()
	if centerH != 30 {
		t.Errorf("center height = %d, want 30 (invisible top uses no space)", centerH)
	}
}

func TestBorder_Style_Unfocused(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	border := SingleLineBorder("#ff0000", "#00ff00")
	unfocusedStyle := border.Style(false)
	focusedStyle := border.Style(true)

	// Render something with each style to verify they differ
	unfocusedOutput := unfocusedStyle.Render("test")
	focusedOutput := focusedStyle.Render("test")

	if unfocusedOutput == focusedOutput {
		t.Error("unfocused and focused border styles should produce different output")
	}
}

func TestBorder_Style_NoBorder(t *testing.T) {
	border := BorderConfig{Type: NoBorder}
	style := border.Style(false)
	// Should return a plain style with no border
	result := style.Render("test")
	if !strings.Contains(result, "test") {
		t.Errorf("no-border style should render content, got %q", result)
	}
}

func TestPanel_SetTitleStyle(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	p := NewPanel("p")
	p.SetBorder(DefaultBorder())
	p.SetSize(40, 10)
	p.SetTitle("Test Title")

	view1 := p.View()

	p.SetTitleStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true))
	view2 := p.View()

	if view1 == view2 {
		t.Error("different title style should produce different output")
	}
}

func TestPanel_TCB_InvalidPosition_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid TCB position")
		}
	}()

	p := NewPanel("p", TCB)
	btn := NewButton("b", "OK", DefaultButtonStyles())
	p.Add(btn, Position(99))
}

func TestPanel_SingleChild_AllLayouts_SameResult(t *testing.T) {
	// A bordered panel with a single child should render the child
	// and fill the panel's allocated height regardless of layout.
	layouts := []struct {
		name      string
		layout    Layout
		pos       Position
		wantWidth int // expected child width (Horizontal uses natural, others use content width)
	}{
		{"Vertical", Vertical, Next, 56},    // content width: 60 - 4 chrome
		{"Horizontal", Horizontal, Next, 5}, // natural width of "Hello"
		{"TCB", TCB, TCBCenter, 56},         // content width: 60 - 4 chrome
	}

	for _, tc := range layouts {
		t.Run(tc.name, func(t *testing.T) {
			panel := NewPanel("p", tc.layout)
			panel.SetBorder(DefaultBorder())
			panel.SetSize(60, 20)

			txt := NewText("t", "Hello", lipgloss.NewStyle())
			panel.Add(txt, tc.pos)

			view := panel.View()
			plain := stripansi.Strip(view)

			if !strings.Contains(plain, "Hello") {
				t.Errorf("should contain 'Hello', got:\n%s", plain)
			}

			w, _ := txt.Size()
			if w != tc.wantWidth {
				t.Errorf("child width = %d, want %d", w, tc.wantWidth)
			}

			// Panel should render at full allocated height
			h := lipgloss.Height(view)
			if h != 20 {
				t.Errorf("panel height = %d, want 20", h)
			}
		})
	}
}
