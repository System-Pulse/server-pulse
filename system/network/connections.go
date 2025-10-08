package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func GetConnections() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("sudo", "-n", "ss", "-tunape")
		output, err := cmd.Output()
		if err != nil {
			cmd = exec.Command("ss", "-tunae")
			output, err = cmd.Output()
			if err != nil {
				return utils.ErrMsg(fmt.Errorf("failed to run ss: %w", err))
			}
		}

		var connections []ConnectionInfo
		lines := strings.Split(string(output), "\n")
		pidRegex := regexp.MustCompile(`users:\(\("([^"]+)",pid=(\d+),.*\)\)`)
		uidRegex := regexp.MustCompile(`uid:(\d+)`)

		for i, line := range lines {
			if i == 0 || strings.TrimSpace(line) == "" { // Skip header and empty lines
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}

			var state, pidInfo, uidInfo string
			proto := fields[0]

			// Parse state based on protocol
			if proto == "tcp" || proto == "tcp6" {
				if len(fields) < 6 {
					continue
				}
				state = fields[1]
			} else if proto == "udp" || proto == "udp6" {
				state = "UNCONN"
			} else {
				state = "-"
			}

			// Parse process info
			matches := pidRegex.FindStringSubmatch(line)
			if len(matches) > 2 {
				processName := matches[1]
				pid := matches[2]
				// Try to get the actual process name from /proc if possible
				if processName == "-" {
					if pidInt, err := strconv.Atoi(pid); err == nil {
						if cmdline, err := getProcessCommand(pidInt); err == nil && cmdline != "" {
							processName = cmdline
						}
					}
				}
				pidInfo = fmt.Sprintf("%s (%s)", processName, pid)
			} else {
				pidInfo = "N/A"
			}

			// Parse UID if available
			uidMatches := uidRegex.FindStringSubmatch(line)
			if len(uidMatches) > 1 {
				uidInfo = uidMatches[1]
				// Try to resolve username from UID
				if username := getUsernameFromUID(uidInfo); username != "" {
					uidInfo = username
				}
			}

			// Parse addresses with better formatting
			localAddr := fields[4]
			foreignAddr := fields[5]

			// Add UID info to PID field if available
			if uidInfo != "" && pidInfo != "N/A" {
				pidInfo = fmt.Sprintf("%s [UID:%s]", pidInfo, uidInfo)
			}

			connections = append(connections, ConnectionInfo{
				Proto:       proto,
				RecvQ:       fields[2],
				SendQ:       fields[3],
				LocalAddr:   localAddr,
				ForeignAddr: foreignAddr,
				State:       state,
				PID:         pidInfo,
			})
		}
		return ConnectionsMsg(connections)
	}
}

// Helper function to get process command from /proc
func getProcessCommand(pid int) (string, error) {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmd := exec.Command("cat", cmdlinePath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Remove null bytes and get the first part (executable name)
	cmdline := strings.ReplaceAll(string(output), "\x00", " ")
	parts := strings.Fields(cmdline)
	if len(parts) > 0 {
		// Extract just the executable name from the path
		executable := parts[0]
		if lastSlash := strings.LastIndex(executable, "/"); lastSlash != -1 {
			executable = executable[lastSlash+1:]
		}
		return executable, nil
	}
	return "", nil
}

// Helper function to get username from UID
func getUsernameFromUID(uid string) string {
	cmd := exec.Command("getent", "passwd", uid)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse the passwd entry: username:x:uid:gid:gecos:home:shell
	parts := strings.Split(strings.TrimSpace(string(output)), ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
