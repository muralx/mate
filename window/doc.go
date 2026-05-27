// Package window manages the terminal screen and popup stack on top of
// the widget tree.
//
// NewWindow creates the main screen. NewPopupWindow creates an overlay
// that pushes onto the popup stack with its own focus scope and closes
// via Close(result) and an OnResult callback. NewApp adapts a Window
// to the Bubble Tea Model interface, so terminal I/O and the event
// loop flow through Bubble Tea unchanged — Mate only owns the
// component-tree layer above it.
//
// See package github.com/muralx/mate/widget for the components that
// populate a window.
package window
