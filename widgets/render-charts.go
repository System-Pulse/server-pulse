package widgets

import (
	"fmt"
	"math"
	// "sort"
	// "strings"

	model "github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

func (m Model) renderCPUChart(width, height int) string {
	// Get container-specific history if available
	var points []float64
	if m.Monitor.SelectedContainer != nil {
		if containerHistory, exists := m.Monitor.ContainerHistories[m.Monitor.SelectedContainer.ID]; exists {
			points = extractValues(containerHistory.CpuHistory.Points)
		}
	}

	// Fallback to global history if no container-specific data
	if len(points) < 1 {
		points = extractValues(m.Monitor.CpuHistory.Points)
	}

	if len(points) < 1 {
		return renderEmptyChart("CPU Usage Over Time (%)", height)
	}

	// Get optimal scale for CPU usage (dynamic scaling for low values)
	points, maxValue, caption := getOptimalCPUScale(points)

	graph := asciigraph.Plot(
		points,
		asciigraph.Width(width),
		asciigraph.Height(height-2),
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(maxValue),
		asciigraph.Caption(caption),
		asciigraph.SeriesColors(asciigraph.Cyan),
	)

	return graph
}

func (m Model) renderMemoryChart(width, height int) string {
	// Get container-specific history if available
	var points []float64
	if m.Monitor.SelectedContainer != nil {
		if containerHistory, exists := m.Monitor.ContainerHistories[m.Monitor.SelectedContainer.ID]; exists {
			points = extractValues(containerHistory.MemoryHistory.Points)
		}
	}

	// Fallback to global history if no container-specific data
	if len(points) < 1 {
		points = extractValues(m.Monitor.MemoryHistory.Points)
	}

	if len(points) < 2 {
		lenP := fmt.Sprintf("%d", len(points))
		return renderEmptyChart("Memory Usage Over Time (%)-> "+lenP, height)
	}

	// Get optimal scale for memory usage (dynamic scaling for low values)
	points, maxValue, caption := getOptimalMemoryScale(points)

	graph := asciigraph.Plot(
		points,
		asciigraph.Width(width),
		asciigraph.Height(height-2),
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(maxValue),
		asciigraph.Caption(caption),
		asciigraph.SeriesColors(asciigraph.Cyan),
	)

	return graph
}

func (m Model) renderNetworkRXChart(width, height int) string {
	// Get container-specific history if available
	var points []float64
	var maxValue float64
	caption := "Network RX"

	if m.Monitor.SelectedContainer != nil {
		if containerHistory, exists := m.Monitor.ContainerHistories[m.Monitor.SelectedContainer.ID]; exists {
			points = extractValues(containerHistory.NetworkRxHistory.Points)
			points, maxValue, caption = getOptimalNetworkScale(points, "RX")
		}
	}

	// Fallback to global history if no container-specific data
	if len(points) < 1 {
		points = extractValues(m.Monitor.NetworkRxHistory.Points)
		points, maxValue, caption = getOptimalNetworkScale(points, "RX")
	}

	if len(points) < 1 {
		return renderEmptyChart("Network RX", height)
	}

	// Ensure minimum scale for visibility
	if maxValue < 0.1 {
		maxValue = 0.1
	}

	graph := asciigraph.Plot(
		points,
		asciigraph.Width(width),
		asciigraph.Height(height-2),
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(maxValue),
		asciigraph.Caption(caption),
		asciigraph.SeriesColors(asciigraph.Cyan),
	)

	return graph
}

func (m Model) renderNetworkTXChart(width, height int) string {
	// Get container-specific history if available
	var points []float64
	var maxValue float64
	caption := "Network TX"

	if m.Monitor.SelectedContainer != nil {
		if containerHistory, exists := m.Monitor.ContainerHistories[m.Monitor.SelectedContainer.ID]; exists {
			points = extractValues(containerHistory.NetworkTxHistory.Points)
			points, maxValue, caption = getOptimalNetworkScale(points, "TX")
		}
	}

	// Fallback to global history if no container-specific data
	if len(points) < 1 {
		points = extractValues(m.Monitor.NetworkTxHistory.Points)
		points, maxValue, caption = getOptimalNetworkScale(points, "TX")
	}

	if len(points) < 1 {
		return renderEmptyChart("Network TX", height)
	}

	// Ensure minimum scale for visibility
	if maxValue < 0.1 {
		maxValue = 0.1
	}

	graph := asciigraph.Plot(
		points,
		asciigraph.Width(width),
		asciigraph.Height(height-2),
		asciigraph.LowerBound(0),
		asciigraph.UpperBound(maxValue),
		asciigraph.Caption(caption),
		asciigraph.SeriesColors(asciigraph.Cyan),
	)
	return graph
}

func extractValues(points []model.DataPoint) []float64 {
	values := make([]float64, len(points))
	for i, point := range points {
		values[i] = point.Value
	}
	return values
}

func renderEmptyChart(title string, height int) string {
	var builder string
	builder = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render(title) + "\n"

	for y := 0; y < height-2; y++ {
		builder += "\n"
	}

	builder += "Collecting data..."
	return builder
}

// getOptimalNetworkScale determines the best scale and unit for network data
func getOptimalNetworkScale(points []float64, direction string) ([]float64, float64, string) {
	if len(points) == 0 {
		return points, 10.0, "Network " + direction + " (MB/s)"
	}

	maxValue := getMaxValueFromFloat64(points)
	caption := "Network " + direction

	// Determine the best unit and scale based on the maximum value
	if maxValue < 0.001 { // Less than 1 KB/s in MB units
		// Convert to KB/s for better visibility
		convertedPoints := make([]float64, len(points))
		for i, val := range points {
			convertedPoints[i] = val * 1024 // Convert MB to KB
		}
		maxConverted := getMaxValueFromFloat64(convertedPoints) * 1.2
		// Ensure minimum scale for KB/s
		if maxConverted < 1.0 {
			maxConverted = 1.0
		}
		return convertedPoints, maxConverted, caption + " (KB/s)"
	} else if maxValue < 0.1 { // Less than 100 KB/s in MB units
		maxValue *= 1.2
		// Ensure minimum scale for visibility with small MB values
		if maxValue < 0.01 {
			maxValue = 0.01
		}
		return points, maxValue, caption + " (MB/s)"
	} else if maxValue < 1.0 { // Less than 1 MB/s
		maxValue *= 1.2
		return points, maxValue, caption + " (MB/s)"
	} else {
		// Normal MB/s scale
		maxValue *= 1.2
		return points, maxValue, caption + " (MB/s)"
	}
}

func getMaxValueFromFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0.001
	}
	max := values[0]
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return math.Max(max, 0.001) // Ensure at least 0.001 to avoid division by zero
}

