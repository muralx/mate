package window

import "github.com/muralx/mate/widget"

// MainWindow is the main application window. It embeds BaseWindow for
// container behavior and event routing. Use NewApp(win) to start
// the application.
type MainWindow struct {
	BaseWindow
}

// NewWindow creates a new main window with the given ID.
// Optional layout parameter defaults to TCB.
func NewWindow(id string, layout ...widget.Layout) *MainWindow {
	w := &MainWindow{}
	w.BaseWindow = newBaseWindow(id, w, layout...)
	return w
}
