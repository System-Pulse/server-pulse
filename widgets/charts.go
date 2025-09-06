package widgets

import (
	"fmt"
	"math"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)


func (m model) renderCPUChart(width, height int) string {
	return m.renderLineChart(m.cpuHistory, "CPU Usage Over Time (%)", width, height, 0, 100)
}

func (m model) renderMemoryChart(width, height int) string {
	return m.renderLineChart(m.memoryHistory, "Memory Usage Over Time (%)", width, height, 0, 100)
}

func (m model) renderNetworkRXChart(width, height int) string {
	maxValue := m.getMaxValue(m.networkRxHistory.Points) * 1.2
	return m.renderLineChart(m.networkRxHistory, "Network RX (MB/s)", width, height, 0, maxValue)
}

func (m model) renderNetworkTXChart(width, height int) string {
	maxValue := m.getMaxValue(m.networkTxHistory.Points) * 1.2
	return m.renderLineChart(m.networkTxHistory, "Network TX (MB/s)", width, height, 0, maxValue)
}

func (m model) renderLineChart(history DataHistory, title string, width, height int, minValue, maxValue float64) string {
	if len(history.Points) < 2 {
		return m.renderEmptyChart(title, width, height)
	}
	
	// Préparer les données pour le rendu
	points := make([]float64, len(history.Points))
	for i, point := range history.Points {
		points[i] = point.Value
	}
	
	// Créer le graphique
	var builder strings.Builder
	
	// Titre
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	builder.WriteString(titleStyle.Render(title))
	builder.WriteString("\n")
	
	// Axe Y et données
	chartHeight := height - 3 // Réserver de l'espace pour le titre et l'axe X
	for y := chartHeight - 1; y >= 0; y-- {
		// Étiquette Y
		yValue := minValue + (maxValue-minValue)*float64(y)/float64(chartHeight-1)
		yLabel := fmt.Sprintf("%4.0f ┤", yValue)
		builder.WriteString(yLabel)
		
		// Points du graphique
		for x := 0; x < len(points) && x < width-len(yLabel)-1; x++ {
			normalizedValue := (points[x] - minValue) / (maxValue - minValue)
			chartY := int(normalizedValue * float64(chartHeight-1))
			
			if chartY == y {
				builder.WriteString("●") // Point actuel
			} else if y < chartY && y > 0 {
				// Ligne de connexion
				prevNormalized := 0.0
				if x > 0 {
					prevNormalized = (points[x-1] - minValue) / (maxValue - minValue)
				}
				prevChartY := int(prevNormalized * float64(chartHeight-1))
				
				if (y >= prevChartY && y <= chartY) || (y >= chartY && y <= prevChartY) {
					builder.WriteString("╲")
				} else {
					builder.WriteString(" ")
				}
			} else {
				builder.WriteString(" ")
			}
		}
		builder.WriteString("\n")
	}
	
	// Axe X
	builder.WriteString("    0 ┼")
	for i := 0; i < width-7; i++ {
		builder.WriteString("─")
	}
	builder.WriteString("\n")
	
	// Étiquettes de temps
	if len(history.Points) > 1 {
		startTime := history.Points[0].Timestamp.Format("15:04")
		endTime := history.Points[len(history.Points)-1].Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("      %s%s%s", startTime, 
			strings.Repeat(" ", width-14), endTime))
	}
	
	return builder.String()
}

func (m model) renderEmptyChart(title string, width, height int) string {
	var builder strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	builder.WriteString(titleStyle.Render(title))
	builder.WriteString("\n")
	
	for y := 0; y < height-2; y++ {
		builder.WriteString(strings.Repeat(" ", width))
		builder.WriteString("\n")
	}
	
	builder.WriteString("Collecting data...")
	return builder.String()
}

func (m model) getMaxValue(points []DataPoint) float64 {
	max := 0.0
	for _, point := range points {
		if point.Value > max {
			max = point.Value
		}
	}
	return math.Max(max, 1.0) // Éviter la division par zéro
}