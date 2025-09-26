package auth

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	AuthRequiredStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")) // Orange

	AuthSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("46")) // Green

	AuthFailedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")) // Red

	AuthInProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange

	AuthPromptStyle = lipgloss.NewStyle().
			Bold(true)

	AuthInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")) // Gray

	LockedCheckStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange

	AccessIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange
)

const (
	AuthRequiredMessage = "🔐 Admin Authentication Required"

	AuthSuccessMessage = "✅ Authentication Successful"

	AuthFailedMessage = "❌ Authentication Failed"

	AuthInProgressMessage = "⏳ Authenticating..."

	AuthPromptMessage = "Enter password for admin access:"

	AuthInstructions = "Press Enter to authenticate, Esc to cancel"

	AuthRetryMessage = "Press 'a' to retry authentication"

	AdminAccessGranted = "✅ Admin access granted"

	AdminAccessRequired = "🔒 Some checks require admin privileges. Press 'a' to authenticate"

	SudoNotAvailable = "Sudo is not available. Please run as root."

	InvalidPasswordMessage = "Invalid password or insufficient privileges"

	AuthenticationFailedGeneric = "Authentication failed"

	LockedCheckIndicator = "🔒"
)

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
