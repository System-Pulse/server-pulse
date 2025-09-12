package widgets

import "github.com/charmbracelet/lipgloss"

var (
	accentColor  = lipgloss.Color("#06b6d4") // Cyan
	successColor = lipgloss.Color("#10b981") // Emerald
	errorColor   = lipgloss.Color("#ef4444") // Red

	surfaceColor          = lipgloss.Color("235") // black
	purpleCollor          = lipgloss.Color("57")  // purple
	buttonCollorDesactive = lipgloss.Color("236")
	clearWhite            = lipgloss.Color("229") // clear white
	whiteColor            = lipgloss.Color("255") // white
	cardStyle             = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(surfaceColor).
				Background(surfaceColor).
				Padding(1, 2).
				Margin(1, 0)

	cardNetworkStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purpleCollor)

	cardTableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surfaceColor).
		// Background(surfaceColor).
		Padding(1, 2).
		Margin(1, 0)

	cardButtonStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleCollor).
			Background(purpleCollor).
			Padding(1, 2).
			Bold(true).
			Margin(1, 0, 0, 0)

	cardButtonStyleDesactive = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(buttonCollorDesactive).
					Background(buttonCollorDesactive).
					Padding(1, 2).
					Margin(1, 0, 0, 0)

	// Styles pour les métriques
	metricLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cbd5e1")).
				Bold(true)

	metricValueStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)
	// ------------------------------------ //

	menuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(1).
			Background(lipgloss.Color("235"))

	searchBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 1).
			MarginBottom(1)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Padding(0, 2).
		// Background(accentColor).
		Foreground(lipgloss.Color("black")).
		Bold(true)

	ipTableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surfaceColor).
			Background(surfaceColor).
			Margin(1, 0)

	ipTableHeaderStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Padding(0, 1)

	ipTableCellStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Padding(0, 1)

	NetworkTableStatusStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true).
				Padding(0, 1)

	NetworkTableStatusDownStyle = lipgloss.NewStyle().
					Foreground(errorColor).
					Bold(true).
					Padding(0, 1)
)
