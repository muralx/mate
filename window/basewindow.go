package window

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/input"
	"github.com/muralx/mate/widget"
)

// BaseWindow is the shared base for Window and PopupWindow. It provides:
//   - Container behavior (embeds widget.BaseContainer)
//   - A borderless content Panel with configurable layout (default TCB)
//   - Keyboard event routing: Tab/Shift-Tab focus cycling, key binding resolution,
//     focused leaf dispatch, OnKeyPress fallthrough
//   - Mouse event routing: hit testing, focus change on press, MouseClickEvent dispatch
//   - ShowPopup method (delegates to stack.push)
//   - OnKeyPress callback setter
//
// BaseWindow is NOT instantiated directly by users. It's embedded by Window and PopupWindow.
type BaseWindow struct {
	widget.BaseContainer
	content    *widget.Panel
	onKeyPress func(tea.KeyMsg) tea.Cmd
	onUpdate   func() tea.Cmd
	stack      *Stack
}

// newBaseWindow creates a BaseWindow with the given ID.
// The self parameter is the concrete outer type (Window or PopupWindow) for
// correct parent back-references, same pattern as widget.NewBaseContainer.
// Optional layout parameter defaults to TCB.
func newBaseWindow(id string, self widget.Container, layout ...widget.Layout) BaseWindow {
	l := widget.TCB
	if len(layout) > 0 {
		l = layout[0]
	}
	bw := BaseWindow{}
	bw.BaseContainer = *widget.NewBaseContainer(id, self)
	bw.content = widget.NewPanel(id+"-content", l)
	bw.BaseContainer.AddChild(bw.content)
	return bw
}

// Add places a child component at the given position in the content panel.
func (bw *BaseWindow) Add(child widget.Component, position widget.Position) {
	bw.content.Add(child, position)
}

// OnKeyPress sets a handler called when a key is not consumed by focus cycling,
// key binding resolution, or the focused leaf's Update.
func (bw *BaseWindow) OnKeyPress(fn func(tea.KeyMsg) tea.Cmd) {
	bw.onKeyPress = fn
}

// OnUpdate sets a handler called after every event that changes state (focus,
// key bindings, etc.). Use this to update status bars or other reactive UI.
func (bw *BaseWindow) OnUpdate(fn func() tea.Cmd) {
	bw.onUpdate = fn
}

// ActiveKeyBindings returns all registered key bindings from visible, active
// components in the current window's component tree. Useful for building
// status bars or help overlays.
func (bw *BaseWindow) ActiveKeyBindings() []key.Binding {
	if bw.stack == nil {
		return nil
	}
	return bw.stack.topFM().AllActiveKeyBindings()
}

// ShowPopup pushes a popup window onto the stack. Returns nil if no stack is set.
func (bw *BaseWindow) ShowPopup(popup *PopupWindow) tea.Cmd {
	if bw.stack == nil {
		return nil
	}
	return bw.stack.push(popup)
}

// View renders the content panel, setting its size and position from the window.
func (bw *BaseWindow) View() string {
	pw, ph := bw.Size()
	px, py := bw.Position()
	bw.content.SetSize(pw, ph)
	bw.content.SetPosition(px, py)
	output := bw.content.View()
	if pw > 0 && ph > 0 {
		output = lipgloss.NewStyle().Width(pw).Height(ph).Render(output)
	}
	return output
}

// update dispatches a tea.Msg to the appropriate handler (key or mouse).
func (bw *BaseWindow) update(msg tea.Msg, fm *input.FocusManager) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return bw.handleKey(msg, fm)
	case tea.MouseMsg:
		return bw.handleMouse(msg, fm)
	}
	return nil
}

// handleKey provides standard keyboard routing:
//  1. Tab → fm.Next(), Shift-Tab → fm.Prev()
//  2. KeyBindingResolver: check registered global bindings
//  3. Route to focused leaf's Update (internal handler)
//  4. OnKeyPress fallthrough for unconsumed keys
func (bw *BaseWindow) handleKey(msg tea.KeyMsg, fm *input.FocusManager) tea.Cmd {
	// Tab/Shift-Tab: focus cycling
	if msg.Type == tea.KeyTab {
		return fm.Next()
	}
	if msg.Type == tea.KeyShiftTab {
		return fm.Prev()
	}

	// Key binding resolution
	if comp, action, ok := fm.ResolveKeyBinding(msg); ok {
		var cmds []tea.Cmd
		if leaf, isLeaf := comp.(widget.Leaf); isLeaf && comp.Focusable() {
			_, cmd := fm.ChangeFocusTo(leaf)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if action != nil {
			cmds = append(cmds, action())
		}
		return tea.Batch(cmds...)
	}

	// Route to focused leaf
	leaf := fm.FocusedLeaf()
	if leaf != nil {
		cmd, consumed := leaf.Update(msg)
		if consumed {
			return cmd
		}
	}

	// OnKeyPress fallthrough
	if bw.onKeyPress != nil {
		return bw.onKeyPress(msg)
	}
	return nil
}

// handleMouse provides standard mouse routing:
//  1. HitTest to find target component
//  2. Focus change if appropriate (press event on focusable component)
//  3. Dispatch MouseClickEvent on press
func (bw *BaseWindow) handleMouse(msg tea.MouseMsg, fm *input.FocusManager) tea.Cmd {
	target := fm.HitTest(msg.X, msg.Y)
	if target == nil {
		return nil
	}

	var cmds []tea.Cmd
	// Focus change on press
	if fm.IsFocusChangingEvent(msg) {
		if fm.CanFocusTo(target) {
			_, cmd := fm.ChangeFocusTo(target)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			return nil
		}
	}

	// Dispatch scroll events to the component under the cursor.
	isWheel := msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown ||
		msg.Button == tea.MouseButtonWheelLeft || msg.Button == tea.MouseButtonWheelRight
	if isWheel {
		dir := 1 // down
		if msg.Button == tea.MouseButtonWheelUp {
			dir = -1
		}
		cmd, _ := target.HandleEvent(widget.MouseScrollEvent{
			X: msg.X, Y: msg.Y, Direction: dir,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return tea.Batch(cmds...)
	}

	// Dispatch click events on button press (not motion, release).
	if msg.Action == tea.MouseActionPress {
		cmd, _ := target.HandleEvent(widget.MouseClickEvent{
			X: msg.X, Y: msg.Y, Button: msg.Button,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}
