package tui

import "github.com/charmbracelet/lipgloss"

// Theme color variables - initialized with Coolify theme, updated by ApplyTheme()
var (
	// Colors - initialized with Coolify theme
	primaryColor   = lipgloss.Color("#9333EA") // Coolify purple
	secondaryColor = lipgloss.Color("#7C3AED") // Coolify lighter purple
	accentColor    = lipgloss.Color("#A855F7") // Coolify accent purple
	warningColor   = lipgloss.Color("#FFC107") // Coolify amber warning
	successColor   = lipgloss.Color("#10B981") // Coolify green success
	errorColor     = lipgloss.Color("#EF4444") // Coolify red error
	mutedColor     = lipgloss.Color("#9CA3AF") // Coolify gray muted
	bgColor = lipgloss.Color("#0a0a0a") // Coolify near-black background
	fgColor = lipgloss.Color("#E5E5E5") // Coolify light gray text
)

// Style variables - mutable, rebuilt by ApplyTheme()
var (
	// Base styles
	baseStyle lipgloss.Style

	// Panel styles
	panelStyle       lipgloss.Style
	activePanelStyle lipgloss.Style
	panelSeparatorStyle lipgloss.Style

	titleStyle lipgloss.Style

	// List item styles
	selectedItemStyle    lipgloss.Style
	normalItemStyle      lipgloss.Style
	currentWorktreeStyle lipgloss.Style

	// Detail styles
	detailKeyStyle   lipgloss.Style
	detailValueStyle lipgloss.Style

	// Help/Status bar
	helpStyle   lipgloss.Style
	statusStyle lipgloss.Style
	errorStyle  lipgloss.Style

	// Modal styles
	modalStyle       lipgloss.Style
	modalTitleStyle  lipgloss.Style
	inputLabelStyle  lipgloss.Style

	buttonStyle              lipgloss.Style
	selectedButtonStyle      lipgloss.Style
	cancelButtonStyle        lipgloss.Style
	selectedCancelButtonStyle lipgloss.Style

	// Delete button styles (red for danger)
	deleteButtonStyle         lipgloss.Style
	selectedDeleteButtonStyle  lipgloss.Style
	disabledButtonStyle        lipgloss.Style

	// Notification styles
	successNotifStyle lipgloss.Style
	errorNotifStyle   lipgloss.Style
	warningNotifStyle lipgloss.Style
	infoNotifStyle    lipgloss.Style
)

// InitStyles initializes styles with the Coolify theme on startup
func InitStyles() {
	// Apply the default Coolify theme
	ApplyTheme("coolify")
}
