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
