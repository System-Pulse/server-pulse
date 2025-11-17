package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/system/security"
)

type ReportStatus string

const (
	StatusHealthy  ReportStatus = "✅ Healthy"
	StatusWarning  ReportStatus = "⚠️ Warning"
	StatusCritical ReportStatus = "❌ Critical"
)

type ReturnToMenuMsg struct{}

type ClearSaveNotificationMsg struct{}

type ReportGenerationMsg struct {
	Report string
}

type ReportModel struct {
	CurrentReport        string
	SavedReports         []string
	SelectedReport       int
	IsGenerating         bool
	IsSaving             bool
	ShowSavedReports     bool
	LastGenerated        time.Time
	ReportDirectory      string
	SaveNotification     string
	SaveNotificationTime time.Time
}

func NewReportModel() ReportModel {
	homeDir, _ := os.UserHomeDir()
	reportDir := filepath.Join(homeDir, ".server-pulse", "reports")

	// Ensure report directory exists
	os.MkdirAll(reportDir, 0755)

	return ReportModel{
		ReportDirectory:      reportDir,
		SavedReports:         []string{},
		SaveNotification:     "",
		SaveNotificationTime: time.Time{},
	}
}

func (rm *ReportModel) GenerateReport(m MonitorModel, diagnostic DiagnosticModel, securityChecks []security.SecurityCheck) string {
	var report strings.Builder

	// 1. Executive Summary
	report.WriteString(rm.generateExecutiveSummary(m, diagnostic, securityChecks))
	report.WriteString("\n\n")

	// 2. System Identification
	report.WriteString(rm.generateSystemIdentification(m))
	report.WriteString("\n\n")

	// 3. Resource Usage
	report.WriteString(rm.generateResourceUsage(m))
	report.WriteString("\n\n")

	// 4. Security Status
	report.WriteString(rm.generateSecurityStatus(securityChecks))
	report.WriteString("\n\n")

	// 5. Docker Container Status
	report.WriteString(rm.generateContainerStatus(m))
	report.WriteString("\n\n")

	// 6. Performance Insights
	report.WriteString(rm.generatePerformanceInsights(m, diagnostic))
	report.WriteString("\n\n")

	// 7. Recommendations
	report.WriteString(rm.generateRecommendations(m, securityChecks))

	rm.CurrentReport = report.String()
	rm.LastGenerated = time.Now()

	return rm.CurrentReport
}

