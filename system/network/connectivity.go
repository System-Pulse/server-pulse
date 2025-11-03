package network

import (
	"fmt"
	"os/exec"
	"strconv"
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
		cmd := exec.Command("ping", "-c", strconv.Itoa(count), "-W", "5", target)
		output, err := cmd.CombinedOutput()

		if err != nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "bytes from") {
				return parseSystemPingOutput(target, outputStr)
			}
			return PingMsg(PingResult{
				Target:  target,
				Success: false,
				Error:   fmt.Sprintf("Ping failed: %v", err),
			})
		}

		return parseSystemPingOutput(target, string(output))
	}
}

func parseSystemPingOutput(target, output string) PingMsg {
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		if strings.Contains(line, "bytes from") {
			if strings.Contains(line, "time=") {
				timePart := strings.Split(line, "time=")[1]
				timePart = strings.Split(timePart, " ")[0]
				if latency, err := time.ParseDuration(timePart); err == nil {
					return PingMsg{
						Target:  target,
						Success: true,
						Latency: latency,
					}
				}
			}
			return PingMsg{
				Target:  target,
				Success: true,
				Latency: time.Millisecond * 10,
			}
		}
	}

	return PingMsg{
		Target:  target,
		Success: false,
		Error:   "No response from target",
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
		outputStr = strings.ReplaceAll(outputStr, "*", "")
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
