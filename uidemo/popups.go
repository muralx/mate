package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
	"github.com/muralx/mate/window"
)

func (s *demoState) buildConfirmPopup(message string, onConfirm func() tea.Cmd) *window.PopupWindow {
	popup := window.NewPopupWindow("confirm-popup", "Confirm", window.DefaultPopupStyles())

	msg := widget.NewText("confirm-msg", message, lipgloss.NewStyle().Foreground(colorText).Width(40).Align(lipgloss.Center))

	yesBtn := widget.NewButton("confirm-yes", "Yes", themeSuccessButtonStyles())
	noBtn := widget.NewButton("confirm-no", "No", themeDangerButtonStyles())

	yesBtn.OnPress(func() tea.Cmd {
		return popup.Close(true)
	})
	noBtn.OnPress(func() tea.Cmd {
		return popup.Close(nil)
	})

	popup.OnResult(func(result any) tea.Cmd {
		if confirmed, ok := result.(bool); ok && confirmed {
			return onConfirm()
		}
		s.setStatus("Action cancelled")
		return nil
	})

	btnRow := widget.NewPanel("confirm-btn-row", widget.Horizontal)
	btnRow.SetSpacing(2)
	btnRow.Add(yesBtn, widget.Next)
	btnRow.Add(noBtn, widget.Next)

	popup.Add(msg, widget.Next)
	popup.Add(btnRow, widget.Next)
	return popup
}

func (s *demoState) buildAddServerPopup() *window.PopupWindow {
	popup := window.NewPopupWindow("add-server-popup", "Add Server", window.DefaultPopupStyles())

	panel := widget.NewPanel("add-server-panel")
	panel.SetBorder(themeBorder())
	panel.SetTitle("New Server")
	panel.SetPreferredWidth(50)
	panel.SetPreferredHeight(12)

	nameInput := widget.NewTextInput("server-name", 30)
	nameInput.WithPlaceholder("e.g. production-web-01")
	panel.Add(widget.NewField("server-name-field", "Name", nameInput, themeFieldStyles()), widget.Next)

	hostInput := widget.NewTextInput("server-host", 30)
	hostInput.WithPlaceholder("e.g. 10.0.1.50")
	panel.Add(widget.NewField("server-host-field", "Host", hostInput, themeFieldStyles()), widget.Next)

	portInput := widget.NewFormattedTextInput("server-port", 8)
	portInput.WithPlaceholder("22")
	portInput.WithCharLimit(5)
	portInput.WithValidation(func(val string) error {
		if val == "" {
			return nil
		}
		var port int
		_, err := fmt.Sscanf(val, "%d", &port)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("must be 1-65535")
		}
		return nil
	})
	panel.Add(widget.NewField("server-port-field", "Port", portInput, themeFieldStyles()), widget.Next)

	saveBtn := widget.NewButton("save-server", "Save", themeSuccessButtonStyles())
	cancelBtn := widget.NewButton("cancel-server", "Cancel", themeDangerButtonStyles())

	saveBtn.OnPress(func() tea.Cmd {
		name := nameInput.Value()
		host := hostInput.Value()
		port := portInput.Value()
		if name == "" || host == "" {
			return nil
		}
		if port == "" {
			port = "22"
		}
		return popup.Close(map[string]string{
			"name": name, "host": host, "port": port,
		})
	})
	cancelBtn.OnPress(func() tea.Cmd {
		return popup.Close(nil)
	})

	btnRow := widget.NewPanel("add-server-btn-row", widget.Horizontal)
	btnRow.SetSpacing(2)
	btnRow.Add(saveBtn, widget.Next)
	btnRow.Add(cancelBtn, widget.Next)
	panel.Add(btnRow, widget.Next)

	popup.Add(panel, widget.Next)

	popup.OnResult(func(result any) tea.Cmd {
		if data, ok := result.(map[string]string); ok {
			s.addServerRow(data["name"], data["host"], data["port"])
			s.setStatus(fmt.Sprintf("Added server: %s (%s:%s)", data["name"], data["host"], data["port"]))
		} else {
			s.setStatus("Popup cancelled")
		}
		return nil
	})

	return popup
}
