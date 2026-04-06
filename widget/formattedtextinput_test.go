package widget

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test validator: requires non-empty
func requireNonEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("required")
	}
	return nil
}

// Test formatter: uppercases
func toUpper(s string) string {
	return strings.ToUpper(s)
}

func TestFormattedTextInput_Interface(t *testing.T) {
	var _ Leaf = (*FormattedTextInput)(nil)
}

func TestFormattedTextInput_ValidateOnBlur(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.WithValidation(requireNonEmpty)
	fti.SetValue("hello")
	fti.SetFocused(true)

	// Blur should trigger validation
	fti.SetFocused(false)
	if fti.Error() != "" {
		t.Errorf("expected no error for non-empty value, got %q", fti.Error())
	}
}

func TestFormattedTextInput_FormatOnBlur(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.WithFormat(toUpper)
	fti.SetValue("hello")
	fti.SetFocused(true)

	fti.SetFocused(false)
	if fti.Value() != "HELLO" {
		t.Errorf("Value() = %q, want %q", fti.Value(), "HELLO")
	}
}

func TestFormattedTextInput_InvalidShowsError(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.WithValidation(requireNonEmpty)
	// Value is empty by default
	fti.SetFocused(true)

	fti.SetFocused(false)
	if fti.Error() != "required" {
		t.Errorf("Error() = %q, want %q", fti.Error(), "required")
	}
}

func TestFormattedTextInput_ValidClearsError(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.WithValidation(requireNonEmpty)

	// First blur with empty — should set error
	fti.SetFocused(true)
	fti.SetFocused(false)
	if fti.Error() == "" {
		t.Fatal("expected error after blur with empty value")
	}

	// Now set value and blur again — should clear error
	fti.SetFocused(true)
	fti.SetValue("valid")
	fti.SetFocused(false)
	if fti.Error() != "" {
		t.Errorf("expected error cleared, got %q", fti.Error())
	}
}

func TestFormattedTextInput_FormatSkippedOnError(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.WithValidation(requireNonEmpty)
	fti.WithFormat(toUpper)

	// Blur with empty value — validation fails, format should not run
	fti.SetFocused(true)
	fti.SetFocused(false)
	// Value remains empty (format not applied)
	if fti.Value() != "" {
		t.Errorf("Value() = %q, want empty (format should be skipped on error)", fti.Value())
	}
}

func TestFormattedTextInput_NoValidate_NoError(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	// No validation set
	fti.SetFocused(true)
	fti.SetFocused(false)
	if fti.Error() != "" {
		t.Errorf("expected no error without validator, got %q", fti.Error())
	}
}

func TestFormattedTextInput_InheritsTextInputBehavior(t *testing.T) {
	fti := NewFormattedTextInput("fti", 20)
	fti.SetFocused(true)

	// Typing should work
	fti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	fti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if fti.Value() != "hi" {
		t.Errorf("Value() = %q, want %q", fti.Value(), "hi")
	}

	// Enter/submit should work
	var submitted string
	fti.OnSubmit(func(v string) tea.Cmd { submitted = v; return nil })
	fti.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if submitted != "hi" {
		t.Errorf("submitted = %q, want %q", submitted, "hi")
	}

	// Focusable should be true
	if !fti.Focusable() {
		t.Error("should be focusable")
	}
}
