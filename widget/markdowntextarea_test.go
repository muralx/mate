package widget

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

var _ Leaf = (*MarkdownTextArea)(nil)

func init() {
	// Force a color profile so style.Render() actually emits ANSI codes
	// during tests.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestMarkdownTextArea_Defaults(t *testing.T) {
	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	if m.ID() != "mt" {
		t.Errorf("ID = %q, want %q", m.ID(), "mt")
	}
	if !m.Focusable() {
		t.Error("should be focusable")
	}
	if m.Markdown() != "" {
		t.Errorf("Markdown() should start empty, got %q", m.Markdown())
	}
}

func TestMarkdownTextArea_SetMarkdown_StoresSource(t *testing.T) {
	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	m.SetSize(40, 5)
	m.SetMarkdown("# Title")
	if m.Markdown() != "# Title" {
		t.Errorf("Markdown() = %q, want %q", m.Markdown(), "# Title")
	}
}

func TestMarkdownTextArea_SetMarkdown_RendersToContent(t *testing.T) {
	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	m.SetSize(40, 5)
	m.SetMarkdown("# Title")
	if m.Content() == "# Title" {
		t.Error("Content() should return rendered ANSI, not the markdown source")
	}
	if visible := ansi.Strip(m.Content()); visible != "Title" {
		t.Errorf("rendered visible text = %q, want %q", visible, "Title")
	}
}

// SetContent must not be a back door — it should route to SetMarkdown
// so the rendered output is always derived from m.md.
func TestMarkdownTextArea_SetContent_RoutesToSetMarkdown(t *testing.T) {
	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	m.SetSize(40, 5)
	m.SetContent("# Routed")
	if m.Markdown() != "# Routed" {
		t.Errorf("SetContent did not route to SetMarkdown: Markdown() = %q", m.Markdown())
	}
	if visible := ansi.Strip(m.Content()); visible != "Routed" {
		t.Errorf("rendered visible text after SetContent = %q, want %q", visible, "Routed")
	}
}

// Width change must trigger re-render so OSC 8 fallback can kick in.
func TestMarkdownTextArea_SetSize_ReRendersOnWidthChange(t *testing.T) {
	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	m.SetSize(80, 5)
	m.SetMarkdown("see [docs](https://example.com) plus a lot of extra text content here")
	wideOut := m.Content()
	if !strings.Contains(wideOut, "\x1b]8;;") {
		t.Fatalf("expected OSC 8 at width 80: %q", wideOut)
	}

	m.SetSize(10, 5)
	narrowOut := m.Content()
	if strings.Contains(narrowOut, "\x1b]8;;") {
		t.Errorf("expected OSC 8 fallback at width 10, got: %q", narrowOut)
	}
}

func TestMarkdownTextArea_InsidePanel(t *testing.T) {
	panel := NewPanel("p")
	panel.SetBorder(DefaultBorder())
	panel.SetSize(60, 10)
	panel.SetPosition(0, 0)

	m := NewMarkdownTextArea("mt", DefaultMarkdownTextAreaStyles())
	m.SetMarkdown("**hello**")
	panel.Add(m, Next)

	view := panel.View()
	if !strings.Contains(view, "hello") {
		t.Error("markdown text area should be visible inside panel")
	}
	if m.Parent() != panel {
		t.Error("parent should be panel")
	}
}
