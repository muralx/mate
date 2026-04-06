package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muralx/mate/widget"
)

// Color palette
var (
	colorPrimary   = lipgloss.Color("#4fc3f7")
	colorSecondary = lipgloss.Color("#81c784")
	colorWarning   = lipgloss.Color("#ffb74d")
	colorDanger    = lipgloss.Color("#ef5350")
	colorText      = lipgloss.Color("#e0e0e0")
	colorDim       = lipgloss.Color("#888888")
	colorBg        = lipgloss.Color("#1e1e2e")
	colorHighlight = lipgloss.Color("#ffeb3b")
	colorAccent    = lipgloss.Color("#ce93d8")
)

func themeBorder() widget.BorderConfig {
	return widget.SingleLineBorder(string(colorDim), string(colorPrimary))
}

func themeFieldStyles() widget.FieldStyles {
	return widget.FieldStyles{
		Label:     lipgloss.NewStyle().Foreground(colorDim),
		LabelHot:  lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),
		Separator: lipgloss.NewStyle().Foreground(colorDim),
	}
}

func themeButtonStyles() widget.ButtonStyles {
	return widget.ButtonStyles{
		Normal:  lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(colorPrimary).Bold(true),
	}
}

func themeDangerButtonStyles() widget.ButtonStyles {
	return widget.ButtonStyles{
		Normal:  lipgloss.NewStyle().Foreground(colorDanger).Bold(true),
		Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(colorDanger).Bold(true),
	}
}

func themeSuccessButtonStyles() widget.ButtonStyles {
	return widget.ButtonStyles{
		Normal:  lipgloss.NewStyle().Foreground(colorSecondary).Bold(true),
		Focused: lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(colorSecondary).Bold(true),
	}
}

func themeCardStyles() widget.CardStyles {
	return widget.CardStyles{
		Border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorDim).Padding(0, 1),
		Title:  lipgloss.NewStyle().Foreground(colorDim),
		Value:  lipgloss.NewStyle().Foreground(colorText).Bold(true),
		Alert:  lipgloss.NewStyle().Foreground(colorDanger).Bold(true),
	}
}

func themeTableStyles() widget.TableStyles {
	return widget.TableStyles{
		Header:   lipgloss.NewStyle().Foreground(colorPrimary).Bold(true),
		Selected: lipgloss.NewStyle().Background(lipgloss.Color("#333355")),
		Cell:     lipgloss.NewStyle().Foreground(colorText),
	}
}

func themeTabBarStyles() widget.TabBarStyles {
	return widget.TabBarStyles{
		Active:   lipgloss.NewStyle().Foreground(lipgloss.Color("#000")).Background(colorPrimary).Bold(true).Padding(0, 2),
		Inactive: lipgloss.NewStyle().Foreground(colorDim).Padding(0, 2),
		Focused:  lipgloss.NewStyle().Foreground(colorHighlight).Padding(0, 2),
	}
}

func themeToggleStyles() widget.ToggleStyles {
	return widget.ToggleStyles{
		Label:       lipgloss.NewStyle().Foreground(colorDim),
		OnActive:    lipgloss.NewStyle().Foreground(colorSecondary).Bold(true),
		OnFocused:   lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),
		OffActive:   lipgloss.NewStyle().Foreground(colorWarning),
		OffFocused:  lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),
		OffInactive: lipgloss.NewStyle().Foreground(colorDim),
	}
}

func themeCheckboxListStyles() widget.CheckboxListStyles {
	return widget.CheckboxListStyles{
		Cursor:    lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),
		Checked:   lipgloss.NewStyle().Foreground(colorSecondary),
		Unchecked: lipgloss.NewStyle().Foreground(colorDim),
		Item:      lipgloss.NewStyle().Foreground(colorText),
		Group:     lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		Dim:       lipgloss.NewStyle().Foreground(colorDim),
	}
}
