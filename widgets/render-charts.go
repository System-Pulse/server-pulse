package widgets

import (
	"fmt"
	"math"
	"strings"

	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderCPUChart(width, height int) string {
	
	return renderLineChart(m.Monitor.CpuHistory, "CPU Usage Over Time (%)", width, height, 0, 100)
}

func (m Model) renderMemoryChart(width, height int) string {
	
	return renderLineChart(m.Monitor.MemoryHistory, "Memory Usage Over Time (%)", width, height, 0, 100)
}

func (m Model) renderNetworkRXChart(width, height int) string {
	
	maxValue := getMaxValue(m.Monitor.NetworkRxHistory.Points) * 1.2
	return renderLineChart(m.Monitor.NetworkRxHistory, "Network RX (MB/s)", width, height, 0, maxValue)
}

func (m Model) renderNetworkTXChart(width, height int) string {
	
	maxValue := getMaxValue(m.Monitor.NetworkTxHistory.Points) * 1.2
	return renderLineChart(m.Monitor.NetworkTxHistory, "Network TX (MB/s)", width, height, 0, maxValue)
}

func renderLineChart(history model.DataHistory, title string, width, height int, minValue, maxValue float64) string {
	if width < 10 || height < 3 {
		return title + "\n[Chart too small to display]"
	}
	if len(history.Points) < 2 {
		return renderEmptyChart(title, width, height)
	}

	points := make([]float64, len(history.Points))
	for i, point := range history.Points {
		points[i] = point.Value
	}

	var builder strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	builder.WriteString(titleStyle.Render(title))
	builder.WriteString("\n")

	chartHeight := height - 3 
	for y := chartHeight - 1; y >= 0; y-- {
		yValue := minValue + (maxValue-minValue)*float64(y)/float64(chartHeight-1)
		yLabel := fmt.Sprintf("%4.0f ┤", yValue)
		builder.WriteString(yLabel)

		for x := 0; x < len(points) && x < width-len(yLabel)-1; x++ {
			normalizedValue := (points[x] - minValue) / (maxValue - minValue)
			chartY := int(normalizedValue * float64(chartHeight-1))

			if chartY == y {
				builder.WriteString("●")
			} else if y < chartY && y > 0 {
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

	builder.WriteString("    0 ┼")
	for i := 0; i < width-7; i++ {
		builder.WriteString("─")
	}
	builder.WriteString("\n")

	if len(history.Points) > 1 {
		startTime := history.Points[0].Timestamp.Format("15:04")
		endTime := history.Points[len(history.Points)-1].Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("      %s%s%s", startTime,
			strings.Repeat(" ", width-14), endTime))
	}

	return builder.String()
}

func renderEmptyChart(title string, width, height int) string {
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

func getMaxValue(points []model.DataPoint) float64 {
	max := 0.0
	for _, point := range points {
		if point.Value > max {
			max = point.Value
		}
	}
	return math.Max(max, 1.0)
}
