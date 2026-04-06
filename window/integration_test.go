package window

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
)

// ---------------------------------------------------------------------------
// Integration tests: verify full layout flow from Window → Panel → children
// ---------------------------------------------------------------------------

// TestIntegration_TCBWindow_SizePropagation verifies that a TCB window
// distributes its viewport size correctly to top/center/bottom.
func TestIntegration_TCBWindow_SizePropagation(t *testing.T) {
	win := NewWindow("main")

	topBar := widget.NewText("top", "Header", lipgloss.NewStyle())
	topBar.SetPreferredHeight(1)

	centerPanel := widget.NewPanel("center", widget.TCB)
	centerPanel.SetBorder(widget.DefaultBorder())

	bottomBar := widget.NewText("bottom", "Status", lipgloss.NewStyle())
	bottomBar.SetPreferredHeight(1)

	win.Add(topBar, widget.TCBTop)
	win.Add(centerPanel, widget.TCBCenter)
	win.Add(bottomBar, widget.TCBBottom)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Top: h=1, full width
	tw, th := topBar.Size()
	if tw != 80 {
		t.Errorf("top width = %d, want 80", tw)
	}
	if th != 1 {
		t.Errorf("top height = %d, want 1", th)
	}

	// Bottom: h=1, full width
	bw, bh := bottomBar.Size()
	if bw != 80 {
		t.Errorf("bottom width = %d, want 80", bw)
	}
	if bh != 1 {
		t.Errorf("bottom height = %d, want 1", bh)
	}

	// Center: gets remaining height (24 - 1 - 1 = 22)
	cw, ch := centerPanel.Size()
	if cw != 80 {
		t.Errorf("center width = %d, want 80", cw)
	}
	if ch != 22 {
		t.Errorf("center height = %d, want 22", ch)
	}
}

// TestIntegration_NestedTCB verifies TCB inside TCB: Window(TCB) → center is Panel(TCB).
func TestIntegration_NestedTCB(t *testing.T) {
	win := NewWindow("main")

	tabs := widget.NewText("tabs", "Tab1 | Tab2", lipgloss.NewStyle())
	tabs.SetPreferredHeight(1)

	// Dashboard panel: TCB with border
	dash := widget.NewPanel("dash", widget.TCB)
	dash.SetBorder(widget.DefaultBorder())

	cardRow := widget.NewText("cards", "CPU | MEM | DISK", lipgloss.NewStyle())
	cardRow.SetPreferredHeight(4)

	table := widget.NewText("table", "row1\nrow2\nrow3", lipgloss.NewStyle())
	// No preferred height — center slot stretches it

	detail := widget.NewText("detail", "Select a row", lipgloss.NewStyle())
	detail.SetPreferredHeight(5)

	dash.Add(cardRow, widget.TCBTop)
	dash.Add(table, widget.TCBCenter)
	dash.Add(detail, widget.TCBBottom)

	status := widget.NewText("status", "Ready", lipgloss.NewStyle())
	status.SetPreferredHeight(1)

	win.Add(tabs, widget.TCBTop)
	win.Add(dash, widget.TCBCenter)
	win.Add(status, widget.TCBBottom)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Window: 80x24
	// tabs: h=1, status: h=1 → dash gets 22
	dw, dh := dash.Size()
	if dw != 80 {
		t.Errorf("dash width = %d, want 80", dw)
	}
	if dh != 22 {
		t.Errorf("dash height = %d, want 22", dh)
	}

	// Inside dash (bordered: chrome 4w, 2h → content 76x20):
	// cardRow: h=4, detail: h=5 → table gets 20 - 4 - 5 = 11
	_, tableH := table.Size()
	expectedTableH := 20 - 4 - 5 // content height - cardRow - detail
	if tableH != expectedTableH {
		t.Errorf("table height = %d, want %d", tableH, expectedTableH)
	}

	// Card row width = content width (76)
	cardW, _ := cardRow.Size()
	if cardW != 76 {
		t.Errorf("cardRow width = %d, want 76", cardW)
	}
}

