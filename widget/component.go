package widget

import (
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Alignment controls horizontal alignment within a component's width.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Event is a high-level UI event dispatched to components.
type Event interface{ isEvent() }

// MouseClickEvent is sent when a mouse button is pressed on a component.
type MouseClickEvent struct {
	X, Y   int
	Button tea.MouseButton
}

func (MouseClickEvent) isEvent() {}

// MouseScrollEvent is sent when the mouse wheel is scrolled over a component.
// Direction: negative = up, positive = down.
type MouseScrollEvent struct {
	X, Y      int
	Direction int
}

func (MouseScrollEvent) isEvent() {}

// Component is the base interface for all UI elements.
type Component interface {
	ID() string
	View() string
	SetSize(w, h int)
	Size() (w, h int)
	SetPosition(x, y int)
	Position() (x, y int)
	Visible() bool
	SetVisible(bool)
	Enabled() bool
	SetEnabled(bool)
	Active() bool
	Focusable() bool
	Focused() bool
	SetFocused(bool) tea.Cmd
	Parent() Container
	SetParent(Container)
	KeyBindings() []key.Binding
	RegisterKeyBinding(keys, description string, action func() tea.Cmd)
	ResolveKeyBinding(tea.KeyMsg) (func() tea.Cmd, bool)
	// HandleEvent receives high-level UI events. If not consumed (returns false),
	// the base implementation bubbles the event up to the parent.
	HandleEvent(event Event) (tea.Cmd, bool)
	PreferredWidth() int
	PreferredHeight() int
	SetPreferredWidth(int)
	SetPreferredHeight(int)
}

// Leaf is a focusable component that handles keyboard input.
// Only focused leaves receive keyboard events. Mouse and other events
// go through HandleEvent on Component.
type Leaf interface {
	Component
	Update(msg tea.KeyMsg) (tea.Cmd, bool)
}

// Container is a component composed of other components.
type Container interface {
	Component
	Children() []Component
	AddChild(child Component)
	InnerFocused() bool
}

// registeredBinding pairs a key binding with its action callback.
type registeredBinding struct {
	binding key.Binding
	action  func() tea.Cmd
}

// BaseComponent provides common fields and methods embedded by all concrete components.
type BaseComponent struct {
	id                 string
	width              int
	height             int
	preferredWidth     int
	preferredHeight    int
	x, y               int
	visible            bool
	enabled            bool
	focused            bool
	alignment          Alignment
	parent             Container
	registeredBindings []registeredBinding
}

// NewBaseComponent creates a BaseComponent with sensible defaults.
func NewBaseComponent(id string) *BaseComponent {
	return &BaseComponent{
		id:      id,
		visible: true,
		enabled: true,
	}
}

func (bc *BaseComponent) ID() string                { return bc.id }
func (bc *BaseComponent) View() string              { return "" }
func (bc *BaseComponent) SetSize(w, h int)          { bc.width = w; bc.height = h }
func (bc *BaseComponent) Size() (int, int)          { return bc.width, bc.height }
func (bc *BaseComponent) PreferredWidth() int       { return bc.preferredWidth }
func (bc *BaseComponent) PreferredHeight() int      { return bc.preferredHeight }
func (bc *BaseComponent) SetPreferredWidth(w int)   { bc.preferredWidth = w }
func (bc *BaseComponent) SetPreferredHeight(h int)  { bc.preferredHeight = h }
func (bc *BaseComponent) SetPosition(x, y int)      { bc.x = x; bc.y = y }
func (bc *BaseComponent) Position() (int, int)      { return bc.x, bc.y }
func (bc *BaseComponent) Visible() bool             { return bc.visible }
func (bc *BaseComponent) SetVisible(v bool)         { bc.visible = v }
func (bc *BaseComponent) Enabled() bool             { return bc.enabled }
func (bc *BaseComponent) SetEnabled(v bool)         { bc.enabled = v }
func (bc *BaseComponent) Focused() bool             { return bc.focused }
func (bc *BaseComponent) SetFocused(v bool) tea.Cmd { bc.focused = v; return nil }
func (bc *BaseComponent) Parent() Container         { return bc.parent }
func (bc *BaseComponent) SetParent(p Container)     { bc.parent = p }
func (bc *BaseComponent) Focusable() bool           { return false }

// RegisterKeyBinding registers a global key binding with its action on this component.
// keys is the key combo (e.g. "ctrl+q"). description is the help text (e.g. "Quit").
// Pass empty description to hide from help/hints.
func (bc *BaseComponent) RegisterKeyBinding(keys, description string, action func() tea.Cmd) {
	binding := key.NewBinding(key.WithKeys(keys))
	if description != "" {
		binding = key.NewBinding(key.WithKeys(keys), key.WithHelp(keys, description))
	}
	bc.registeredBindings = append(bc.registeredBindings, registeredBinding{binding: binding, action: action})
}

// RemoveKeyBinding removes a previously registered key binding by matching its key(s).
// No-op if the binding is not found.
func (bc *BaseComponent) RemoveKeyBinding(binding key.Binding) {
	keys := binding.Keys()
	for i, rb := range bc.registeredBindings {
		if slices.Equal(rb.binding.Keys(), keys) {
			bc.registeredBindings = slices.Delete(bc.registeredBindings, i, i+1)
			return
		}
	}
}

// ResolveKeyBinding checks if the key message matches any registered binding.
// Returns the action and true if a match is found.
func (bc *BaseComponent) ResolveKeyBinding(msg tea.KeyMsg) (func() tea.Cmd, bool) {
	for _, rb := range bc.registeredBindings {
		if key.Matches(msg, rb.binding) {
			return rb.action, true
		}
	}
	return nil, false
}

// FocusableComponent extends BaseComponent for components that can receive focus.
// Adds OnKeyPress handler and overrides Focusable() to return true.
// Leaf components embed this instead of BaseComponent.
type FocusableComponent struct {
	BaseComponent
	onKeyPress func(tea.KeyMsg) tea.Cmd
}

// NewFocusableComponent creates a FocusableComponent with sensible defaults.
func NewFocusableComponent(id string) FocusableComponent {
	return FocusableComponent{
		BaseComponent: *NewBaseComponent(id),
	}
}

func (fc *FocusableComponent) Focusable() bool { return true }

// OnKeyPress sets a handler called when the component has focus and receives
// a key that the component's internal Update doesn't consume.
func (fc *FocusableComponent) OnKeyPress(fn func(tea.KeyMsg) tea.Cmd) {
	fc.onKeyPress = fn
}

// HandleEvent default: not consumed, bubble to parent.
func (bc *BaseComponent) HandleEvent(event Event) (tea.Cmd, bool) {
	if bc.parent != nil {
		return bc.parent.HandleEvent(event)
	}
	return nil, false
}
func (bc *BaseComponent) KeyBindings() []key.Binding {
	bindings := make([]key.Binding, 0, len(bc.registeredBindings))
	for _, rb := range bc.registeredBindings {
		bindings = append(bindings, rb.binding)
	}
	if len(bindings) == 0 {
		return nil
	}
	return bindings
}

// Active returns true if this component is enabled and its parent (if any) is also active.
func (bc *BaseComponent) Active() bool {
	return bc.enabled && (bc.parent == nil || bc.parent.Active())
}

// SetAlignment sets the horizontal alignment used by RenderInSize.
func (bc *BaseComponent) SetAlignment(a Alignment) { bc.alignment = a }

// Alignment returns the current horizontal alignment.
func (bc *BaseComponent) Alignment() Alignment { return bc.alignment }

// RenderInSize pads content to fill the component's stored width and height
// using its alignment. MaxWidth prevents wrapping — content is truncated if
// too wide. Components must render at their allocated size.
func (bc *BaseComponent) RenderInSize(content string) string {
	if bc.width <= 0 && bc.height <= 0 {
		return content
	}
	s := lipgloss.NewStyle().Align(bc.alignmentToLipgloss())
	if bc.width > 0 {
		s = s.Width(bc.width).MaxWidth(bc.width)
	}
	if bc.height > 0 {
		s = s.Height(bc.height)
	}
	return s.Render(content)
}

func (bc *BaseComponent) alignmentToLipgloss() lipgloss.Position {
	switch bc.alignment {
	case AlignCenter:
		return lipgloss.Center
	case AlignRight:
		return lipgloss.Right
	default:
		return lipgloss.Left
	}
}

// BaseContainer is a component composed of other components.
type BaseContainer struct {
	BaseComponent
	children []Component
	self     Container // pointer to the outer concrete type
}

// NewBaseContainer creates a BaseContainer with sensible defaults.
// If self is non-nil it is used as the parent reference for AddChild;
// otherwise the BaseContainer itself is used (useful in tests).
func NewBaseContainer(id string, self Container) *BaseContainer {
	bc := &BaseContainer{}
	bc.id = id
	bc.visible = true
	bc.enabled = true
	if self != nil {
		bc.self = self
	} else {
		bc.self = bc
	}
	return bc
}

// AddChild appends a child component and sets its parent to bc.self.
func (bc *BaseContainer) AddChild(child Component) {
	bc.children = append(bc.children, child)
	child.SetParent(bc.self)
}

// Children returns the child components.
func (bc *BaseContainer) Children() []Component {
	return bc.children
}

// Focusable returns false; containers do not receive focus directly.
func (bc *BaseContainer) Focusable() bool {
	return false
}

// View returns an empty string; concrete containers override with their layout.
func (bc *BaseContainer) View() string {
	return ""
}

// InnerFocused returns true if any descendant has focus.
func (bc *BaseContainer) InnerFocused() bool {
	for _, child := range bc.children {
		if child.Focused() {
			return true
		}
		if c, ok := child.(Container); ok {
			if c.InnerFocused() {
				return true
			}
		}
	}
	return false
}
