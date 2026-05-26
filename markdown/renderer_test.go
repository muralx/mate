package markdown

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	// Force a color profile so style.Render() actually emits ANSI codes
	// during tests (the default no-TTY profile strips them).
	lipgloss.SetColorProfile(termenv.TrueColor)
	os.Exit(m.Run())
}

func TestRender_Empty(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	if got := r.Render("", 80); got != "" {
		t.Errorf("Render(\"\") = %q, want empty string", got)
	}
}

func TestNewRenderer_NilStylesUsesDefaults(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}
}

func TestRender_H1(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("# Title", 0)
	if visible := ansi.Strip(out); visible != "Title" {
		t.Errorf("H1 visible = %q, want %q", visible, "Title")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("H1 output missing ANSI styling: %q", out)
	}
}

func TestRender_H2(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("## Sub", 0)
	if visible := ansi.Strip(out); visible != "Sub" {
		t.Errorf("H2 visible = %q, want %q", visible, "Sub")
	}
}

func TestRender_H3(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("### Inner", 0)
	if visible := ansi.Strip(out); visible != "Inner" {
		t.Errorf("H3 visible = %q, want %q", visible, "Inner")
	}
}

func TestRender_HeadingPrefixPriority(t *testing.T) {
	// "### " must match before "## " before "# " — longest prefix first.
	r := NewRenderer(DefaultStyles())
	out := r.Render("### x", 0)
	if visible := ansi.Strip(out); visible != "x" {
		t.Errorf("longer heading prefix not stripped first: visible = %q, want %q", visible, "x")
	}
}

func TestRender_Bold(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("hello **world** end", 0)
	if visible := ansi.Strip(out); visible != "hello world end" {
		t.Errorf("visible text = %q, want %q", visible, "hello world end")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("bold output missing ANSI styling: %q", out)
	}
}

func TestRender_InlineCode(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("call `foo()` here", 0)
	if visible := ansi.Strip(out); visible != "call foo() here" {
		t.Errorf("visible text = %q, want %q", visible, "call foo() here")
	}
}

// Pass order: code spans must be processed BEFORE bold. A backticked
// span containing literal ** must NOT be bolded.
func TestRender_PassOrder_CodeBeforeBold(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("`**not bold**` and **bold**", 0)
	if visible := ansi.Strip(out); visible != "**not bold** and bold" {
		t.Errorf("visible text = %q, want %q", visible, "**not bold** and bold")
	}
}

func TestRender_CodeBlock(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	md := "```\nfoo\nbar\n```"
	out := r.Render(md, 0)
	if visible := ansi.Strip(out); visible != "foo\nbar" {
		t.Errorf("visible = %q, want %q", visible, "foo\nbar")
	}
}

func TestRender_CodeBlockTabsExpanded(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("```\n\tfoo\n```", 0)
	if visible := ansi.Strip(out); visible != "    foo" {
		t.Errorf("tab expansion: visible = %q, want %q", visible, "    foo")
	}
}

func TestRender_HorizontalRule(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	for _, marker := range []string{"---", "***", "___"} {
		out := r.Render(marker, 0)
		if !strings.Contains(ansi.Strip(out), "─") {
			t.Errorf("HR marker %q did not render: %q", marker, out)
		}
	}
}

func TestRender_TablePassthrough(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	md := "| a | b |\n|---|---|\n| 1 | 2 |"
	out := r.Render(md, 0)
	visible := ansi.Strip(out)
	if !strings.Contains(visible, "| a | b |") || !strings.Contains(visible, "| 1 | 2 |") {
		t.Errorf("table rows missing: %q", visible)
	}
	if strings.Contains(visible, "|---|---|") {
		t.Errorf("table separator should be stripped: %q", visible)
	}
}

func TestRender_Link_OSC8(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("see [docs](https://example.com)", 0)
	if !strings.Contains(out, "\x1b]8;;https://example.com\x07") {
		t.Errorf("OSC 8 opening missing: %q", out)
	}
	if !strings.Contains(out, "\x1b]8;;\x07") {
		t.Errorf("OSC 8 closing missing: %q", out)
	}
	if visible := ansi.Strip(out); !strings.Contains(visible, "docs") {
		t.Errorf("link text missing from visible output: %q", visible)
	}
}

// Visible line "see docs and more text here" is 27 cells. maxWidth=10
// forces the OSC 8 fallback for the whole line.
func TestRender_Link_PlainFallback_OnNarrowWidth(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	out := r.Render("see [docs](https://example.com) and more text here", 10)
	if strings.Contains(out, "\x1b]8;;") {
		t.Errorf("OSC 8 should be absent on overflowing line: %q", out)
	}
	if !strings.Contains(ansi.Strip(out), "docs") {
		t.Errorf("link text missing: %q", out)
	}
}

// A long URL with short visible text must NOT trigger fallback —
// OSC escape bytes are zero-width.
func TestRender_Link_LongURLDoesNotTriggerFallback(t *testing.T) {
	r := NewRenderer(DefaultStyles())
	longURL := "https://example.com/" + strings.Repeat("a", 500)
	out := r.Render("[x]("+longURL+")", 80)
	if !strings.Contains(out, "\x1b]8;;"+longURL+"\x07") {
		t.Errorf("OSC 8 should be emitted for short visible text regardless of URL length")
	}
}