func (rm *ReportModel) generateExecutiveSummary(m MonitorModel, diagnostic DiagnosticModel, securityChecks []security.SecurityCheck) string {
	var summary strings.Builder

	summary.WriteString("# System Health Report - Executive Summary\n\n")
	summary.WriteString(fmt.Sprintf("**Report Date/Time**: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	// Overall Status
	status := rm.calculateOverallStatus(m, diagnostic, securityChecks)
	summary.WriteString(fmt.Sprintf("**Overall Status**: %s\n", status))

	// System Health Score
	if diagnostic.Performance.HealthScore != nil {
		summary.WriteString(fmt.Sprintf("**System Health Score**: %d/100\n", diagnostic.Performance.HealthScore.Score))
	}

	// Security Score
	securityScore := rm.calculateSecurityScore(securityChecks)
	summary.WriteString(fmt.Sprintf("**Security Score**: %d/%d checks passed\n", securityScore, len(securityChecks)))

	// Critical Points
	criticalPoints := rm.getCriticalPoints(m, securityChecks)
	if len(criticalPoints) > 0 {
		summary.WriteString("\n**Critical Issues**:\n")
		for _, point := range criticalPoints {
			summary.WriteString(fmt.Sprintf("- %s\n", point))
		}
	} else {
		summary.WriteString("\n**No critical issues detected**\n")
	}

	return summary.String()
}

func (rm *ReportModel) generateSystemIdentification(m MonitorModel) string {
	var system strings.Builder

	system.WriteString("## System Identification\n\n")

	if m.System.Hostname != "" {
		system.WriteString(fmt.Sprintf("**Hostname**: %s\n", m.System.Hostname))
	}
	if m.System.OS != "" {
		system.WriteString(fmt.Sprintf("**Operating System**: %s\n", m.System.OS))
	}
	if m.System.Kernel != "" {
		system.WriteString(fmt.Sprintf("**Kernel**: %s\n", m.System.Kernel))
	}
	if m.System.Uptime > 0 {
		uptime := formatUptime(m.System.Uptime)
		system.WriteString(fmt.Sprintf("**Uptime**: %s\n", uptime))
	}

	return system.String()
}

func (rm *ReportModel) generateResourceUsage(m MonitorModel) string {
	var resources strings.Builder

	resources.WriteString("## Resource Usage\n\n")

	// CPU
	if m.Cpu.Usage > 0 {
		resources.WriteString("### CPU\n")
		resources.WriteString(fmt.Sprintf("- **Average Usage**: %.1f%%\n", m.Cpu.Usage))
		resources.WriteString(fmt.Sprintf("- **Load Average**: %.2f (1m), %.2f (5m), %.2f (15m)\n",
			m.Cpu.LoadAvg1, m.Cpu.LoadAvg5, m.Cpu.LoadAvg15))
		resources.WriteString("\n")
	}

	// Memory
	if m.Memory.Total > 0 {
		resources.WriteString("### Memory (RAM)\n")
		resources.WriteString(fmt.Sprintf("- **Usage**: %.1f%%\n", m.Memory.Usage))
		resources.WriteString(fmt.Sprintf("- **Details**: %s / %s / %s (Used/Total/Free)\n",
			formatBytes(m.Memory.Used), formatBytes(m.Memory.Total), formatBytes(m.Memory.Free)))
		resources.WriteString("\n")
	}

	// Swap
	if m.Memory.SwapTotal > 0 {
		resources.WriteString("### Swap\n")
		resources.WriteString(fmt.Sprintf("- **Usage**: %.1f%%\n", m.Memory.SwapUsage))
		resources.WriteString(fmt.Sprintf("- **Details**: %s / %s / %s (Used/Total/Free)\n",
			formatBytes(m.Memory.SwapUsed), formatBytes(m.Memory.SwapTotal), formatBytes(m.Memory.SwapFree)))
		resources.WriteString("\n")
	}

	// Disks
	if len(m.Disks) > 0 {
		resources.WriteString("### Disk Usage\n")
		resources.WriteString("| Mount Point | Usage (%) | Total | Used |\n")
		resources.WriteString("| :---------- | :-------- | :---- | :--- |\n")
		for _, disk := range m.Disks {
			resources.WriteString(fmt.Sprintf("| %s | %.1f%% | %s | %s |\n",
				disk.Mountpoint, disk.Usage, formatBytes(disk.Total), formatBytes(disk.Used)))
		}
	}

	return resources.String()
}

func (rm *ReportModel) generateSecurityStatus(securityChecks []security.SecurityCheck) string {
	var security strings.Builder

	security.WriteString("## Security Status\n\n")
	security.WriteString("| Check | Status | Details |\n")
	security.WriteString("| :---- | :----- | :------ |\n")

	for _, check := range securityChecks {
		status := "❌ Failed"
		if strings.Contains(check.Status, "OK") || strings.Contains(check.Status, "Valide") {
			status = "✅ Passed"
		} else if strings.Contains(check.Status, "Warning") {
			status = "⚠️ Warning"
		}

		security.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			check.Name, status, check.Details))
	}

	return security.String()
}

