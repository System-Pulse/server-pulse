package network

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		cmd := exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", "5", target)

		output, err := cmd.Output()

		if err != nil {
			// Check if it's a timeout or unreachable
			outputStr := string(output)
			if strings.Contains(outputStr, "100% packet loss") {
				return PingMsg(PingResult{
					Target:     target,
					Success:    false,
					PacketLoss: 100.0,
					Error:      "Destination unreachable",
				})
			}
			return PingMsg(PingResult{
				Target:  target,
				Success: false,
				Error:   fmt.Sprintf("ping failed: %v", err),
			})
		}

		// Parse ping output
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")

		var latency time.Duration
		var packetLoss float64

		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Parse latency (e.g., "64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=12.3 ms")
			if strings.Contains(line, "time=") {
				timeStart := strings.Index(line, "time=")
				if timeStart != -1 {
					timePart := line[timeStart+5:]
					timeEnd := strings.Index(timePart, " ")
					if timeEnd != -1 {
						timeStr := timePart[:timeEnd]
						if parsed, err := time.ParseDuration(timeStr + "ms"); err == nil {
							latency = parsed
						}
					}
				}
			}

			// Parse packet loss (e.g., "3 packets transmitted, 3 received, 0% packet loss")
			if strings.Contains(line, "packet loss") {
				lossStart := strings.Index(line, "received,")
				if lossStart != -1 {
					lossPart := line[lossStart+9:]
					lossEnd := strings.Index(lossPart, "%")
					if lossEnd != -1 {
						lossStr := strings.TrimSpace(lossPart[:lossEnd])
						fmt.Sscanf(lossStr, "%f", &packetLoss)
					}
				}
			}
		}

		return PingMsg(PingResult{
			Target:     target,
			Success:    true,
			Latency:    latency,
			PacketLoss: packetLoss,
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
