package input

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muralx/mate/widget"
)

// KeyBindingResolver resolves global key bindings by walking the component tree.
// It checks registered key bindings from ALL active components before focus-based routing.
type KeyBindingResolver struct {
	root widget.Container
}

// NewKeyBindingResolver creates a resolver rooted at the given container.
func NewKeyBindingResolver(root widget.Container) *KeyBindingResolver {
	return &KeyBindingResolver{root: root}
}

// SetRoot updates the root container (e.g., when window changes).
func (r *KeyBindingResolver) SetRoot(root widget.Container) {
	r.root = root
}

// Resolve checks if the key matches any active component's registered key binding.
// Children are checked before their parent (more-specific wins).
// Returns the matching component, the bound action, and true if a match is found.
func (r *KeyBindingResolver) Resolve(msg tea.KeyMsg) (widget.Component, func() tea.Cmd, bool) {
	var match widget.Component
	var action func() tea.Cmd
	r.walk(r.root, msg, &match, &action)
	return match, action, match != nil
}

func (r *KeyBindingResolver) walk(c widget.Container, msg tea.KeyMsg, match *widget.Component, action *func() tea.Cmd) {
	if *match != nil {
		return
	}
	for _, child := range c.Children() {
		if !child.Visible() || !child.Active() {
			continue
		}
		// Recurse into child containers first (deeper = more specific)
		if container, ok := child.(widget.Container); ok {
			r.walk(container, msg, match, action)
			if *match != nil {
				return
			}
		}
		// Then check this child's own bindings
		if a, ok := child.ResolveKeyBinding(msg); ok {
			*match = child
			*action = a
			return
		}
	}
	// Finally check this container's own bindings
	if c.Visible() && c.Active() {
		if a, ok := c.ResolveKeyBinding(msg); ok {
			*match = c
			*action = a
		}
	}
}
