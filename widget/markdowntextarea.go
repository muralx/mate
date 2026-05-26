package widget

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/muralx/mate/markdown"
)

// MarkdownTextArea is a focusable, scrollable component that takes
// Markdown source via SetMarkdown and renders it to ANSI-styled
// terminal text on top of ScrollableText. It re-renders on width
// changes so OSC 8 links fall back to plain styled text when a line
// would exceed the viewport.
type MarkdownTextArea struct {
	ScrollableText
	renderer *markdown.Renderer
	md       string
	lastW    int
}

// MarkdownTextAreaStyles bundles the styles used by MarkdownTextArea —
// the ScrollableText viewport styles and the per-element markdown
// styles.
type MarkdownTextAreaStyles struct {
	Scroll   ScrollableTextStyles
	Markdown markdown.Styles
}

// DefaultMarkdownTextAreaStyles returns sensible defaults.
func DefaultMarkdownTextAreaStyles() MarkdownTextAreaStyles {
	return MarkdownTextAreaStyles{
		Scroll:   DefaultScrollableTextStyles(),
		Markdown: markdown.DefaultStyles(),
	}
}

// NewMarkdownTextArea returns a MarkdownTextArea with the given styles.
func NewMarkdownTextArea(id string, styles MarkdownTextAreaStyles) *MarkdownTextArea {
	m := &MarkdownTextArea{
		renderer: markdown.NewRenderer(styles.Markdown),
	}
	m.ScrollableText = *NewScrollableText(id, styles.Scroll)
	return m
}

// SetMarkdown replaces the markdown source and re-renders at the
// current viewport width.
func (m *MarkdownTextArea) SetMarkdown(md string) {
	m.md = md
	m.render()
}

// Markdown returns the current markdown source. For the rendered ANSI
// string see Content() (inherited from ScrollableText).
func (m *MarkdownTextArea) Markdown() string { return m.md }

// SetStyles swaps the styles and re-renders.
func (m *MarkdownTextArea) SetStyles(s MarkdownTextAreaStyles) {
	m.renderer = markdown.NewRenderer(s.Markdown)
	m.render()
}

// SetContent overrides ScrollableText.SetContent to route raw input
// through the markdown renderer. Without this override callers could
// silently desync Markdown() from the displayed content.
func (m *MarkdownTextArea) SetContent(s string) { m.SetMarkdown(s) }

// SetSize re-renders only when the width changes — Markdown→ANSI is
// cheap but repeated per-layout-pass renders are still wasteful.
func (m *MarkdownTextArea) SetSize(w, h int) {
	m.ScrollableText.SetSize(w, h)
	if m.md != "" && w != m.lastW {
		m.lastW = w
		m.render()
	}
}

// Update forwards to ScrollableText so MarkdownTextArea is a proper Leaf.
func (m *MarkdownTextArea) Update(msg tea.KeyMsg) (tea.Cmd, bool) {
	return m.ScrollableText.Update(msg)
}

func (m *MarkdownTextArea) render() {
	w, _ := m.Size()
	m.ScrollableText.SetContent(m.renderer.Render(m.md, w))
}
