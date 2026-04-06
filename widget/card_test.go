package widget

import (
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
)

// Compile-time interface check
var _ Component = (*Card)(nil)

func TestCard_Defaults(t *testing.T) {
	c := NewCard("card1", "Errors", "42", DefaultCardStyles())
	if c.ID() != "card1" {
		t.Errorf("ID = %q", c.ID())
	}
	if !c.Visible() {
		t.Error("should be visible")
	}
	if !c.Enabled() {
		t.Error("should be enabled")
	}
}

func TestCard_Interface(t *testing.T) {
	// var _ Component already checked at compile time above;
	// this test exercises it at runtime for clarity.
	var comp Component = NewCard("c", "T", "V", DefaultCardStyles())
	if comp.ID() != "c" {
		t.Errorf("ID = %q", comp.ID())
	}
}

func TestCard_NotFocusable(t *testing.T) {
	c := NewCard("c", "T", "V", DefaultCardStyles())
	if c.Focusable() {
		t.Error("card should not be focusable")
	}
}

func TestCard_View_Normal(t *testing.T) {
	c := NewCard("c", "Errors", "42", DefaultCardStyles())
	output := stripansi.Strip(c.View())
	if !strings.Contains(output, "Errors") {
		t.Errorf("view should contain title 'Errors', got %q", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("view should contain value '42', got %q", output)
	}
}

func TestCard_View_Alert(t *testing.T) {
	c := NewCard("c", "Errors", "99", DefaultCardStyles())
	c.SetAlert(true)
	output := stripansi.Strip(c.View())
	if !strings.Contains(output, "99") {
		t.Errorf("alert view should contain value '99', got %q", output)
	}
}

func TestCard_SetValue(t *testing.T) {
	c := NewCard("c", "Errors", "0", DefaultCardStyles())
	c.SetValue("15")
	output := stripansi.Strip(c.View())
	if !strings.Contains(output, "15") {
		t.Errorf("view should contain updated value '15', got %q", output)
	}
}

func TestCard_SetAlert(t *testing.T) {
	c := NewCard("c", "Errors", "5", DefaultCardStyles())
	if c.alert {
		t.Error("alert should be false by default")
	}
	c.SetAlert(true)
	if !c.alert {
		t.Error("alert should be true after SetAlert(true)")
	}
	c.SetAlert(false)
	if c.alert {
		t.Error("alert should be false after SetAlert(false)")
	}
}

func TestCard_View_Inactive(t *testing.T) {
	c := NewCard("c", "Errors", "42", DefaultCardStyles())
	c.SetEnabled(false)
	output := stripansi.Strip(c.View())
	if !strings.Contains(output, "Errors") {
		t.Errorf("inactive view should still contain title, got %q", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("inactive view should still contain value, got %q", output)
	}
}

func TestCard_KeyBindings_Nil(t *testing.T) {
	c := NewCard("c", "T", "V", DefaultCardStyles())
	if c.KeyBindings() != nil {
		t.Error("card should have nil key bindings")
	}
}

func TestCard_Value(t *testing.T) {
	c := NewCard("c", "Errors", "42", DefaultCardStyles())
	if c.Value() != "42" {
		t.Errorf("Value() = %q, want %q", c.Value(), "42")
	}
	c.SetValue("99")
	if c.Value() != "99" {
		t.Errorf("Value() = %q after SetValue, want %q", c.Value(), "99")
	}
}
