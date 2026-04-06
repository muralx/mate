package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestTabComponent_Construction(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	if tc.ID() != "tabs" {
		t.Errorf("ID = %q, want %q", tc.ID(), "tabs")
	}
	if tc.Focusable() {
		t.Error("TabComponent should not be focusable (container)")
	}
}

func TestTabComponent_AddTab(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())

	p1 := NewPanel("p1")
	p1.Add(NewText("t1", "Panel One", lipgloss.NewStyle()), Next)

	p2 := NewPanel("p2")
	p2.Add(NewText("t2", "Panel Two", lipgloss.NewStyle()), Next)

	tc.AddTab("Tab1", p1)
	tc.AddTab("Tab2", p2)

	if tc.ActiveTab() != 0 {
		t.Errorf("active = %d, want 0", tc.ActiveTab())
	}
	// First tab visible, second hidden
	if !p1.Visible() {
		t.Error("first tab panel should be visible")
	}
	if p2.Visible() {
		t.Error("second tab panel should be hidden")
	}
}

func TestTabComponent_View_ShowsActiveContent(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(40, 10)

	p1 := NewPanel("p1")
	p1.Add(NewText("t1", "CONTENT_ONE", lipgloss.NewStyle()), Next)

	p2 := NewPanel("p2")
	p2.Add(NewText("t2", "CONTENT_TWO", lipgloss.NewStyle()), Next)

	tc.AddTab("First", p1)
	tc.AddTab("Second", p2)

	view := stripansi.Strip(tc.View())
	if !strings.Contains(view, "First") {
		t.Error("should show tab header 'First'")
	}
	if !strings.Contains(view, "CONTENT_ONE") {
		t.Error("should show active tab content")
	}
	if strings.Contains(view, "CONTENT_TWO") {
		t.Error("should not show inactive tab content")
	}
}

func TestTabComponent_View_CenterFillsHeight(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(40, 20)

	p1 := NewPanel("p1")
	p1.Add(NewText("t1", "short", lipgloss.NewStyle()), Next)

	tc.AddTab("Tab", p1)
	tc.View()

	// Panel should get remaining height: 20 - 1 (tab bar) = 19
	_, ph := p1.Size()
	if ph != 19 {
		t.Errorf("panel height = %d, want 19 (20 - 1 tab bar)", ph)
	}
}

func TestTabComponent_View_WidthPropagates(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(80, 20)

	p1 := NewPanel("p1")
	p1.Add(NewText("t1", "test", lipgloss.NewStyle()), Next)

	tc.AddTab("Tab", p1)
	tc.View()

	pw, _ := p1.Size()
	if pw != 80 {
		t.Errorf("panel width = %d, want 80", pw)
	}
}

func TestTabComponent_TabBarIsChild(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())

	p1 := NewPanel("p1")
	tc.AddTab("Tab1", p1)

	// TabBar and panel should be children (for focus management)
	children := tc.Children()
	if len(children) != 2 { // bar + 1 panel
		t.Errorf("children = %d, want 2", len(children))
	}
}

func TestTabComponent_SetTabKeyBinding(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.AddTab("Tab1", NewPanel("p1"))
	tc.AddTab("Tab2", NewPanel("p2"))

	tc.SetTabKeyBinding(0, "ctrl+d")
	tc.SetTabKeyBinding(1, "ctrl+e")

	bindings := tc.bar.KeyBindings()
	if len(bindings) != 2 {
		t.Errorf("bindings = %d, want 2", len(bindings))
	}
}