func (rm *ReportModel) generateContainerStatus(m MonitorModel) string {
	var containers strings.Builder

	containers.WriteString("## Docker Container Status\n\n")

	if m.App != nil {
		containerList, _ := m.App.RefreshContainers()
		totalContainers := len(containerList)

		containers.WriteString(fmt.Sprintf("**Total Containers**: %d\n", totalContainers))

		// Count by status
		running := 0
		stopped := 0
		paused := 0
		unhealthy := []string{}

		for _, container := range containerList {
			switch container.State {
			case "running":
				running++
			case "exited", "stopped":
				stopped++
			case "paused":
				paused++
			}

			if container.Health == "unhealthy" {
				unhealthy = append(unhealthy, container.Name)
			}
		}

		containers.WriteString(fmt.Sprintf("- **Running**: %d\n", running))
		containers.WriteString(fmt.Sprintf("- **Stopped**: %d\n", stopped))
		containers.WriteString(fmt.Sprintf("- **Paused**: %d\n", paused))

		if len(unhealthy) > 0 {
			containers.WriteString("\n**Unhealthy Containers**:\n")
			for _, name := range unhealthy {
				containers.WriteString(fmt.Sprintf("- %s\n", name))
			}
		} else {
			containers.WriteString("\n**No unhealthy containers detected**\n")
		}
	} else {
		containers.WriteString("**Docker not available**\n")
	}

	return containers.String()
}

func (rm *ReportModel) generatePerformanceInsights(m MonitorModel, diagnostic DiagnosticModel) string {
	var performance strings.Builder

	performance.WriteString("## Performance Insights\n\n")

	// Top processes by CPU
	if len(m.Processes) > 0 {
		performance.WriteString("### Top Processes by CPU Usage\n")
		sort.Slice(m.Processes, func(i, j int) bool {
			return m.Processes[i].CPU > m.Processes[j].CPU
		})

		for i := 0; i < len(m.Processes) && i < 5; i++ {
			p := m.Processes[i]
			performance.WriteString(fmt.Sprintf("- **%s** (PID: %d): %.1f%% CPU\n",
				p.Command, p.PID, p.CPU))
		}
		performance.WriteString("\n")
	}

	// Top processes by Memory
	if len(m.Processes) > 0 {
		performance.WriteString("### Top Processes by Memory Usage\n")
		sort.Slice(m.Processes, func(i, j int) bool {
			return m.Processes[i].Mem > m.Processes[j].Mem
		})

		for i := 0; i < len(m.Processes) && i < 5; i++ {
			p := m.Processes[i]
			performance.WriteString(fmt.Sprintf("- **%s** (PID: %d): %.1f%% Memory\n",
				p.Command, p.PID, p.Mem))
		}
		performance.WriteString("\n")
	}

	// System health metrics
	if diagnostic.Performance.HealthMetrics != nil {
		performance.WriteString("### System Health Metrics\n")
		metrics := diagnostic.Performance.HealthMetrics

		performance.WriteString(fmt.Sprintf("- **I/O Wait**: %.2f%%\n", metrics.IOWait))
		performance.WriteString(fmt.Sprintf("- **Steal Time**: %.2f%%\n", metrics.StealTime))
		performance.WriteString(fmt.Sprintf("- **Major Page Faults**: %s\n", formatNumber(metrics.MajorFaults)))
		performance.WriteString(fmt.Sprintf("- **Context Switches**: %s\n", formatNumber(metrics.ContextSwitches)))
	}

	return performance.String()
}

