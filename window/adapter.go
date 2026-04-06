package window

import tea "github.com/charmbracelet/bubbletea"

// App bridges an MTUI Window to Bubble Tea's Model interface.
type App struct {
	stack  *Stack
	win    *MainWindow
	width  int
	height int
}

// NewApp creates a Bubble Tea Model that manages the given Window.
func NewApp(win *MainWindow) *App {
	stack := newStack(win)
	return &App{stack: stack, win: win}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.width = sizeMsg.Width
		a.height = sizeMsg.Height
		return a, nil
	}
	cmd := a.stack.update(msg)
	return a, cmd
}

// View implements tea.Model.
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return ""
	}
	return a.stack.view(a.width, a.height)
}
