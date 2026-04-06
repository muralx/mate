package input

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muralx/mate/widget"
)

// FocusManager walks the component tree to find focusable+active leaves
// and manages focus cycling (Tab/Shift-Tab), click-to-focus, and ID-based focus.
type FocusManager struct {
	root widget.Container
}

// NewFocusManager creates a FocusManager rooted at the given container.
func NewFocusManager(root widget.Container) *FocusManager {
	return &FocusManager{root: root}
}

// SetRoot re-roots the manager to a new container tree.
func (fm *FocusManager) SetRoot(root widget.Container) {
	fm.root = root
}

// Leaves returns all focusable+active leaves in tree order (depth-first walk).
func (fm *FocusManager) Leaves() []widget.Leaf {
	var leaves []widget.Leaf
	fm.walk(fm.root, &leaves)
	return leaves
}

func (fm *FocusManager) walk(c widget.Container, leaves *[]widget.Leaf) {
	for _, child := range c.Children() {
		if !child.Visible() || !child.Active() {
			continue
		}
		if leaf, ok := child.(widget.Leaf); ok && child.Focusable() {
			*leaves = append(*leaves, leaf)
		}
		if container, ok := child.(widget.Container); ok {
			fm.walk(container, leaves)
		}
	}
}

// FocusedLeaf returns the currently focused leaf, or nil.
func (fm *FocusManager) FocusedLeaf() widget.Leaf {
	for _, leaf := range fm.Leaves() {
		if leaf.Focused() {
			return leaf
		}
	}
	return nil
}

// Next moves focus to the next leaf (Tab). Wraps around. Uses ChangeFocusTo.
func (fm *FocusManager) Next() tea.Cmd {
	leaves := fm.Leaves()
	if len(leaves) == 0 {
		return nil
	}
	current := fm.focusedIndex(leaves)
	next := (current + 1) % len(leaves)
	_, cmd := fm.ChangeFocusTo(leaves[next])
	return cmd
}

// Prev moves focus to the previous leaf (Shift-Tab). Wraps around. Uses ChangeFocusTo.
func (fm *FocusManager) Prev() tea.Cmd {
	leaves := fm.Leaves()
	if len(leaves) == 0 {
		return nil
	}
	current := fm.focusedIndex(leaves)
	prev := (current - 1 + len(leaves)) % len(leaves)
	_, cmd := fm.ChangeFocusTo(leaves[prev])
	return cmd
}

// FocusByID focuses a specific leaf by ID. Uses ChangeFocusTo.
func (fm *FocusManager) FocusByID(id string) (bool, tea.Cmd) {
	for _, leaf := range fm.Leaves() {
		if leaf.ID() == id {
			return fm.ChangeFocusTo(leaf)
		}
	}
	return false, nil
}

// FocusFirst focuses the first available leaf. Uses ChangeFocusTo.
func (fm *FocusManager) FocusFirst() tea.Cmd {
	leaves := fm.Leaves()
	if len(leaves) > 0 {
		_, cmd := fm.ChangeFocusTo(leaves[0])
		return cmd
	}
	return nil
}

// HitTest finds the leaf at screen coordinates. Pure lookup, no side effects.
func (fm *FocusManager) HitTest(x, y int) widget.Leaf {
	leaves := fm.Leaves()
	for i := len(leaves) - 1; i >= 0; i-- {
		leaf := leaves[i]
		px, py := leaf.Position()
		w, h := leaf.Size()
		if x >= px && x < px+w && y >= py && y < py+h {
			return leaf
		}
	}
	return nil
}

// IsFocusChangingEvent returns true if the mouse event should trigger a focus change.
func (fm *FocusManager) IsFocusChangingEvent(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
}

// CanFocusTo returns true if the leaf can receive focus.
func (fm *FocusManager) CanFocusTo(leaf widget.Leaf) bool {
	return leaf != nil && leaf.Focusable() && leaf.Active()
}

// ChangeFocusTo blurs the current focused leaf and focuses the new one.
// This is the shared primitive used by Next(), Prev(), FocusByID(), and mouse press.
func (fm *FocusManager) ChangeFocusTo(leaf widget.Leaf) (bool, tea.Cmd) {
	if !fm.CanFocusTo(leaf) {
		return false, nil
	}
	var cmds []tea.Cmd
	if current := fm.FocusedLeaf(); current != nil {
		if cmd := current.SetFocused(false); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if cmd := leaf.SetFocused(true); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return true, tea.Batch(cmds...)
}

// ResolveKeyBinding checks if the key matches any active component's registered binding.
// Delegates to a KeyBindingResolver rooted at the same root.
func (fm *FocusManager) ResolveKeyBinding(msg tea.KeyMsg) (widget.Component, func() tea.Cmd, bool) {
	resolver := NewKeyBindingResolver(fm.root)
	return resolver.Resolve(msg)
}

// AllActiveKeyBindings returns registered key bindings from visible, active
// components in tree-walk order. Duplicate key combinations are deduplicated
// (first-match wins, matching KeyBindingResolver.Resolve behavior).
// This enables building status bars, help screens, or shortcut overlays
// without reimplementing the component tree walk.
func (fm *FocusManager) AllActiveKeyBindings() []key.Binding {
	var bindings []key.Binding
	seen := map[string]bool{}
	fm.walkBindings(fm.root, &bindings, seen)
	return bindings
}

func (fm *FocusManager) walkBindings(c widget.Component, out *[]key.Binding, seen map[string]bool) {
	if !c.Visible() || !c.Active() {
		return
	}
	for _, b := range c.KeyBindings() {
		k := b.Help().Key
		if !seen[k] {
			seen[k] = true
			*out = append(*out, b)
		}
	}
	if container, ok := c.(widget.Container); ok {
		for _, child := range container.Children() {
			fm.walkBindings(child, out, seen)
		}
	}
}

// FocusedKeyBindings returns the key bindings from the focused leaf.
func (fm *FocusManager) FocusedKeyBindings() []key.Binding {
	leaf := fm.FocusedLeaf()
	if leaf == nil {
		return nil
	}
	return leaf.KeyBindings()
}

func (fm *FocusManager) focusedIndex(leaves []widget.Leaf) int {
	for i, leaf := range leaves {
		if leaf.Focused() {
			return i
		}
	}
	return -1
}