func (rm *ReportModel) generateRecommendations(m MonitorModel, securityChecks []security.SecurityCheck) string {
	var recommendations strings.Builder

	recommendations.WriteString("## Recommendations\n\n")

	// Security recommendations
	securityRecs := rm.getSecurityRecommendations(securityChecks)
	if len(securityRecs) > 0 {
		recommendations.WriteString("### Security\n")
		for _, rec := range securityRecs {
			recommendations.WriteString(fmt.Sprintf("- %s\n", rec))
		}
		recommendations.WriteString("\n")
	}

	// Performance recommendations
	performanceRecs := rm.getPerformanceRecommendations(m)
	if len(performanceRecs) > 0 {
		recommendations.WriteString("### Performance\n")
		for _, rec := range performanceRecs {
			recommendations.WriteString(fmt.Sprintf("- %s\n", rec))
		}
		recommendations.WriteString("\n")
	}

	// Docker recommendations
	dockerRecs := rm.getDockerRecommendations(m)
	if len(dockerRecs) > 0 {
		recommendations.WriteString("### Docker\n")
		for _, rec := range dockerRecs {
			recommendations.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	if !strings.Contains(recommendations.String(), "###") {
		recommendations.WriteString("**No specific recommendations at this time.**\n")
	}

	return recommendations.String()
}

func (rm *ReportModel) calculateOverallStatus(m MonitorModel, diagnostic DiagnosticModel, securityChecks []security.SecurityCheck) ReportStatus {
	// Check for critical conditions
	if m.Memory.Usage > 95 || (diagnostic.Performance.HealthScore != nil && diagnostic.Performance.HealthScore.Score < 50) {
		return StatusCritical
	}

	// Check for warning conditions
	if m.Memory.Usage > 85 || m.Cpu.Usage > 85 ||
		(diagnostic.Performance.HealthScore != nil && diagnostic.Performance.HealthScore.Score < 70) {
		return StatusWarning
	}

	// Check security status
	for _, check := range securityChecks {
		if strings.Contains(check.Status, "Failed") || strings.Contains(check.Status, "Critical") {
			return StatusWarning
		}
	}

	return StatusHealthy
}

func (rm *ReportModel) calculateSecurityScore(securityChecks []security.SecurityCheck) int {
	passed := 0
	for _, check := range securityChecks {
		if strings.Contains(check.Status, "OK") || strings.Contains(check.Status, "Valide") {
			passed++
		}
	}
	return passed
}

func (rm *ReportModel) getCriticalPoints(m MonitorModel, securityChecks []security.SecurityCheck) []string {
	var points []string

	// Memory usage
	if m.Memory.Usage > 95 {
		points = append(points, fmt.Sprintf("Memory usage at %.1f%% - risk of system instability", m.Memory.Usage))
	}

	// CPU usage
	if m.Cpu.Usage > 95 {
		points = append(points, fmt.Sprintf("CPU usage at %.1f%% - system may be overloaded", m.Cpu.Usage))
	}

	// Security issues
	for _, check := range securityChecks {
		if strings.Contains(check.Status, "Failed") || strings.Contains(check.Status, "Critical") {
			points = append(points, fmt.Sprintf("Security: %s - %s", check.Name, check.Details))
		}
	}

	// Docker unhealthy containers
	if m.App != nil {
		containers, _ := m.App.RefreshContainers()
		for _, container := range containers {
			if container.Health == "unhealthy" {
				points = append(points, fmt.Sprintf("Docker container '%s' is unhealthy", container.Name))
			}
		}
	}

	return points
}

func (rm *ReportModel) getSecurityRecommendations(securityChecks []security.SecurityCheck) []string {
	var recommendations []string

	for _, check := range securityChecks {
		if strings.Contains(check.Status, "Warning") {
			switch check.Name {
			case "SSH Password Authentication":
				recommendations = append(recommendations, "SSH password authentication is enabled. Consider using SSH keys only for better security.")
			case "System Updates":
				recommendations = append(recommendations, "System updates are available. Run 'sudo apt update && sudo apt upgrade' to apply security patches.")
			case "System Restart Required":
				recommendations = append(recommendations, "System restart required for kernel updates. Schedule a maintenance window to reboot the system.")
			}
		} else if strings.Contains(check.Status, "Failed") {
			switch check.Name {
			case "Firewall Status":
				recommendations = append(recommendations, "Firewall is not active. Enable it with 'sudo ufw enable'.")
			case "Open Ports":
				recommendations = append(recommendations, "Risky ports are open. Close unnecessary ports to reduce attack surface.")
			}
		}
	}

	return recommendations
}

func (rm *ReportModel) getPerformanceRecommendations(m MonitorModel) []string {
	var recommendations []string

	// Memory recommendations
	if m.Memory.Usage > 90 {
		if len(m.Processes) > 0 {
			topProcess := m.Processes[0]
			recommendations = append(recommendations,
				fmt.Sprintf("Memory usage very high (%.1f%%). Process '%s' is consuming %.1f%% memory. Consider optimizing or allocating more resources.",
					m.Memory.Usage, topProcess.Command, topProcess.Mem))
		} else {
			recommendations = append(recommendations,
				fmt.Sprintf("Memory usage very high (%.1f%%). Consider adding more RAM or optimizing memory usage.", m.Memory.Usage))
		}
	}

	// CPU recommendations
	if m.Cpu.Usage > 90 {
		if len(m.Processes) > 0 {
			topProcess := m.Processes[0]
			recommendations = append(recommendations,
				fmt.Sprintf("CPU usage very high (%.1f%%). Process '%s' is consuming %.1f%% CPU. Consider optimizing or distributing load.",
					m.Cpu.Usage, topProcess.Command, topProcess.CPU))
		} else {
			recommendations = append(recommendations,
				fmt.Sprintf("CPU usage very high (%.1f%%). Consider upgrading CPU or optimizing processes.", m.Cpu.Usage))
		}
	}

	// Disk recommendations
	for _, disk := range m.Disks {
		if disk.Usage > 90 {
			recommendations = append(recommendations,
				fmt.Sprintf("Disk usage on '%s' is very high (%.1f%%). Consider cleaning up or expanding storage.",
					disk.Mountpoint, disk.Usage))
		}
	}

	return recommendations
}

func (rm *ReportModel) getDockerRecommendations(m MonitorModel) []string {
	var recommendations []string

	if m.App != nil {
		containers, _ := m.App.RefreshContainers()
		for _, container := range containers {
			if container.Health == "unhealthy" {
				recommendations = append(recommendations,
					fmt.Sprintf("Container '%s' is unhealthy. Check its logs with 'docker logs %s'.",
						container.Name, container.ID[:12]))
			}
		}
	}

	return recommendations
}

func (rm *ReportModel) SaveReport() (string, error) {
	if rm.CurrentReport == "" {
		return "", fmt.Errorf("no report to save")
	}

	filename := fmt.Sprintf("report_%s_%s.md",
		time.Now().Format("2006-01-02_150405"),
		"system") // In practice, you might want to get the actual hostname

	filepath := filepath.Join(rm.ReportDirectory, filename)
	err := os.WriteFile(filepath, []byte(rm.CurrentReport), 0644)
	if err != nil {
		return "", err
	}

	// Update saved reports list
	rm.RefreshSavedReports()

	return filepath, nil
}

func (rm *ReportModel) HandleClearSaveNotification() {
	if rm.SaveNotification != "" && time.Since(rm.SaveNotificationTime) >= (2900*time.Millisecond) {
		rm.SaveNotification = ""
		rm.SaveNotificationTime = time.Time{}
	}
}

func (rm *ReportModel) RefreshSavedReports() {
	files, err := os.ReadDir(rm.ReportDirectory)
	if err != nil {
		return
	}

	rm.SavedReports = []string{}
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "report_") {
			rm.SavedReports = append(rm.SavedReports, file.Name())
		}
	}

	// Sort by name (which includes timestamp) in descending order
	sort.Sort(sort.Reverse(sort.StringSlice(rm.SavedReports)))
}

func (rm *ReportModel) LoadReport(filename string) (string, error) {
	filepath := filepath.Join(rm.ReportDirectory, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	rm.CurrentReport = string(content)
	return rm.CurrentReport, nil
}

func formatUptime(uptime uint64) string {
	seconds := int(uptime)
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(n uint64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) < 4 {
		return s
	}
	var result []string
	for i := len(s); i > 0; i -= 3 {
		start := max(i-3, 0)
		result = append([]string{s[start:i]}, result...)
	}
	return strings.Join(result, ",")
}
