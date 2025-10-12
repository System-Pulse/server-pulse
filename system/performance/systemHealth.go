package performance

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v4/cpu"
)

func GetHealthMetrics() tea.Cmd {
	return func() tea.Msg {
		// This introduces a 1-second delay.
		cpuTimes1, err := cpu.Times(false)
		if err != nil || len(cpuTimes1) == 0 {
			return HealthMetricsMsg{nil, nil}
		}

		time.Sleep(time.Second)

		cpuTimes2, err := cpu.Times(false)
		if err != nil || len(cpuTimes2) == 0 {
			return HealthMetricsMsg{nil, nil}
		}

		totalDiff := (cpuTimes2[0].User - cpuTimes1[0].User) +
			(cpuTimes2[0].System - cpuTimes1[0].System) +
			(cpuTimes2[0].Idle - cpuTimes1[0].Idle) +
			(cpuTimes2[0].Nice - cpuTimes1[0].Nice) +
			(cpuTimes2[0].Iowait - cpuTimes1[0].Iowait) +
			(cpuTimes2[0].Irq - cpuTimes1[0].Irq) +
			(cpuTimes2[0].Softirq - cpuTimes1[0].Softirq) +
			(cpuTimes2[0].Steal - cpuTimes1[0].Steal)

		metrics := &HealthMetrics{}
		if totalDiff > 0 {
			metrics.IOWait = (cpuTimes2[0].Iowait - cpuTimes1[0].Iowait) / totalDiff * 100
			metrics.StealTime = (cpuTimes2[0].Steal - cpuTimes1[0].Steal) / totalDiff * 100
		}

		ctxt, err := getStatValue("/proc/stat", "ctxt")
		if err == nil {
			metrics.ContextSwitches = ctxt
		}

		intr, err := getStatValue("/proc/stat", "intr")
		if err == nil {
			metrics.Interrupts = intr
		}

		pgfault, err := getStatValue("/proc/vmstat", "pgfault")
		if err == nil {
			metrics.MinorFaults = pgfault
		}

		pgmajfault, err := getStatValue("/proc/vmstat", "pgmajfault")
		if err == nil {
			metrics.MajorFaults = pgmajfault
		}

		score := CalculateHealthScore(metrics)

		return HealthMetricsMsg{metrics, score}
	}
}

func formatNumber(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) < 4 {
		return s
	}
	var result []string
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		result = append([]string{s[start:i]}, result...)
	}
	return strings.Join(result, ",")
}

func getStatValue(filePath, key string) (uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return val, nil
			}
		}
	}
	return 0, fmt.Errorf("key %s not found in %s", key, filePath)
}

func CalculateHealthScore(metrics *HealthMetrics) *HealthScore {
	score := 100
	var issues []string
	var recommendations []string

	// IOWait
	if metrics.IOWait > 20 {
		score -= 30
		issues = append(issues, fmt.Sprintf("IOWait very high (%.2f%%)", metrics.IOWait))
		recommendations = append(recommendations, "Investigate disk activity, upgrade to faster storage (SSD/NVMe)")
	} else if metrics.IOWait > 10 {
		score -= 15
		issues = append(issues, fmt.Sprintf("IOWait elevated (%.2f%%)", metrics.IOWait))
		recommendations = append(recommendations, "Check for processes with high disk I/O, consider upgrading storage")
	}

	// CPU Steal Time
	if metrics.StealTime > 10 {
		score -= 20
		issues = append(issues, fmt.Sprintf("CPU Steal Time high (%.2f%%)", metrics.StealTime))
		recommendations = append(recommendations, "Check host machine load, consider migrating VM or upgrading host")
	} else if metrics.StealTime > 5 {
		score -= 10
		issues = append(issues, fmt.Sprintf("CPU Steal Time elevated (%.2f%%)", metrics.StealTime))
		recommendations = append(recommendations, "Monitor host machine performance")
	}

	// Major Page Faults
	if metrics.MajorFaults > 10000 {
		score -= 15
		issues = append(issues, fmt.Sprintf("High number of major page faults (%s)", formatNumber(metrics.MajorFaults)))
		recommendations = append(recommendations, "Investigate memory usage, consider increasing RAM")
	}

	// Context Switches
	if metrics.ContextSwitches > 100000000 { // 100 million
		score -= 10
		issues = append(issues, fmt.Sprintf("High number of context switches (%s)", formatNumber(metrics.ContextSwitches)))
		recommendations = append(recommendations, "Optimize multi-threaded applications, check for excessive context switching")
	}

	if score < 0 {
		score = 0
	}

	return &HealthScore{
		Score:           score,
		Issues:          issues,
		Recommendations: recommendations,
	}
}
