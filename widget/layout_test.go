package widget

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// --- LayoutVertical tests ---

func TestLayoutVertical_ChildrenGetAvailableWidth(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	b := NewText("b", "world", lipgloss.NewStyle())
	children := []Component{a, b}

	LayoutVertical(children, 0, 0, 80, 24)

	w1, _ := a.Size()
	w2, _ := b.Size()
	if w1 != 80 {
		t.Errorf("child a width = %d, want 80", w1)
	}
	if w2 != 80 {
		t.Errorf("child b width = %d, want 80", w2)
	}
}

func TestLayoutVertical_PreferredWidthHonored(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	a.SetPreferredWidth(30)
	children := []Component{a}

	LayoutVertical(children, 0, 0, 80, 24)

	w, _ := a.Size()
	if w != 30 {
		t.Errorf("child width = %d, want 30 (preferred)", w)
	}
}

func TestLayoutVertical_PreferredHeightHonored(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	a.SetPreferredHeight(5)
	children := []Component{a}

	LayoutVertical(children, 0, 0, 80, 24)

	_, h := a.Size()
	if h != 5 {
		t.Errorf("child height = %d, want 5 (preferred)", h)
	}
}

func TestLayoutVertical_NaturalHeightMeasured(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	// No preferred height — layout should measure via View()
	children := []Component{a}

	LayoutVertical(children, 0, 0, 80, 24)

	_, h := a.Size()
	if h < 1 {
		t.Errorf("child height = %d, want >= 1 (measured)", h)
	}
}

func TestLayoutVertical_PositionsSequential(t *testing.T) {
	a := NewButton("a", "A", DefaultButtonStyles())
	a.SetPreferredHeight(3)
	b := NewButton("b", "B", DefaultButtonStyles())
	b.SetPreferredHeight(5)
	c := NewButton("c", "C", DefaultButtonStyles())
	c.SetPreferredHeight(2)
	children := []Component{a, b, c}

	LayoutVertical(children, 10, 20, 80, 24)

	ax, ay := a.Position()
	bx, by := b.Position()
	cx, cy := c.Position()

	if ax != 10 || ay != 20 {
		t.Errorf("a position = (%d,%d), want (10,20)", ax, ay)
	}
	if bx != 10 || by != 23 {
		t.Errorf("b position = (%d,%d), want (10,23)", bx, by)
	}
	if cx != 10 || cy != 28 {
		t.Errorf("c position = (%d,%d), want (10,28)", cx, cy)
	}
}

// --- LayoutHorizontal tests ---

func TestLayoutHorizontal_ChildrenGetAvailableHeight(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	b := NewText("b", "world", lipgloss.NewStyle())
	children := []Component{a, b}

	LayoutHorizontal(children, 0, 0, 80, 24, 0)

	_, h1 := a.Size()
	_, h2 := b.Size()
	if h1 != 24 {
		t.Errorf("child a height = %d, want 24", h1)
	}
	if h2 != 24 {
		t.Errorf("child b height = %d, want 24", h2)
	}
}

func TestLayoutHorizontal_PreferredWidthHonored(t *testing.T) {
	a := NewText("a", "hello", lipgloss.NewStyle())
	a.SetPreferredWidth(20)
	children := []Component{a}

	LayoutHorizontal(children, 0, 0, 80, 24, 0)

	w, _ := a.Size()
	if w != 20 {
		t.Errorf("child width = %d, want 20 (preferred)", w)
	}
}

func TestLayoutHorizontal_NaturalWidthMeasured(t *testing.T) {
	a := NewText("a", "hi", lipgloss.NewStyle())
	// No preferred width — layout should measure via View()
	children := []Component{a}

	LayoutHorizontal(children, 0, 0, 80, 24, 0)

	w, _ := a.Size()
	if w < 1 {
		t.Errorf("child width = %d, want >= 1 (measured)", w)
	}
}

func TestLayoutHorizontal_Spacing(t *testing.T) {
	a := NewButton("a", "A", DefaultButtonStyles())
	a.SetPreferredWidth(10)
	b := NewButton("b", "B", DefaultButtonStyles())
	b.SetPreferredWidth(10)
	c := NewButton("c", "C", DefaultButtonStyles())
	c.SetPreferredWidth(10)
	children := []Component{a, b, c}

	LayoutHorizontal(children, 5, 0, 80, 24, 3)

	ax, _ := a.Position()
	bx, _ := b.Position()
	cx, _ := c.Position()

	if ax != 5 {
		t.Errorf("a x = %d, want 5", ax)
	}
	if bx != 18 { // 5 + 10 + 3
		t.Errorf("b x = %d, want 18", bx)
	}
	if cx != 31 { // 18 + 10 + 3
		t.Errorf("c x = %d, want 31", cx)
	}
}

