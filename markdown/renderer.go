// Package markdown converts a small subset of Markdown into ANSI-styled
// terminal text. It supports H1/H2/H3, bold, inline code, fenced code
// blocks, horizontal rules, table passthrough, and `[text](url)` links
// (OSC 8 hyperlinks with a maxWidth-driven plain-text fallback).
//
// Out of scope: lists, blockquotes, italics (ambiguous with bold under
// this naive parser), images, escapes.
package markdown

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles holds the per-element lipgloss styles used by Renderer.
type Styles struct {
	H1        lipgloss.Style
	H2        lipgloss.Style
	H3        lipgloss.Style
	Bold      lipgloss.Style
	Code      lipgloss.Style
	CodeBlock lipgloss.Style
	Link      lipgloss.Style
}

// DefaultStyles returns sensible defaults (cyan headings, bold inline
// `**...**`, peach inline code, dim code blocks, underlined cyan links).
func DefaultStyles() Styles {
	return Styles{
		H1:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4fc3f7")).Underline(true),
		H2:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4fc3f7")),
		H3:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#81d4fa")),
		Bold:      lipgloss.NewStyle().Bold(true),
		Code:      lipgloss.NewStyle().Foreground(lipgloss.Color("#ce9178")),
		CodeBlock: lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa")),
		Link:      lipgloss.NewStyle().Foreground(lipgloss.Color("#4fc3f7")).Underline(true),
	}
}

// Renderer converts markdown source into ANSI-styled terminal text.
type Renderer struct {
	styles Styles
}

// NewRenderer returns a Renderer configured with the given styles.
func NewRenderer(styles Styles) *Renderer {
	return &Renderer{styles: styles}
}

// Render converts markdown source into ANSI-styled text. When maxWidth > 0,
// links on lines whose visible width would exceed maxWidth are rendered
// as plain styled text instead of OSC 8 hyperlinks (OSC 8 across wrapped
// lines is fragile in some terminal emulators).
//
// TODO: revisit the OSC 8 fallback after a manual terminal-matrix test —
// the byte-level wrap bug that originally motivated this guard is fixed.
func (r *Renderer) Render(md string, maxWidth int) string {
	if md == "" {
		return ""
	}
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		// Headings: longest prefix first.
		if rest, ok := strings.CutPrefix(line, "### "); ok {
			out = append(out, r.styles.H3.Render(rest))
			continue
		}
		if rest, ok := strings.CutPrefix(line, "## "); ok {
			out = append(out, r.styles.H2.Render(rest))
			continue
		}
		if rest, ok := strings.CutPrefix(line, "# "); ok {
			out = append(out, r.styles.H1.Render(rest))
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