// getOptimalCPUScale determines the best scale for CPU usage data
func getOptimalCPUScale(points []float64) ([]float64, float64, string) {
	if len(points) == 0 {
		return points, 10.0, "CPU Usage (%)"
	}

	maxValue := getMaxValueFromFloat64(points)
	caption := "CPU Usage (%)"

	// Dynamic scaling for low CPU usage
	if maxValue < 1.0 { // Less than 1%
		// Scale up for better visibility
		convertedPoints := make([]float64, len(points))
		for i, val := range points {
			convertedPoints[i] = val * 10 // Scale by 10x for better visibility
		}
		maxConverted := getMaxValueFromFloat64(convertedPoints) * 1.2
		// Ensure minimum scale for visibility
		if maxConverted < 5.0 {
			maxConverted = 5.0
		}
		return convertedPoints, maxConverted, "CPU Usage (×10%)"
	} else if maxValue < 5.0 { // Less than 5%
		maxValue *= 1.2
		// Ensure minimum scale for visibility
		if maxValue < 2.0 {
			maxValue = 2.0
		}
		return points, maxValue, caption
	} else if maxValue < 20.0 { // Less than 20%
		maxValue *= 1.2
		return points, maxValue, caption
	} else {
		// Normal scale for higher usage
		maxValue = math.Min(maxValue*1.2, 100.0)
		return points, maxValue, caption
	}
}