func TestTabComponent_SwitchTab_Visibility(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(60, 20)

	p1 := NewPanel("p1")
	p1.Add(NewText("t1", "FIRST", lipgloss.NewStyle()), Next)
	p2 := NewPanel("p2")
	p2.Add(NewText("t2", "SECOND", lipgloss.NewStyle()), Next)
	p3 := NewPanel("p3")
	p3.Add(NewText("t3", "THIRD", lipgloss.NewStyle()), Next)

	tc.AddTab("Tab1", p1)
	tc.AddTab("Tab2", p2)
	tc.AddTab("Tab3", p3)

	// Initial: only first visible
	view := stripansi.Strip(tc.View())
	if !strings.Contains(view, "FIRST") {
		t.Error("tab 0: should show FIRST")
	}
	if strings.Contains(view, "SECOND") {
		t.Error("tab 0: should not show SECOND")
	}
	if strings.Contains(view, "THIRD") {
		t.Error("tab 0: should not show THIRD")
	}

	// Switch to tab 1
	tc.SetActiveTab(1)
	view = stripansi.Strip(tc.View())
	if strings.Contains(view, "FIRST") {
		t.Error("tab 1: should not show FIRST")
	}
	if !strings.Contains(view, "SECOND") {
		t.Error("tab 1: should show SECOND")
	}

	// Switch to tab 2
	tc.SetActiveTab(2)
	view = stripansi.Strip(tc.View())
	if strings.Contains(view, "SECOND") {
		t.Error("tab 2: should not show SECOND")
	}
	if !strings.Contains(view, "THIRD") {
		t.Error("tab 2: should show THIRD")
	}
}

func TestTabComponent_Nested(t *testing.T) {
	// Outer TabComponent with two tabs
	outer := NewTabComponent("outer", DefaultTabBarStyles())
	outer.SetSize(80, 30)

	// Inner TabComponent as content of outer tab 0
	inner := NewTabComponent("inner", DefaultTabBarStyles())

	innerP1 := NewPanel("inner-p1")
	innerP1.Add(NewText("it1", "INNER_TAB_ONE", lipgloss.NewStyle()), Next)
	innerP2 := NewPanel("inner-p2")
	innerP2.Add(NewText("it2", "INNER_TAB_TWO", lipgloss.NewStyle()), Next)
	inner.AddTab("InnerA", innerP1)
	inner.AddTab("InnerB", innerP2)

	outerP2 := NewPanel("outer-p2")
	outerP2.Add(NewText("ot2", "OUTER_TAB_TWO", lipgloss.NewStyle()), Next)

	outer.AddTab("Outer1", inner)
	outer.AddTab("Outer2", outerP2)

	// Initial: outer tab 0 active, inner tab 0 active
	view := stripansi.Strip(outer.View())
	if !strings.Contains(view, "Outer1") {
		t.Error("should show outer tab header 'Outer1'")
	}
	if !strings.Contains(view, "InnerA") {
		t.Error("should show inner tab header 'InnerA'")
	}
	if !strings.Contains(view, "INNER_TAB_ONE") {
		t.Error("should show inner tab 0 content")
	}
	if strings.Contains(view, "INNER_TAB_TWO") {
		t.Error("should not show inner tab 1 content")
	}
	if strings.Contains(view, "OUTER_TAB_TWO") {
		t.Error("should not show outer tab 1 content")
	}

	// Switch inner tab to 1
	inner.SetActiveTab(1)
	view = stripansi.Strip(outer.View())
	if !strings.Contains(view, "INNER_TAB_TWO") {
		t.Error("after inner switch: should show INNER_TAB_TWO")
	}
	if strings.Contains(view, "INNER_TAB_ONE") {
		t.Error("after inner switch: should not show INNER_TAB_ONE")
	}

	// Switch outer tab to 1 — inner content should disappear
	outer.SetActiveTab(1)
	view = stripansi.Strip(outer.View())
	if !strings.Contains(view, "OUTER_TAB_TWO") {
		t.Error("after outer switch: should show OUTER_TAB_TWO")
	}
	if strings.Contains(view, "INNER_TAB_TWO") {
		t.Error("after outer switch: should not show inner content")
	}
	if strings.Contains(view, "InnerA") {
		t.Error("after outer switch: should not show inner tab headers")
	}

	// Switch back to outer tab 0 — inner should still be on tab 1
	outer.SetActiveTab(0)
	view = stripansi.Strip(outer.View())
	if !strings.Contains(view, "INNER_TAB_TWO") {
		t.Error("after outer switch back: inner should still be on tab 1")
	}
	if strings.Contains(view, "INNER_TAB_ONE") {
		t.Error("after outer switch back: inner tab 0 should still be hidden")
	}
}

