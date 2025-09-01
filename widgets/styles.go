package widgets

import "github.com/charmbracelet/lipgloss"

var (
	// Couleurs principales
	accentColor  = lipgloss.Color("#06b6d4") // Cyan
	successColor = lipgloss.Color("#10b981") // Emerald
	errorColor   = lipgloss.Color("#ef4444") // Red

	surfaceColor = lipgloss.Color("#1e293b") // Slate-800
	buttonCollor = lipgloss.Color("57")
	buttonCollorDesactive = lipgloss.Color("236")
	

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surfaceColor).
			Background(surfaceColor).
			Padding(1, 2).
			Margin(1, 0)
	
	cardButtonStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(buttonCollor).
			Background(buttonCollor).
			Padding(1, 2).
			Bold(true).
			Margin(1, 0)
	
	cardButtonStyleDesactive = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(buttonCollorDesactive).
			Background(buttonCollorDesactive).
			Padding(1, 2).
			Margin(1, 0)

	// Styles pour les métriques
	metricLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cbd5e1")).
				Bold(true)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)
)