// --- LayoutTCB tests ---

func TestLayoutTCB_CenterGetsRemaining(t *testing.T) {
	top := NewButton("top", "T", DefaultButtonStyles())
	top.SetPreferredHeight(3)
	center := NewButton("center", "C", DefaultButtonStyles())
	bottom := NewButton("bottom", "B", DefaultButtonStyles())
	bottom.SetPreferredHeight(5)

	LayoutTCB(top, center, bottom, 0, 0, 80, 24)

	_, ch := center.Size()
	want := 24 - 3 - 5
	if ch != want {
		t.Errorf("center height = %d, want %d", ch, want)
	}
}

func TestLayoutTCB_NilTop(t *testing.T) {
	center := NewButton("center", "C", DefaultButtonStyles())
	bottom := NewButton("bottom", "B", DefaultButtonStyles())
	bottom.SetPreferredHeight(4)

	LayoutTCB(nil, center, bottom, 0, 0, 80, 24)

	cx, cy := center.Position()
	_, ch := center.Size()
	if cy != 0 {
		t.Errorf("center y = %d, want 0 (no top)", cy)
	}
	if cx != 0 {
		t.Errorf("center x = %d, want 0", cx)
	}
	if ch != 20 { // 24 - 4
		t.Errorf("center height = %d, want 20", ch)
	}
}

func TestLayoutTCB_NilBottom(t *testing.T) {
	top := NewButton("top", "T", DefaultButtonStyles())
	top.SetPreferredHeight(6)
	center := NewButton("center", "C", DefaultButtonStyles())

	LayoutTCB(top, center, nil, 0, 0, 80, 24)

	_, ch := center.Size()
	if ch != 18 { // 24 - 6
		t.Errorf("center height = %d, want 18", ch)
	}
}

func TestLayoutTCB_NilTopAndBottom(t *testing.T) {
	center := NewButton("center", "C", DefaultButtonStyles())

	LayoutTCB(nil, center, nil, 0, 0, 80, 24)

	_, ch := center.Size()
	if ch != 24 {
		t.Errorf("center height = %d, want 24 (full avail)", ch)
	}
}

func TestLayoutTCB_AllGetAvailableWidth(t *testing.T) {
	top := NewButton("top", "T", DefaultButtonStyles())
	top.SetPreferredHeight(3)
	center := NewButton("center", "C", DefaultButtonStyles())
	bottom := NewButton("bottom", "B", DefaultButtonStyles())
	bottom.SetPreferredHeight(3)

	LayoutTCB(top, center, bottom, 0, 0, 60, 24)

	tw, _ := top.Size()
	cw, _ := center.Size()
	bw, _ := bottom.Size()
	if tw != 60 {
		t.Errorf("top width = %d, want 60", tw)
	}
	if cw != 60 {
		t.Errorf("center width = %d, want 60", cw)
	}
	if bw != 60 {
		t.Errorf("bottom width = %d, want 60", bw)
	}
}

func TestLayoutTCB_Positions(t *testing.T) {
	top := NewButton("top", "T", DefaultButtonStyles())
	top.SetPreferredHeight(4)
	center := NewButton("center", "C", DefaultButtonStyles())
	bottom := NewButton("bottom", "B", DefaultButtonStyles())
	bottom.SetPreferredHeight(6)

	LayoutTCB(top, center, bottom, 5, 10, 80, 30)

	tx, ty := top.Position()
	cx, cy := center.Position()
	bx, by := bottom.Position()

	if tx != 5 || ty != 10 {
		t.Errorf("top position = (%d,%d), want (5,10)", tx, ty)
	}
	if cx != 5 || cy != 14 { // 10 + 4
		t.Errorf("center position = (%d,%d), want (5,14)", cx, cy)
	}
	centerH := 30 - 4 - 6      // 20
	wantBy := 10 + 4 + centerH // 34
	if bx != 5 || by != wantBy {
		t.Errorf("bottom position = (%d,%d), want (5,%d)", bx, by, wantBy)
	}
}
