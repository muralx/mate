package widget

import tea "github.com/charmbracelet/bubbletea"

// FormattedTextInput extends TextInput with validation and formatting.
type FormattedTextInput struct {
	TextInput
	validate func(string) error
	format   func(string) string
}

// NewFormattedTextInput creates a new FormattedTextInput with the given ID and input width.
func NewFormattedTextInput(id string, inputWidth int) *FormattedTextInput {
	fti := &FormattedTextInput{}
	fti.TextInput = *NewTextInput(id, inputWidth)
	return fti
}

// WithValidation sets the validation function. Validation runs on blur.
func (fti *FormattedTextInput) WithValidation(fn func(string) error) {
	fti.validate = fn
}

// WithFormat sets the format function applied on blur after successful validation.
func (fti *FormattedTextInput) WithFormat(fn func(string) string) {
	fti.format = fn
}

// SetFocused sets the focused state. On blur, it runs validation and formatting.
func (fti *FormattedTextInput) SetFocused(focused bool) tea.Cmd {
	cmd := fti.TextInput.SetFocused(focused)
	if !focused {
		fti.runValidation()
		fti.runFormat()
	}
	return cmd
}

func (fti *FormattedTextInput) runValidation() {
	if fti.validate == nil {
		return
	}
	err := fti.validate(fti.Value())
	if err != nil {
		fti.SetError(err.Error())
	} else {
		fti.SetError("")
	}
}

func (fti *FormattedTextInput) runFormat() {
	if fti.format == nil || fti.Error() != "" {
		return
	}
	formatted := fti.format(fti.Value())
	fti.SetValue(formatted)
}
