package window

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/input"
)

// popupEntry holds a popup with its focus manager and the previously
// focused leaf ID for focus restoration on pop.
type popupEntry struct {
	popup         *PopupWindow
	fm            *input.FocusManager
	prevFocusedID string
}

// Stack manages a MainWindow and its popup overlays. Internal to the framework.
type Stack struct {
	base        *MainWindow
	baseFM      *input.FocusManager
	popups      []popupEntry
	popupOffset OverlayOffset
}

func newStack(base *MainWindow) *Stack {
	fm := input.NewFocusManager(base)
	fm.FocusFirst()
	s := &Stack{base: base, baseFM: fm}
	base.stack = s
	return s
}

func (s *Stack) len() int       { return 1 + len(s.popups) }
func (s *Stack) hasPopup() bool { return len(s.popups) > 0 }

func (s *Stack) topFM() *input.FocusManager {
	if len(s.popups) > 0 {
		return s.popups[len(s.popups)-1].fm
	}
	return s.baseFM
}

func (s *Stack) push(popup *PopupWindow) tea.Cmd {
	var cmds []tea.Cmd

	// Blur current focused leaf and save its ID for restoration
	currentFM := s.topFM()
	var prevFocusedID string
	if leaf := currentFM.FocusedLeaf(); leaf != nil {
		prevFocusedID = leaf.ID()
		if cmd := leaf.SetFocused(false); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Disable current top
	if len(s.popups) > 0 {
		s.popups[len(s.popups)-1].popup.SetEnabled(false)
	} else {
		s.base.SetEnabled(false)
	}

	popup.stack = s
	popup.SetEnabled(true)
	fm := input.NewFocusManager(popup)
	s.popups = append(s.popups, popupEntry{
		popup:         popup,
		fm:            fm,
		prevFocusedID: prevFocusedID,
	})
	if cmd := fm.FocusFirst(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *Stack) pop() tea.Cmd {
	if len(s.popups) == 0 {
		return nil
	}
	entry := s.popups[len(s.popups)-1]
	s.popups = s.popups[:len(s.popups)-1]
	s.popupOffset = OverlayOffset{}

	// Re-enable the new top
	if len(s.popups) > 0 {
		s.popups[len(s.popups)-1].popup.SetEnabled(true)
	} else {
		s.base.SetEnabled(true)
	}

	// Restore focus
	currentFM := s.topFM()
	var cmd tea.Cmd
	if entry.prevFocusedID != "" {
		_, cmd = currentFM.FocusByID(entry.prevFocusedID)
	}
	if currentFM.FocusedLeaf() == nil {
		cmd = currentFM.FocusFirst()
	}
	return cmd
}

func (s *Stack) closePopup(popup *PopupWindow, result any) tea.Cmd {
	var cmds []tea.Cmd
	if cmd := s.pop(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if popup.onResult != nil {
		if cmd := popup.onResult(result); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (s *Stack) update(msg tea.Msg) tea.Cmd {
	// Handle closePopupMsg
	if closeMsg, ok := msg.(closePopupMsg); ok {
		cmd := s.closePopup(closeMsg.popup, closeMsg.result)
		return s.withOnUpdate(cmd)
	}

	// Adjust mouse coordinates for popups
	if mouse, ok := msg.(tea.MouseMsg); ok && s.hasPopup() {
		mouse.X -= s.popupOffset.X
		mouse.Y -= s.popupOffset.Y
		msg = mouse
	}

	fm := s.topFM()
	var cmd tea.Cmd
	if len(s.popups) > 0 {
		cmd = s.popups[len(s.popups)-1].popup.update(msg, fm)
	} else {
		cmd = s.base.update(msg, fm)
	}
	return s.withOnUpdate(cmd)
}

// withOnUpdate appends the current top window's onUpdate callback to the command.
func (s *Stack) withOnUpdate(cmd tea.Cmd) tea.Cmd {
	var onUpdate func() tea.Cmd
	if len(s.popups) > 0 {
		onUpdate = s.popups[len(s.popups)-1].popup.onUpdate
	} else {
		onUpdate = s.base.onUpdate
	}
	if onUpdate == nil {
		return cmd
	}
	updateCmd := onUpdate()
	if cmd == nil {
		return updateCmd
	}
	if updateCmd == nil {
		return cmd
	}
	return tea.Batch(cmd, updateCmd)
}

func (s *Stack) view(width, height int) string {
	s.base.SetSize(width, height)

	if len(s.popups) == 0 {
		return s.base.View()
	}

	top := s.popups[len(s.popups)-1].popup
	// Set popup size: use preferred dimensions if set, otherwise measure natural size.
	pw := top.PreferredWidth()
	ph := top.PreferredHeight()
	if pw > 0 || ph > 0 {
		top.SetSize(pw, ph)
	}
	content := top.View()
	if pw == 0 || ph == 0 {
		// Measure from rendered output for dimensions without preference
		if pw == 0 {
			pw = lipgloss.Width(content)
		}
		if ph == 0 {
			ph = lipgloss.Height(content)
		}
		top.SetSize(pw, ph)
		content = top.View()
	}
	rendered, offset := RenderOverlay(content, top.Title(), width, height)
	s.popupOffset = offset
	return rendered
}
