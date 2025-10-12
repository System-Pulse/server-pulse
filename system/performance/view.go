package performance

import (
	"fmt"
	"strings"

	"github.com/System-Pulse/server-pulse/widgets/vars"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

func RenderSystemHealthView(healthLoading bool, healthMetrics *HealthMetrics, healthScore *HealthScore) string {
	if healthLoading {
		return vars.CardStyle.Render("⏳ Loading System Health...")
	}

	if healthMetrics == nil || healthScore == nil {
		return vars.CardStyle.Render("Press 'r' to load system health metrics.")
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(vars.AccentColor).MarginBottom(1)
	b.WriteString(titleStyle.Render("─ SYSTEM HEALTH ANALYSIS ─"))
	b.WriteString("\n")

	// Health Score
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 20
	prog.Full = '█'
	prog.Empty = '░'

	scoreStr := fmt.Sprintf(" %d/100", healthScore.Score)
	healthStatus := " [Good]"
	if healthScore.Score < 50 {
		healthStatus = " [Poor]"
	} else if healthScore.Score < 80 {
		healthStatus = " [Fair]"
	}
	b.WriteString(fmt.Sprintf("│ Health Score: %s%s%s \n\n", prog.ViewAs(float64(healthScore.Score)/100.0), scoreStr, healthStatus))

	// Detected Issues
	if len(healthScore.Issues) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(vars.ErrorColor).Render("⚠️ DETECTED ISSUES"))
		b.WriteString("\n")
		for _, issue := range healthScore.Issues {
			b.WriteString(fmt.Sprintf("- %s \n", issue))
		}
		b.WriteString("\n")
	}

	// Recommendations
	if len(healthScore.Recommendations) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(vars.SuccessColor).Render("💡 RECOMMENDATIONS"))
		b.WriteString("\n")
		for _, rec := range healthScore.Recommendations {
			b.WriteString(fmt.Sprintf("- %s \n", rec))
		}
		b.WriteString("\n")
	}

	return vars.CardStyle.Render(b.String())
}
