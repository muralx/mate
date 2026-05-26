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
	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			expanded := strings.ReplaceAll(line, "\t", "    ")
			out = append(out, r.styles.CodeBlock.Render(expanded))
			continue
		}

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

		// Horizontal rules.
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			out = append(out, strings.Repeat("─", 60))
			continue
		}

		// Table separator — strip.
		if isTableSeparator(line) {
			continue
		}

		// Table rows — pass through with tab expansion.
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			out = append(out, strings.ReplaceAll(line, "\t", "    "))
			continue
		}

		// Tab expansion before inline processing.
		line = strings.ReplaceAll(line, "\t", "    ")
		out = append(out, r.renderInlineLine(line, maxWidth))
	}
	return strings.Join(out, "\n")
}

// renderInlineLine renders one line with OSC 8 hyperlinks, falling
// back to plain styled link text when the visible width would exceed
// maxWidth (OSC 8 across wrapped lines is fragile in some terminals).
func (r *Renderer) renderInlineLine(line string, maxWidth int) string {
	osc8 := r.renderInline(line, true)
	if maxWidth > 0 && lipgloss.Width(osc8) > maxWidth {
		return r.renderInline(line, false)
	}
	return osc8
}

// isTableSeparator reports whether line is a markdown table separator
// row (only `|`, `-`, `:`, and whitespace).
func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	cleaned := strings.ReplaceAll(trimmed, "|", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	return strings.TrimSpace(cleaned) == ""
}

// renderInline tokenizes one line of inline markdown in a single
// left-to-right pass, handling code spans, bold, and links. Doing it
// in one pass (rather than sequential string scans) keeps later
// patterns from looking inside earlier ones — e.g., the `**` inside
// a code span is skipped because we jump past the closing backtick
// before scanning for bold markers. With osc8=false, links render as
// plain styled text (used by the per-line fallback in renderInlineLine).
func (r *Renderer) renderInline(line string, osc8 bool) string {
	var sb strings.Builder
	sb.Grow(len(line))
	i := 0
	for i < len(line) {
		// Inline code: `...`
		if line[i] == '`' {
			end := strings.Index(line[i+1:], "`")
			if end < 0 {
				sb.WriteByte(line[i])
				i++
				continue
			}
			end += i + 1
			sb.WriteString(r.styles.Code.Render(line[i+1 : end]))
			i = end + 1
			continue
		}
		// Bold: **...**
		if i+1 < len(line) && line[i] == '*' && line[i+1] == '*' {
			end := strings.Index(line[i+2:], "**")
			if end < 0 {
				sb.WriteByte(line[i])
				i++
				continue
			}
			end += i + 2
			sb.WriteString(r.styles.Bold.Render(line[i+2 : end]))
			i = end + 2
			continue
		}
		// Link: [text](url)
		if line[i] == '[' {
			closeBracket := strings.Index(line[i+1:], "](")
			if closeBracket < 0 {
				sb.WriteByte(line[i])
				i++
				continue
			}
			urlStart := i + 1 + closeBracket + 2
			urlLen := strings.Index(line[urlStart:], ")")
			if urlLen < 0 {
				sb.WriteByte(line[i])
				i++
				continue
			}
			text := line[i+1 : i+1+closeBracket]
			url := line[urlStart : urlStart+urlLen]
			styled := r.styles.Link.Render(text)
			if osc8 {
				styled = osc8Link(styled, url)
			}
			sb.WriteString(styled)
			i = urlStart + urlLen + 1
			continue
		}
		sb.WriteByte(line[i])
		i++
	}
	return sb.String()
}

// osc8Link wraps text in an OSC 8 terminal hyperlink that opens url
// on click. Uses BEL (\x07) as the OSC terminator — more widely
// supported than ST (\x1b\\), especially under tmux/screen and older
// terminals.
func osc8Link(text, url string) string {
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}
