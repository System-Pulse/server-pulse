package network

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prometheus-community/pro-bing"
)

type PingResult struct {
	Target     string
	Success    bool
	Latency    time.Duration
	PacketLoss float64
	Error      string
}

type TracerouteResult struct {
	Target string
	Hops   []TracerouteHop
	Error  string
}

type TracerouteHop struct {
	HopNumber int
	IP        string
	Hostname  string
	Latency1  time.Duration
	Latency2  time.Duration
	Latency3  time.Duration
}

type PingMsg PingResult
type TracerouteMsg TracerouteResult

func Ping(target string, count int) tea.Cmd {
	return func() tea.Msg {
		pinger, err := probing.NewPinger(target)
		if err != nil {
			return PingMsg(PingResult{
				Target:  target,
				Success: false,
				Error:   fmt.Sprintf("Failed to create pinger: %v", err),
			})
		}

		pinger.Count = count
		pinger.Timeout = 5 * time.Second
		pinger.Interval = 100 * time.Millisecond
		pinger.SetPrivileged(utils.IsRoot())

		err = pinger.Run()
		if err != nil {
			return PingMsg(PingResult{
				Target:  target,
				Success: false,
				Error:   fmt.Sprintf("Ping failed: %v", err),
			})
		}

		stats := pinger.Statistics()

		return PingMsg(PingResult{
			Target:     target,
			Success:    stats.PacketsRecv > 0,
			Latency:    stats.AvgRtt,
			PacketLoss: stats.PacketLoss,
			Error:      "",
		})
	}
}

func Traceroute(target string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("traceroute", "-n", "-w", "2", "-q", "1", "-m", "30", target)

		output, err := cmd.Output()

		if err != nil {
			return TracerouteMsg(TracerouteResult{
				Target: target,
				Error:  fmt.Sprintf("traceroute failed: %v", err),
			})
		}

		var hops []TracerouteHop
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")

		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			// Skip header line
			if fields[0] == "traceroute" {
				continue
			}

			hop := TracerouteHop{HopNumber: i}

			// Parse hop number
			fmt.Sscanf(fields[0], "%d", &hop.HopNumber)

			// Parse IP addresses and latencies
			if len(fields) >= 2 {
				hop.IP = fields[1]

				// Try to resolve hostname if available
				if len(fields) >= 3 && !strings.HasPrefix(fields[2], "(") {
					hop.Hostname = fields[2]
				}

				// Parse latencies
				for j := 2; j < len(fields); j++ {
					if strings.HasSuffix(fields[j], "ms") {
						latencyStr := strings.TrimSuffix(fields[j], "ms")
						if parsed, err := time.ParseDuration(latencyStr + "ms"); err == nil {
							switch j - 2 {
							case 0:
								hop.Latency1 = parsed
							case 1:
								hop.Latency2 = parsed
							case 2:
								hop.Latency3 = parsed
							}
						}
					}
				}
			}

			hops = append(hops, hop)
		}

		return TracerouteMsg(TracerouteResult{
			Target: target,
			Hops:   hops,
		})
	}
}
