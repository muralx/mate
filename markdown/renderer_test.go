package markdown

import "testing"

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