// TestIntegration_VerticalWindow verifies a Window with Vertical layout.
func TestIntegration_VerticalWindow(t *testing.T) {
	win := NewWindow("main", widget.Vertical)

	btn1 := widget.NewButton("b1", "First", widget.DefaultButtonStyles())
	btn2 := widget.NewButton("b2", "Second", widget.DefaultButtonStyles())

	win.Add(btn1, widget.Next)
	win.Add(btn2, widget.Next)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := app.View()

	plain := stripansi.Strip(view)
	if !strings.Contains(plain, "First") || !strings.Contains(plain, "Second") {
		t.Errorf("should contain both buttons, got:\n%s", plain)
	}

	// Both buttons get full width
	w1, _ := btn1.Size()
	w2, _ := btn2.Size()
	if w1 != 80 {
		t.Errorf("btn1 width = %d, want 80", w1)
	}
	if w2 != 80 {
		t.Errorf("btn2 width = %d, want 80", w2)
	}
}

// TestIntegration_HorizontalPanel verifies a Panel with Horizontal layout inside a Window.
func TestIntegration_HorizontalPanel(t *testing.T) {
	win := NewWindow("main")

	row := widget.NewPanel("row", widget.Horizontal)
	row.SetSpacing(2)

	btn1 := widget.NewButton("b1", "A", widget.DefaultButtonStyles())
	btn1.SetPreferredWidth(10)
	btn2 := widget.NewButton("b2", "B", widget.DefaultButtonStyles())
	btn2.SetPreferredWidth(10)

	row.Add(btn1, widget.Next)
	row.Add(btn2, widget.Next)

	win.Add(row, widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Row gets full width and center height (24, since no top/bottom)
	rw, rh := row.Size()
	if rw != 80 {
		t.Errorf("row width = %d, want 80", rw)
	}
	if rh != 24 {
		t.Errorf("row height = %d, want 24", rh)
	}

	// Buttons get preferred width and row's height
	w1, h1 := btn1.Size()
	if w1 != 10 {
		t.Errorf("btn1 width = %d, want 10", w1)
	}
	if h1 != 24 {
		t.Errorf("btn1 height = %d, want 24 (row height)", h1)
	}
}

// TestIntegration_Positions verifies mouse hit testing positions are correct
// in a nested TCB layout.
func TestIntegration_Positions(t *testing.T) {
	win := NewWindow("main")

	header := widget.NewText("header", "Title", lipgloss.NewStyle())
	header.SetPreferredHeight(1)

	btn := widget.NewButton("btn", "Click Me", widget.DefaultButtonStyles())

	footer := widget.NewText("footer", "Status", lipgloss.NewStyle())
	footer.SetPreferredHeight(1)

	win.Add(header, widget.TCBTop)
	win.Add(btn, widget.TCBCenter)
	win.Add(footer, widget.TCBBottom)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Header at y=0
	_, hy := header.Position()
	if hy != 0 {
		t.Errorf("header y = %d, want 0", hy)
	}

	// Button at y=1 (after header)
	_, by := btn.Position()
	if by != 1 {
		t.Errorf("btn y = %d, want 1", by)
	}

	// Footer at y=23 (24 - 1)
	_, fy := footer.Position()
	if fy != 23 {
		t.Errorf("footer y = %d, want 23", fy)
	}
}

// TestIntegration_BorderedPanelInTCB verifies a bordered panel inside a
// window's center slot has correct content area.
func TestIntegration_BorderedPanelInTCB(t *testing.T) {
	win := NewWindow("main")

	panel := widget.NewPanel("p", widget.Vertical)
	panel.SetBorder(widget.DefaultBorder())
	panel.SetTitle("Settings")

	btn := widget.NewButton("b", "Save", widget.DefaultButtonStyles())
	panel.Add(btn, widget.Next)

	win.Add(panel, widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Panel gets 80x24 (full viewport, only center slot)
	pw, ph := panel.Size()
	if pw != 80 {
		t.Errorf("panel width = %d, want 80", pw)
	}
	if ph != 24 {
		t.Errorf("panel height = %d, want 24", ph)
	}

	// Button gets content width: 80 - 4 (border+padding) = 76
	bw, _ := btn.Size()
	if bw != 76 {
		t.Errorf("btn width = %d, want 76 (content width)", bw)
	}
}

// TestIntegration_FocusCycling verifies Tab works across nested layout.
func TestIntegration_FocusCycling(t *testing.T) {
	win := NewWindow("main")

	btn1 := widget.NewButton("b1", "Top", widget.DefaultButtonStyles())
	btn1.SetPreferredHeight(1)
	btn2 := widget.NewButton("b2", "Center", widget.DefaultButtonStyles())
	btn3 := widget.NewButton("b3", "Bottom", widget.DefaultButtonStyles())
	btn3.SetPreferredHeight(1)

	win.Add(btn1, widget.TCBTop)
	win.Add(btn2, widget.TCBCenter)
	win.Add(btn3, widget.TCBBottom)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app.View()

	// Initial focus on first
	if !btn1.Focused() {
		t.Error("btn1 should be focused initially")
	}

	// Tab cycles through all three
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !btn2.Focused() {
		t.Error("after Tab 1: btn2 should be focused")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !btn3.Focused() {
		t.Error("after Tab 2: btn3 should be focused")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !btn1.Focused() {
		t.Error("after Tab 3: btn1 should be focused (wrap)")
	}
}

// --- RenderOverlay coverage ---

func TestRenderOverlay_SmallTerminal_ClampsPosition(t *testing.T) {
	// Very small terminal that forces clamping of startY and startX
	content := "Some popup content here"
	rendered, offset := RenderOverlay(content, "Title", 10, 5)

	// With a 10-wide terminal the popup can't center normally; startX clamps to 0
	if offset.X < 0 {
		t.Error("offset X should not be negative")
	}
	if offset.Y < 1 {
		// startY clamps to minimum 1
		t.Errorf("offset Y = %d, should be at least 1", offset.Y)
	}
	if rendered == "" {
		t.Error("should produce non-empty output even for small terminal")
	}
}

func TestRenderOverlay_NoTitle(t *testing.T) {
	content := "Just content"
	rendered, offset := RenderOverlay(content, "", 80, 24)

	if !strings.Contains(rendered, "Just content") {
		t.Error("should contain the content")
	}
	// Without title, offsetY should not include title line adjustment
	// Title adds +1 to offsetY, so no-title should be 1 less
	renderedWithTitle, offsetWithTitle := RenderOverlay(content, "Title", 80, 24)
	_ = renderedWithTitle
	if offset.Y >= offsetWithTitle.Y {
		t.Error("no-title offsetY should be less than with-title offsetY")
	}
}

func TestRenderOverlay_WidePopup_ClampsWidth(t *testing.T) {
	// Create content wider than the terminal
	wideContent := strings.Repeat("X", 100)
	rendered, offset := RenderOverlay(wideContent, "Title", 40, 20)

	if offset.X < 0 {
		t.Error("offset X should not be negative")
	}
	if rendered == "" {
		t.Error("should produce non-empty output")
	}
	// The popup width should be clamped to width-4 = 36
	lines := strings.Split(rendered, "\n")
	if len(lines) != 20 {
		t.Errorf("output should have %d lines, got %d", 20, len(lines))
	}
}

// TestIntegration_PopupLayout verifies popups render correctly.
func TestIntegration_PopupLayout(t *testing.T) {
	win := NewWindow("main")
	win.Add(widget.NewButton("b", "Main", widget.DefaultButtonStyles()), widget.TCBCenter)

	app := NewApp(win)
	app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Create and show popup
	popup := NewPopupWindow("popup", "Confirm", DefaultPopupStyles())
	yesBtn := widget.NewButton("yes", "Yes", widget.DefaultButtonStyles())
	noBtn := widget.NewButton("no", "No", widget.DefaultButtonStyles())
	popup.Add(yesBtn, widget.TCBCenter)
	popup.Add(noBtn, widget.TCBBottom)

	win.ShowPopup(popup)

	view := app.View()
	plain := stripansi.Strip(view)

	if !strings.Contains(plain, "Yes") {
		t.Error("popup should contain 'Yes' button")
	}
	if !strings.Contains(plain, "No") {
		t.Error("popup should contain 'No' button")
	}
	if !strings.Contains(plain, "Confirm") {
		t.Error("popup should contain title 'Confirm'")
	}
}
