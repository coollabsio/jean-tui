package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ThemeInfo contains metadata about a theme
type ThemeInfo struct {
	Name        string
	Description string
	Primary     string
	Secondary   string
}

// ThemeColors holds all colors for a theme
type ThemeColors struct {
	Primary        lipgloss.Color
	Secondary      lipgloss.Color
	Accent         lipgloss.Color
	Background     lipgloss.Color
	Surface        lipgloss.Color // For panels/cards
	Surface2       lipgloss.Color // For hover/elevated states
	Border         lipgloss.Color
	Foreground     lipgloss.Color
	Muted          lipgloss.Color
	Warning        lipgloss.Color
	Success        lipgloss.Color
	Error          lipgloss.Color
}

// Theme definitions
var themes = map[string]ThemeColors{
	"matrix": {
		Primary:    lipgloss.Color("#00FF41"),
		Secondary:  lipgloss.Color("#008F11"),
		Accent:     lipgloss.Color("#00FF41"),
		Background: lipgloss.Color("#000000"),
		Surface:    lipgloss.Color("#001a00"),
		Surface2:   lipgloss.Color("#003300"),
		Border:     lipgloss.Color("#00AA00"),
		Foreground: lipgloss.Color("#00FF41"),
		Muted:      lipgloss.Color("#00AA00"),
		Warning:    lipgloss.Color("#AAFF00"),
		Success:    lipgloss.Color("#00FF41"),
		Error:      lipgloss.Color("#FF0000"),
	},
	"coolify": {
		Primary:    lipgloss.Color("#9333EA"),
		Secondary:  lipgloss.Color("#7C3AED"),
		Accent:     lipgloss.Color("#A855F7"),
		Background: lipgloss.Color("#0a0a0a"),
		Surface:    lipgloss.Color("#1a1a1a"),
		Surface2:   lipgloss.Color("#2a2a2a"),
		Border:     lipgloss.Color("#3a3a3a"),
		Foreground: lipgloss.Color("#E5E5E5"),
		Muted:      lipgloss.Color("#9CA3AF"),
		Warning:    lipgloss.Color("#FFC107"),
		Success:    lipgloss.Color("#10B981"),
		Error:      lipgloss.Color("#EF4444"),
	},
	"dracula": {
		Primary:    lipgloss.Color("#FF79C6"),
		Secondary:  lipgloss.Color("#BD93F9"),
		Accent:     lipgloss.Color("#FF79C6"),
		Background: lipgloss.Color("#282A36"),
		Surface:    lipgloss.Color("#21222C"),
		Surface2:   lipgloss.Color("#44475A"),
		Border:     lipgloss.Color("#6272A4"),
		Foreground: lipgloss.Color("#F8F8F2"),
		Muted:      lipgloss.Color("#6272A4"),
		Warning:    lipgloss.Color("#F1FA8C"),
		Success:    lipgloss.Color("#50FA7B"),
		Error:      lipgloss.Color("#FF5555"),
	},
	"nord": {
		Primary:    lipgloss.Color("#88C0D0"),
		Secondary:  lipgloss.Color("#81A1C1"),
		Accent:     lipgloss.Color("#8FBCBB"),
		Background: lipgloss.Color("#2E3440"),
		Surface:    lipgloss.Color("#3B4252"),
		Surface2:   lipgloss.Color("#434C5E"),
		Border:     lipgloss.Color("#4C566A"),
		Foreground: lipgloss.Color("#ECEFF4"),
		Muted:      lipgloss.Color("#D08770"),
		Warning:    lipgloss.Color("#EBCB8B"),
		Success:    lipgloss.Color("#A3BE8C"),
		Error:      lipgloss.Color("#BF616A"),
	},
	"solarized": {
		Primary:    lipgloss.Color("#268BD2"),
		Secondary:  lipgloss.Color("#2AA198"),
		Accent:     lipgloss.Color("#D33682"),
		Background: lipgloss.Color("#002B36"),
		Surface:    lipgloss.Color("#073642"),
		Surface2:   lipgloss.Color("#586E75"),
		Border:     lipgloss.Color("#657B83"),
		Foreground: lipgloss.Color("#93A1A1"),
		Muted:      lipgloss.Color("#839496"),
		Warning:    lipgloss.Color("#B58900"),
		Success:    lipgloss.Color("#859900"),
		Error:      lipgloss.Color("#DC322F"),
	},
}