func TestTabComponent_Nested_SizePropagation(t *testing.T) {
	// Outer: 80x30. Inner should get outer's center height minus inner's tab bar.
	outer := NewTabComponent("outer", DefaultTabBarStyles())
	outer.SetSize(80, 30)

	inner := NewTabComponent("inner", DefaultTabBarStyles())
	innerPanel := NewPanel("ip")
	innerPanel.Add(NewText("it", "content", lipgloss.NewStyle()), Next)
	inner.AddTab("A", innerPanel)

	outer.AddTab("Tab", inner)
	outer.View()

	// Outer: bar=1, center=29
	// Inner gets 29 from outer center. Inner: bar=1, center=28
	_, innerH := inner.Size()
	if innerH != 29 {
		t.Errorf("inner height = %d, want 29", innerH)
	}
	_, panelH := innerPanel.Size()
	if panelH != 28 {
		t.Errorf("inner panel height = %d, want 28 (29 - 1 tab bar)", panelH)
	}
}

func TestTabComponent_TCBPanel_InsideTab(t *testing.T) {
	// Tab content is a TCB panel with top/center/bottom.
	// Center should fill remaining height.
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(80, 30)

	content := NewPanel("content", TCB)
	content.SetBorder(DefaultBorder())

	header := NewText("header", "HEADER", lipgloss.NewStyle())
	header.SetPreferredHeight(2)

	body := NewText("body", "BODY", lipgloss.NewStyle())

	footer := NewText("footer", "FOOTER", lipgloss.NewStyle())
	footer.SetPreferredHeight(1)

	content.Add(header, TCBTop)
	content.Add(body, TCBCenter)
	content.Add(footer, TCBBottom)

	tc.AddTab("Main", content)
	tc.View()

	// tc: bar=1, center=29 → content gets 29
	// content has border (chrome 4w, 2h) → content area = 29-2 = 27
	// header=2, footer=1 → body = 27-2-1 = 24
	_, bodyH := body.Size()
	if bodyH != 24 {
		t.Errorf("body height = %d, want 24", bodyH)
	}

	// Verify rendering contains all parts
	view := stripansi.Strip(tc.View())
	if !strings.Contains(view, "HEADER") {
		t.Error("should contain HEADER")
	}
	if !strings.Contains(view, "BODY") {
		t.Error("should contain BODY")
	}
	if !strings.Contains(view, "FOOTER") {
		t.Error("should contain FOOTER")
	}
}

func TestTabComponent_OnChange(t *testing.T) {
	called := -1
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(40, 10)
	tc.OnChange(func(i int) tea.Cmd { called = i; return nil })

	p1 := NewPanel("p1")
	p2 := NewPanel("p2")
	tc.AddTab("Tab1", p1)
	tc.AddTab("Tab2", p2)

	// Switch via SetActiveTab triggers bar's onChange which triggers tc.onChange
	tc.SetActiveTab(1)
	// SetActiveTab calls activate + bar.SetActiveTab. The bar's onChange is only
	// fired from bar's Update/HandleEvent, not SetActiveTab. So test via bar.
	// Reset
	called = -1

	// Simulate the bar activating tab — focus on bar, move cursor, press space
	tc.bar.SetFocused(true)
	tc.bar.Update(tea.KeyMsg{Type: tea.KeyRight})
	tc.bar.Update(tea.KeyMsg{Type: tea.KeySpace})

	// The bar's onChange calls tc.activate and tc.onChange
	// But bar.active is already 1 from SetActiveTab above, so cursor is 1 after right
	// and space won't change since cursor==active. Let's set bar back to 0.
	tc.SetActiveTab(0)
	called = -1
	tc.bar.SetFocused(true)
	tc.bar.Update(tea.KeyMsg{Type: tea.KeyRight}) // cursor now 1
	tc.bar.Update(tea.KeyMsg{Type: tea.KeySpace}) // activates tab 1

	if called != 1 {
		t.Errorf("onChange called with %d, want 1", called)
	}
}

