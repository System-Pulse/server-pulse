package performance

import (
	"fmt"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/widgets/model"
	"github.com/System-Pulse/server-pulse/widgets/vars"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/mem"
)

// RenderMemoryOverview renders the always-visible memory overview section
func RenderMemoryOverview(metrics *model.MemoryMetrics) string {
	if metrics == nil {
		return vars.CardStyle.Render("Loading memory data...")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Width(20).
		Foreground(lipgloss.Color("245"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250"))

	progressStyle := lipgloss.NewStyle().
		Width(50)

	// Create progress bar for memory usage
	usageBar := createMemoryProgressBar(metrics.UsedPercent, 50)

	overview := titleStyle.Render("MEMORY OVERVIEW") + "\n\n"

	overview += labelStyle.Render("Total Memory:") + " " +
		valueStyle.Render(formatBytes(metrics.Total)) + "\n"

	overview += labelStyle.Render("Used:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Used), metrics.UsedPercent)) + "\n"

	overview += labelStyle.Render("Available:") + " " +
		valueStyle.Render(formatBytes(metrics.Available)) + "\n"

	overview += labelStyle.Render("Free:") + " " +
		valueStyle.Render(formatBytes(metrics.Free)) + "\n\n"

	overview += progressStyle.Render(usageBar) + "\n"

	return vars.CardStyle.Render(overview)
}

// RenderMemoryUsageBreakdown renders the Usage Breakdown sub-tab
func RenderMemoryUsageBreakdown(metrics *model.MemoryMetrics) string {
	if metrics == nil {
		return vars.CardStyle.Render("Loading memory usage data...")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Width(25).
		Foreground(lipgloss.Color("245"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	content := titleStyle.Render("USAGE BREAKDOWN") + "\n\n"

	// Application Memory
	appMemPercent := float64(metrics.ApplicationMem) / float64(metrics.Total) * 100
	content += labelStyle.Render("Application Memory:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.ApplicationMem), appMemPercent)) + "\n"
	content += "  " + descStyle.Render("Memory actively used by applications") + "\n\n"

	// Buffers
	buffersPercent := float64(metrics.Buffers) / float64(metrics.Total) * 100
	content += labelStyle.Render("Buffers:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Buffers), buffersPercent)) + "\n"
	content += "  " + descStyle.Render("I/O buffers for disk operations") + "\n\n"

	// Cached
	cachedPercent := float64(metrics.Cached) / float64(metrics.Total) * 100
	content += labelStyle.Render("Cached:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Cached), cachedPercent)) + "\n"
	content += "  " + descStyle.Render("File system cache (can be freed if needed)") + "\n\n"

	// Available
	availablePercent := float64(metrics.Available) / float64(metrics.Total) * 100
	content += labelStyle.Render("Available:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Available), availablePercent)) + "\n"
	content += "  " + descStyle.Render("Memory available for applications") + "\n\n"

	// Shared
	sharedPercent := float64(metrics.Shared) / float64(metrics.Total) * 100
	content += labelStyle.Render("Shared:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Shared), sharedPercent)) + "\n"
	content += "  " + descStyle.Render("Memory shared between processes") + "\n"

	return vars.CardStyle.Render(content)
}

// RenderMemorySwapAnalysis renders the Swap Analysis sub-tab
func RenderMemorySwapAnalysis(metrics *model.MemoryMetrics) string {
	if metrics == nil {
		return vars.CardStyle.Render("Loading swap data...")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Width(25).
		Foreground(lipgloss.Color("245"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250"))

	progressStyle := lipgloss.NewStyle().
		Width(50)

	content := titleStyle.Render("SWAP ANALYSIS") + "\n\n"

	// Swap Overview
	if metrics.SwapTotal == 0 {
		content += labelStyle.Render("Swap Status:") + " " +
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No swap configured") + "\n"
	} else {
		content += labelStyle.Render("Swap Total:") + " " +
			valueStyle.Render(formatBytes(metrics.SwapTotal)) + "\n"

		content += labelStyle.Render("Swap Used:") + " " +
			valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
				formatBytes(metrics.SwapUsed), metrics.SwapUsedPercent)) + "\n"

		content += labelStyle.Render("Swap Free:") + " " +
			valueStyle.Render(formatBytes(metrics.SwapFree)) + "\n\n"

		// Swap progress bar
		swapBar := createMemoryProgressBar(metrics.SwapUsedPercent, 50)
		content += progressStyle.Render(swapBar) + "\n\n"

		// SwapCached
		content += labelStyle.Render("Swap Cached:") + " " +
			valueStyle.Render(formatBytes(metrics.SwapCached)) + "\n"
		content += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).
			Render("Swap memory also present in RAM") + "\n\n"

		// Health Status
		content += titleStyle.Render("STATUS") + "\n\n"
		status := getSwapHealthStatus(metrics)
		content += status
	}

	return vars.CardStyle.Render(content)
}

// RenderMemorySystemMemory renders the System Memory sub-tab
func RenderMemorySystemMemory(metrics *model.MemoryMetrics) string {
	if metrics == nil {
		return vars.CardStyle.Render("Loading system memory data...")
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Width(25).
		Foreground(lipgloss.Color("245"))

	valueStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	content := titleStyle.Render("SYSTEM MEMORY") + "\n\n"

	// Dirty Pages
	dirtyPercent := float64(metrics.Dirty) / float64(metrics.Total) * 100
	content += labelStyle.Render("Dirty Pages:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Dirty), dirtyPercent)) + "\n"
	content += "  " + descStyle.Render("Modified data waiting to be written to disk") + "\n\n"

	// WriteBack
	writebackPercent := float64(metrics.WriteBack) / float64(metrics.Total) * 100
	content += labelStyle.Render("WriteBack:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.WriteBack), writebackPercent)) + "\n"
	content += "  " + descStyle.Render("Data currently being written to disk") + "\n\n"

	// Slab
	slabPercent := float64(metrics.Slab) / float64(metrics.Total) * 100
	content += labelStyle.Render("Slab:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.Slab), slabPercent)) + "\n"
	content += "  " + descStyle.Render("Kernel data structure cache") + "\n\n"

	// PageTables
	pageTablesPercent := float64(metrics.PageTables) / float64(metrics.Total) * 100
	content += labelStyle.Render("Page Tables:") + " " +
		valueStyle.Render(fmt.Sprintf("%s (%.1f%%)",
			formatBytes(metrics.PageTables), pageTablesPercent)) + "\n"
	content += "  " + descStyle.Render("Memory used for page table management") + "\n\n"

	// Diagnostics
	content += titleStyle.Render("DIAGNOSTICS") + "\n\n"

	if metrics.Dirty > 1024*1024*1024 { // > 1GB
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⚠ ") +
			"High dirty pages - data waiting to be written\n"
	} else {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("✓ ") +
			"Dirty pages within normal range\n"
	}

	if metrics.WriteBack > 1024*1024*100 { // > 100MB
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⚠ ") +
			"Active writeback - I/O operations in progress\n"
	} else {
		content += lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("✓ ") +
			"Minimal writeback activity\n"
	}

	return vars.CardStyle.Render(content)
}

// RenderMemoryContent renders the complete memory content with navigation
func RenderMemoryContent(metrics *model.MemoryMetrics) string {
	overview := RenderMemoryOverview(metrics)
	return overview
}

// Helper function to get swap health status
func getSwapHealthStatus(metrics *model.MemoryMetrics) string {
	goodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	criticalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	// No swap used - good
	if metrics.SwapUsed == 0 {
		return goodStyle.Render("✓ No swap used") + " - System has sufficient RAM\n"
	}

	// Swap used but memory available - configuration issue
	availablePercent := float64(metrics.Available) / float64(metrics.Total) * 100
	if metrics.SwapUsed > 0 && availablePercent > 20 {
		return warningStyle.Render("⚠ Swap used with available RAM") +
			" - Check swappiness setting\n" +
			"  Recommendation: Reduce vm.swappiness value\n"
	}

	// Swap used and memory low - critical
	if metrics.SwapUsedPercent > 50 {
		return criticalStyle.Render("✗ High swap usage") +
			" - System under memory pressure\n" +
			"  Recommendation: Add more RAM or reduce memory usage\n"
	}

	// Moderate swap usage
	return warningStyle.Render("⚠ Moderate swap usage") +
		" - Monitor for performance issues\n"
}

// Helper function to create a memory progress bar
func createMemoryProgressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	var color string
	if percent < 70 {
		color = "82" // Green
	} else if percent < 85 {
		color = "220" // Yellow
	} else {
		color = "196" // Red
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	styledBar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(bar)

	return fmt.Sprintf("%s %.1f%%", styledBar, percent)
}

// Helper function to format bytes to human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetMemoryMetrics collects comprehensive memory performance metrics
func GetMemoryMetrics() tea.Cmd {
	return func() tea.Msg {
		metrics, err := collectMemoryMetrics()
		return MemoryMetricsMsg{
			Metrics: metrics,
			Error:   err,
		}
	}
}

func collectMemoryMetrics() (*model.MemoryMetrics, error) {
	metrics := &model.MemoryMetrics{
		LastUpdate: time.Now(),
	}

	// Get virtual memory stats
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}

	// Overview data
	metrics.Total = vmem.Total
	metrics.Used = vmem.Used
	metrics.Available = vmem.Available
	metrics.Free = vmem.Free
	metrics.UsedPercent = vmem.UsedPercent

	// Usage Breakdown data
	metrics.Buffers = vmem.Buffers
	metrics.Cached = vmem.Cached
	metrics.Shared = vmem.Shared

	// Calculate application memory (Used - Buffers - Cached)
	// This represents memory actually used by applications
	if vmem.Used > (vmem.Buffers + vmem.Cached) {
		metrics.ApplicationMem = vmem.Used - vmem.Buffers - vmem.Cached
	} else {
		metrics.ApplicationMem = vmem.Used
	}

	// Get swap memory stats
	swap, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get swap info: %v", err)
	}

	// Swap Analysis data
	metrics.SwapTotal = swap.Total
	metrics.SwapUsed = swap.Used
	metrics.SwapFree = swap.Free
	metrics.SwapUsedPercent = swap.UsedPercent
	metrics.SwapCached = vmem.SwapCached

	// System Memory data
	metrics.Dirty = vmem.Dirty
	metrics.WriteBack = vmem.WriteBack
	metrics.Slab = vmem.Slab
	metrics.PageTables = vmem.PageTables

	return metrics, nil
}
