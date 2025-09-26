package auth

import (
	"github.com/charmbracelet/lipgloss"
)

// Authentication styles
var (
	// AuthRequiredStyle - Style for authentication required messages
	AuthRequiredStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")) // Orange

	// AuthSuccessStyle - Style for successful authentication
	AuthSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("46")) // Green

	// AuthFailedStyle - Style for failed authentication
	AuthFailedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")) // Red

	// AuthInProgressStyle - Style for authentication in progress
	AuthInProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange

	// AuthPromptStyle - Style for authentication prompts
	AuthPromptStyle = lipgloss.NewStyle().
			Bold(true)

	// AuthInfoStyle - Style for authentication information
	AuthInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")) // Gray

	// LockedCheckStyle - Style for locked/restricted checks
	LockedCheckStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange

	// AccessIndicatorStyle - Style for access indicators
	AccessIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange
)

// Authentication messages
const (
	// AuthRequiredMessage - Message when authentication is required
	AuthRequiredMessage = "🔐 Admin Authentication Required"

	// AuthSuccessMessage - Message when authentication is successful
	AuthSuccessMessage = "✅ Authentication Successful"

	// AuthFailedMessage - Message when authentication fails
	AuthFailedMessage = "❌ Authentication Failed"

	// AuthInProgressMessage - Message when authentication is in progress
	AuthInProgressMessage = "⏳ Authenticating..."

	// AuthPromptMessage - Default authentication prompt
	AuthPromptMessage = "Enter password for admin access:"

	// AuthInstructions - Authentication instructions
	AuthInstructions = "Press Enter to authenticate, Esc to cancel"

	// AuthRetryMessage - Message for retrying authentication
	AuthRetryMessage = "Press 'a' to retry authentication"

	// AdminAccessGranted - Message when admin access is granted
	AdminAccessGranted = "✅ Admin access granted"

	// AdminAccessRequired - Message when admin access is required
	AdminAccessRequired = "🔒 Some checks require admin privileges. Press 'a' to authenticate"

	// SudoNotAvailable - Message when sudo is not available
	SudoNotAvailable = "Sudo is not available. Please run as root."

	// InvalidPasswordMessage - Message for invalid password
	InvalidPasswordMessage = "Invalid password or insufficient privileges"

	// AuthenticationFailedGeneric - Generic authentication failure message
	AuthenticationFailedGeneric = "Authentication failed"

	// LockedCheckIndicator - Indicator for locked checks
	LockedCheckIndicator = "🔒"
)

// GetAuthMessage returns the appropriate authentication message based on state
func GetAuthMessage(state int, customMessage string) string {
	switch state {
	case 1: // AuthRequired
		if customMessage != "" {
			return customMessage
		}
		return AuthRequiredMessage
	case 2: // AuthInProgress
		return AuthInProgressMessage
	case 3: // AuthSuccess
		return AuthSuccessMessage
	case 4: // AuthFailed
		if customMessage != "" {
			return customMessage
		}
		return AuthFailedMessage
	default:
		return ""
	}
}

// GetAuthStyle returns the appropriate style based on authentication state
func GetAuthStyle(state int) lipgloss.Style {
	switch state {
	case 1: // AuthRequired
		return AuthRequiredStyle
	case 2: // AuthInProgress
		return AuthInProgressStyle
	case 3: // AuthSuccess
		return AuthSuccessStyle
	case 4: // AuthFailed
		return AuthFailedStyle
	default:
		return lipgloss.NewStyle()
	}
}
