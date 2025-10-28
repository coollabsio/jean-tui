package tui

import "github.com/charmbracelet/lipgloss"

// TODO: Make theming system
// Future implementation should allow users to:
// 1. Choose from predefined themes (Matrix, Dracula, Nord, etc.)
// 2. Define custom themes via config file
// 3. Store theme preference in ~/.config/gcool/config.json
// 4. Load theme on startup and apply dynamically
//
// Theme structure should include:
// - Primary, secondary, accent colors
// - Background and foreground colors
// - Error, warning, and success colors
// - Muted/dimmed text colors
//
// Implementation approach:
// - Add Theme struct in config package
// - Add ApplyTheme() function to load colors
// - Add theme selection modal to settings (press 's' -> 't')
// - Support both built-in themes and custom color definitions

var (
	// Colors - Matrix Theme (black with green)
	primaryColor   = lipgloss.Color("#00FF41") // Bright Matrix green
	secondaryColor = lipgloss.Color("#008F11") // Medium green
	accentColor    = lipgloss.Color("#00FF41") // Bright Matrix green for highlights
	warningColor   = lipgloss.Color("#AAFF00") // Yellow-green for warnings
	successColor   = lipgloss.Color("#00FF41") // Bright green for success
	errorColor     = lipgloss.Color("#FF0000") // Red for errors
	mutedColor     = lipgloss.Color("#003B00") // Dark green for muted text
	bgColor        = lipgloss.Color("#000000") // Pure black background
	fgColor        = lipgloss.Color("#00FF41") // Bright green text

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(fgColor)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF41")).
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AA00")).
			PaddingLeft(2)

	currentWorktreeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF41")).
				Bold(true).
				PaddingLeft(2)

	// Detail styles
	detailKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(fgColor)

	// Help/Status bar
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#008800")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF41")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// Modal styles
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Width(60)

	modalTitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Align(lipgloss.Center)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AA00")).
			Padding(0, 3).
			MarginRight(2)

	selectedButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF41")).
				Padding(0, 3).
				MarginRight(2).
				Bold(true)

	cancelButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00AA00")).
				Padding(0, 3)

	selectedCancelButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(lipgloss.Color("#FF0000")).
					Padding(0, 3).
					Bold(true)

	// Delete button styles (red for danger)
	deleteButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AA0000")).
				Padding(0, 3).
				MarginRight(2)

	selectedDeleteButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(lipgloss.Color("#FF0000")).
					Padding(0, 3).
					MarginRight(2).
					Bold(true)

	disabledButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333")).
				Padding(0, 3).
				MarginRight(2)
)