// getOptimalMemoryScale determines the best scale for memory usage data
func getOptimalMemoryScale(points []float64) ([]float64, float64, string) {
	if len(points) == 0 {
		return points, 10.0, "Memory Usage (%)"
	}

	maxValue := getMaxValueFromFloat64(points)
	caption := "Memory Usage (%)"

	// Dynamic scaling for low memory usage
	if maxValue < 1.0 { // Less than 1%
		// Scale up for better visibility
		convertedPoints := make([]float64, len(points))
		for i, val := range points {
			convertedPoints[i] = val * 10 // Scale by 10x for better visibility
		}
		maxConverted := getMaxValueFromFloat64(convertedPoints) * 1.2
		// Ensure minimum scale for visibility
		if maxConverted < 5.0 {
			maxConverted = 5.0
		}
		return convertedPoints, maxConverted, "Memory Usage (×10%)"
	} else if maxValue < 5.0 { // Less than 5%
		maxValue *= 1.2
		// Ensure minimum scale for visibility
		if maxValue < 2.0 {
			maxValue = 2.0
		}
		return points, maxValue, caption
	} else if maxValue < 20.0 { // Less than 20%
		maxValue *= 1.2
		return points, maxValue, caption
	} else {
		// Normal scale for higher usage
		maxValue = math.Min(maxValue*1.2, 100.0)
		return points, maxValue, caption
	}
}

/*/ renderPerCPUChart renders individual CPU core usage charts
func (m Model) renderPerCPUChart(width, height int) []string {
	if m.Monitor.SelectedContainer == nil {
		return []string{renderEmptyChart("Per-CPU Usage", height)}
	}

	var charts []string
	containerID := m.Monitor.SelectedContainer.ID

	// Get container-specific history
	if containerHistory, exists := m.Monitor.ContainerHistories[containerID]; exists {
		if len(containerHistory.PerCpuHistory) == 0 {
			return []string{renderEmptyChart("Per-CPU Usage (No data)", height)}
		}

		// Get sorted core indices for consistent ordering
		coreIndices := make([]int, 0, len(containerHistory.PerCpuHistory))
		for coreIndex := range containerHistory.PerCpuHistory {
			coreIndices = append(coreIndices, coreIndex)
		}
		sort.Ints(coreIndices)

		for _, coreIndex := range coreIndices {
			coreHistory := containerHistory.PerCpuHistory[coreIndex]
			points := extractValues(coreHistory.Points)

			if len(points) < 1 {
				charts = append(charts, renderEmptyChart(fmt.Sprintf("CPU %d", coreIndex), height))
				continue
			}

			// Get optimal scale for per-core CPU usage
			scaledPoints, maxValue, caption := getOptimalCPUScale(points)
			coreCaption := fmt.Sprintf("CPU %d %s", coreIndex, strings.TrimPrefix(caption, "CPU Usage"))

			graph := asciigraph.Plot(
				scaledPoints,
				asciigraph.Width(width/2-2),
				asciigraph.Height(height-2),
				asciigraph.LowerBound(0),
				asciigraph.UpperBound(maxValue),
				asciigraph.Caption(coreCaption),
				asciigraph.SeriesColors(asciigraph.Cyan),
			)
			charts = append(charts, graph)
		}
	} else {
		charts = append(charts, renderEmptyChart("Per-CPU Usage", height))
	}

	return charts
}

// renderAllPerCPUCharts renders all CPU core charts in a grid layout
func (m Model) renderAllPerCPUCharts(totalWidth, height int) string {
	charts := m.renderPerCPUChart(totalWidth/2, height)

	if len(charts) == 0 {
		return renderEmptyChart("Per-CPU Usage", height)
	}

	// Create a grid layout (2 charts per row)
	var rows []string
	for i := 0; i < len(charts); i += 2 {
		var rowCharts []string
		if i < len(charts) {
			rowCharts = append(rowCharts, charts[i])
		}
		if i+1 < len(charts) {
			rowCharts = append(rowCharts, charts[i+1])
		}

		if len(rowCharts) > 0 {
			row := lipgloss.JoinHorizontal(lipgloss.Top, rowCharts...)
			rows = append(rows, row)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
*/