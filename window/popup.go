package window

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/input"
	"github.com/muralx/mate/widget"
)

// PopupStyles defines the visual appearance of a popup window.
type PopupStyles struct {
	Border lipgloss.Style
}

// DefaultPopupStyles returns sensible defaults for popup windows.
func DefaultPopupStyles() PopupStyles {
	return PopupStyles{
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1),
	}
}

// closePopupMsg is an internal message that triggers popup close via the Stack.
type closePopupMsg struct {
	popup  *PopupWindow
	result any
}

// PopupWindow is a container rendered as a centered overlay.
// It embeds BaseWindow for event routing and adds Close/OnResult/Escape.
type PopupWindow struct {
	BaseWindow
	title    string
	styles   PopupStyles
	onResult func(value any) tea.Cmd
}

// NewPopupWindow creates a PopupWindow with the given ID, title, and styles.
// Optional layout parameter defaults to TCB.
func NewPopupWindow(id, title string, styles PopupStyles, layout ...widget.Layout) *PopupWindow {
	p := &PopupWindow{
		title:  title,
		styles: styles,
	}
	p.BaseWindow = newBaseWindow(id, p, layout...)
	return p
}

// Title returns the popup's title.
func (p *PopupWindow) Title() string { return p.title }

// OnResult sets a handler called when the popup is closed.
// Called for both confirmed (non-nil result) and cancelled (nil result).
// Fires AFTER the popup is popped from the stack.
func (p *PopupWindow) OnResult(fn func(value any) tea.Cmd) {
	p.onResult = fn
}

// Close returns a command that closes this popup and delivers the result.
// The Stack processes this: pops the popup, then calls OnResult.
func (p *PopupWindow) Close(result any) tea.Cmd {
	return func() tea.Msg {
		return closePopupMsg{popup: p, result: result}
	}
}

// update intercepts Escape to close, then delegates to BaseWindow for all other events.
func (p *PopupWindow) update(msg tea.Msg, fm *input.FocusManager) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEscape {
		return p.Close(nil)
	}
	return p.BaseWindow.update(msg, fm)
}
