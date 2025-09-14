package vars

import "github.com/charmbracelet/lipgloss"

var (
	AccentColor  = lipgloss.Color("#06b6d4") // Cyan
	SuccessColor = lipgloss.Color("#10b981") // Emerald
	ErrorColor   = lipgloss.Color("#ef4444") // Red

	SurfaceColor          = lipgloss.Color("235") // black
	PurpleCollor          = lipgloss.Color("57")  // purple
	ButtonCollorDesactive = lipgloss.Color("236")
	ClearWhite            = lipgloss.Color("229") // clear white
	WhiteColor            = lipgloss.Color("255") // white
	CardStyle             = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SurfaceColor).
				Background(SurfaceColor).
				Padding(1, 2).
				Margin(1, 0)

	CardNetworkStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PurpleCollor)

	CardTableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SurfaceColor).
		// Background(surfaceColor).
		Padding(1, 2).
		Margin(1, 0)

	CardButtonStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PurpleCollor).
			Background(PurpleCollor).
			Padding(1, 2).
			Bold(true).
			Margin(1, 0, 0, 0)

	CardButtonStyleDesactive = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ButtonCollorDesactive).
					Background(ButtonCollorDesactive).
					Padding(1, 2).
					Margin(1, 0, 0, 0)

	MetricLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cbd5e1")).
				Bold(true)

	MetricValueStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true)
	// ------------------------------------ //

	MenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(1).
			Background(lipgloss.Color("235"))

	SearchBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 1).
			MarginBottom(1)

	NetworkTableStatusStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true).
				Padding(0, 1)

	NetworkTableStatusDownStyle = lipgloss.NewStyle().
					Foreground(ErrorColor).
					Bold(true).
					Padding(0, 1)

	AsciiArt = `
 ▗▄▄▖▗▄▄▄▖▗▄▄▖ ▗▖  ▗▖▗▄▄▄▖▗▄▄▖     ▗▄▄▖ ▗▖ ▗▖▗▖    ▗▄▄▖▗▄▄▄▖
▐▌   ▐▌   ▐▌ ▐▌▐▌  ▐▌▐▌   ▐▌ ▐▌    ▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌   ▐▌
 ▝▀▚▖▐▛▀▀▘▐▛▀▚▖▐▌  ▐▌▐▛▀▀▘▐▛▀▚▖    ▐▛▀▘ ▐▌ ▐▌▐▌    ▝▀▚▖▐▛▀▀▘
▗▄▄▞▘▐▙▄▄▖▐▌ ▐▌ ▝▚▞▘ ▐▙▄▄▖▐▌ ▐▌    ▐▌   ▝▚▄▞▘▐▙▄▄▖▗▄▄▞▘▐▙▄▄▖
		`
)