// themeMetadata contains information about each theme
var themeMetadata = map[string]ThemeInfo{
	"matrix": {
		Name:        "Matrix",
		Description: "Bright green on black - Cyberpunk style",
		Primary:     "#00FF41",
		Secondary:   "#008F11",
	},
	"coolify": {
		Name:        "Coolify",
		Description: "Purple accents, dark gray backgrounds - Modern minimalist",
		Primary:     "#9333EA",
		Secondary:   "#1a1a1a",
	},
	"dracula": {
		Name:        "Dracula",
		Description: "Pink and purple on dark gray - Popular dark theme",
		Primary:     "#FF79C6",
		Secondary:   "#282A36",
	},
	"nord": {
		Name:        "Nord",
		Description: "Arctic blue-green palette - Cool and professional",
		Primary:     "#88C0D0",
		Secondary:   "#2E3440",
	},
	"solarized": {
		Name:        "Solarized",
		Description: "Warm blue with yellow accents - Precision colors",
		Primary:     "#268BD2",
		Secondary:   "#002B36",
	},
}

// GetAvailableThemes returns a list of all available themes
func GetAvailableThemes() []ThemeInfo {
	themeList := []ThemeInfo{}
	themeNames := []string{"matrix", "coolify", "dracula", "nord", "solarized"}
	for _, name := range themeNames {
		if info, ok := themeMetadata[name]; ok {
			themeList = append(themeList, info)
		}
	}
	return themeList
}

// ApplyTheme applies a theme by name and rebuilds all styles
func ApplyTheme(themeName string) error {
	colors, ok := themes[themeName]
	if !ok {
		return fmt.Errorf("unknown theme: %s", themeName)
	}

	// Update module-level color variables
	primaryColor = colors.Primary
	secondaryColor = colors.Secondary
	accentColor = colors.Accent
	warningColor = colors.Warning
	successColor = colors.Success
	errorColor = colors.Error
	mutedColor = colors.Muted
	bgColor = colors.Background
	fgColor = colors.Foreground

	// Rebuild all styles with new colors
	rebuildStyles(colors)

	return nil
}

// rebuildStyles recreates all lipgloss styles with current colors
func rebuildStyles(colors ThemeColors) {
	// Base styles
	baseStyle = lipgloss.NewStyle().
		Foreground(colors.Foreground)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.Primary).
		Padding(1, 2)

	panelSeparatorStyle = lipgloss.NewStyle().
		Foreground(colors.Border)

	titleStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		Padding(0, 1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Primary).
		Bold(true).
		PaddingLeft(2).
		PaddingRight(2)

	normalItemStyle = lipgloss.NewStyle().
		Foreground(colors.Foreground).
		PaddingLeft(2)

	currentWorktreeStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		PaddingLeft(2)

	// Detail styles
	detailKeyStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true)

	detailValueStyle = lipgloss.NewStyle().
		Foreground(colors.Foreground)

	// Help/Status bar
	helpStyle = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)

	// Modal styles
	modalStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colors.Border).
		Padding(1, 2).
		Width(100)

	modalTitleStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		Align(lipgloss.Center)

	inputLabelStyle = lipgloss.NewStyle().
		Foreground(colors.Secondary).
		Bold(true)

	buttonStyle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Padding(0, 3).
		MarginRight(2)

	selectedButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Primary).
		Padding(0, 3).
		MarginRight(2).
		Bold(true)

	cancelButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Padding(0, 3)

	selectedCancelButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Primary).
		Padding(0, 3).
		Bold(true)

	// Delete button styles (red for danger)
	deleteButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Error).
		Padding(0, 3).
		MarginRight(2)

	selectedDeleteButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Error).
		Padding(0, 3).
		MarginRight(2).
		Bold(true)

	disabledButtonStyle = lipgloss.NewStyle().
		Foreground(colors.Muted).
		Padding(0, 3).
		MarginRight(2)

	// Notification styles
	successNotifStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Success).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(colors.Success)

	errorNotifStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Error).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Border(lipgloss.DoubleBorder(), true, true, true, true).
		BorderForeground(colors.Error)

	warningNotifStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Warning).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(colors.Warning)

	infoNotifStyle = lipgloss.NewStyle().
		Foreground(colors.Background).
		Background(colors.Accent).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(colors.Accent)
}
