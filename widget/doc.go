// Package widget is the core of the Mate component framework: it defines
// the Component interface, the base types (BaseComponent, FocusableComponent,
// BaseContainer), the layout primitives (Panel, Field), and the built-in
// widgets — Button, TextInput, FormattedTextInput, Toggle, CheckboxList,
// Table, ScrollableText, MarkdownTextArea, TabBar, TabComponent, Card,
// Text.
//
// Components form a tree. Containers hold children and lay them out
// (Vertical, Horizontal, or TCB — Top/Center/Bottom). Leaves render
// terminal text and receive keyboard input when focused. Every component
// has an ID, a size, a parent back-reference, and is rendered through
// View().
//
// The intended usage is to compose a tree, attach callbacks (OnPress,
// OnChange, OnSubmit, ...), and let the framework do focus management,
// event routing, and rendering. There are no custom Update or View
// methods to write.
//
// See package github.com/muralx/mate/window for how a widget tree binds
// to a terminal screen, and package github.com/muralx/mate/input for
// the focus and key-binding machinery that runs in the background.
package widget