func TestTabComponent_TabBar(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	bar := tc.TabBar()
	if bar == nil {
		t.Fatal("TabBar() should not return nil")
	}
	if bar.ID() != "tabs-bar" {
		t.Errorf("TabBar ID = %q, want %q", bar.ID(), "tabs-bar")
	}
}

func TestTabComponent_Activate_OutOfRange(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	p1 := NewPanel("p1")
	tc.AddTab("Tab1", p1)

	// activate with out-of-range should be no-op
	tc.activate(-1)
	if tc.ActiveTab() != 0 {
		t.Errorf("active = %d, want 0 (no change from out-of-range)", tc.ActiveTab())
	}
	tc.activate(5)
	if tc.ActiveTab() != 0 {
		t.Errorf("active = %d, want 0 (no change from out-of-range)", tc.ActiveTab())
	}
}

func TestTabComponent_View_NotVisible(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(40, 10)
	p1 := NewPanel("p1")
	tc.AddTab("Tab1", p1)

	tc.SetVisible(false)
	if tc.View() != "" {
		t.Error("invisible tab component should return empty string")
	}
}

func TestTabComponent_View_NoPanelsVisible(t *testing.T) {
	tc := NewTabComponent("tabs", DefaultTabBarStyles())
	tc.SetSize(40, 10)
	p1 := NewPanel("p1")
	tc.AddTab("Tab1", p1)
	// Hide the panel
	p1.SetVisible(false)

	view := stripansi.Strip(tc.View())
	if !strings.Contains(view, "Tab1") {
		t.Error("should still show tab header even when no panel visible")
	}
}

func TestTabComponent_DeepNesting_ThreeLevels(t *testing.T) {
	// Level 1: outer tabs (80x40)
	// Level 2: inner tabs inside outer tab 0
	// Level 3: a TCB panel inside inner tab 0
	level1 := NewTabComponent("l1", DefaultTabBarStyles())
	level1.SetSize(80, 40)

	level2 := NewTabComponent("l2", DefaultTabBarStyles())

	level3 := NewPanel("l3", TCB)
	topText := NewText("top", "L3_TOP", lipgloss.NewStyle())
	topText.SetPreferredHeight(1)
	centerText := NewText("center", "L3_CENTER", lipgloss.NewStyle())
	bottomText := NewText("bottom", "L3_BOTTOM", lipgloss.NewStyle())
	bottomText.SetPreferredHeight(1)
	level3.Add(topText, TCBTop)
	level3.Add(centerText, TCBCenter)
	level3.Add(bottomText, TCBBottom)

	level2.AddTab("Inner", level3)
	level1.AddTab("Outer", level2)

	view := stripansi.Strip(level1.View())

	// All three levels should render
	if !strings.Contains(view, "Outer") {
		t.Error("should show level 1 tab header")
	}
	if !strings.Contains(view, "Inner") {
		t.Error("should show level 2 tab header")
	}
	if !strings.Contains(view, "L3_TOP") {
		t.Error("should show level 3 top")
	}
	if !strings.Contains(view, "L3_CENTER") {
		t.Error("should show level 3 center")
	}
	if !strings.Contains(view, "L3_BOTTOM") {
		t.Error("should show level 3 bottom")
	}

	// Size propagation: 40 - 1 (l1 bar) - 1 (l2 bar) = 38 for level3
	// level3 has no border: top=1, bottom=1, center=36
	_, centerH := centerText.Size()
	if centerH != 36 {
		t.Errorf("level 3 center height = %d, want 36", centerH)
	}
}
