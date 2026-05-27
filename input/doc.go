// Package input handles focus management and key-binding resolution
// across a widget component tree.
//
// FocusManager walks the tree to find focusable leaves, implements
// Tab and Shift-Tab cycling, click-to-focus via hit testing, and
// ID-based focus (FocusByID). KeyBindingResolver walks the same tree
// to resolve global key bindings registered on any component.
//
// Apps don't usually instantiate these directly — package
// github.com/muralx/mate/window wires them in when you call NewApp.
package input
