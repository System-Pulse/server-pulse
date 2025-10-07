package utils

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ErrMsg error
type InfoMsg string
type TickMsg time.Time

func FormatPercentage(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

func GetUsageIcon(percent float64) string {
	switch {
	case percent < 25:
		return "🟢"
	case percent < 50:
		return "🟡"
	case percent < 75:
		return "🟠"
	default:
		return "🔴"
	}
}

func FormatCompactUptime(seconds uint64) string {
	days := seconds / (60 * 60 * 24)
	hours := (seconds % (60 * 60 * 24)) / (60 * 60)
	minutes := (seconds % (60 * 60)) / 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatCompactBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatUptime(seconds uint64) string {
	days := seconds / (60 * 60 * 24)
	hours := (seconds % (60 * 60 * 24)) / (60 * 60)
	minutes := (seconds % (60 * 60)) / 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

func Ellipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func LoadAvg() ([3]float64, error) {
	var loadAvg [3]float64

	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return loadAvg, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil {
		return loadAvg, err
	}

	fields := strings.Fields(line)
	if len(fields) < 3 {
		return loadAvg, fmt.Errorf("format incorrect: moins de 3 valeurs disponibles")
	}

	for i := range 3 {
		loadAvg[i], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return loadAvg, fmt.Errorf("erreur de conversion: %v", err)
		}
	}

	return loadAvg, nil
}

func CheckDockerPermissions() (bool, string) {
	// Step 1: Check if the 'docker' command exists in the user's PATH (cross-platform)
	_, err := exec.LookPath("docker")
	if err != nil {
		return false, "The 'docker' command was not found. Please ensure Docker is installed."
	}

	// Step 2: Platform-specific permission checks
	switch runtime.GOOS {
	case "linux": // Linux
		currentUser, err := user.Current()
		if err != nil {
			return false, fmt.Sprintf("Error getting the current user: %v", err)
		}

		groups, err := currentUser.GroupIds()
		if err != nil {
			return false, fmt.Sprintf("Error getting user's groups: %v", err)
		}

		for _, groupID := range groups {
			group, err := user.LookupGroupId(groupID)
			if err != nil {
				continue
			}
			if group.Name == "docker" || group.Name == "root" {
				return true, "The user has permissions to run Docker."
			}
		}

		return false, "The user is not in the 'docker' group. To add them, run 'sudo usermod -aG docker <your_username>' and then log out and log back in."

	default:
		return false, fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS)
	}
}

func FormatOperationMessage(operation string, success bool, err error) string {
	operationLabels := map[string]string{
		"restart":      "Container restarted",
		"start":        "Container started",
		"stop":         "Container stopped",
		"pause":        "Container paused",
		"unpause":      "Container resumed",
		"delete":       "Container deleted",
		"toggle_start": "Container state changed",
		"toggle_pause": "Container pause state changed",
		"exec":         "Shell opened",
		"logs":         "Logs loaded",
	}

	label, exists := operationLabels[operation]
	if !exists {
		label = fmt.Sprintf("Operation '%s'", operation)
	}

	if success {
		return fmt.Sprintf("✅ %s successfully", label)
	} else {
		if err != nil {
			return fmt.Sprintf("❌ %s failed: %v", label, err)
		}
		return fmt.Sprintf("❌ %s failed", label)
	}
}

func GetOperationIcon(operation string) string {
	icons := map[string]string{
		"restart":      "🔄",
		"start":        "▶️",
		"stop":         "⏹️",
		"pause":        "⏸️",
		"unpause":      "▶️",
		"delete":       "🗑️",
		"toggle_start": "🔄",
		"toggle_pause": "⏯️",
		"exec":         "💻",
		"logs":         "📄",
	}

	if icon, exists := icons[operation]; exists {
		return icon
	}
	return "⚙️"
}

// hexToIPv4 convert hex (little-endian) to IPv4
func HexToIPv4(hexStr string) string {
	if len(hexStr) != 8 {
		return "0.0.0.0"
	}

	// hex to uint32 (little-endian)
	var ipInt uint32
	_, err := fmt.Sscanf(hexStr, "%x", &ipInt)
	if err != nil {
		return "0.0.0.0"
	}

	
	ipBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(ipBytes, ipInt)


	ip := net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])

	if ip.Equal(net.IPv4zero) {
		return "0.0.0.0"
	}

	return ip.String()
}

func ParseRouteFlags(flagsStr string) string {
	var flags uint32
	_, err := fmt.Sscanf(flagsStr, "%x", &flags)
	if err != nil {
		return ""
	}

	const (
		RTF_UP        = 0x0001 
		RTF_GATEWAY   = 0x0002 
		RTF_HOST      = 0x0004 
		RTF_REINSTATE = 0x0008 
		RTF_DYNAMIC   = 0x0010 
		RTF_MODIFIED  = 0x0020 
		RTF_REJECT    = 0x0200 
	)

	var flagChars []byte

	if flags&RTF_UP != 0 {
		flagChars = append(flagChars, 'U')
	}
	if flags&RTF_GATEWAY != 0 {
		flagChars = append(flagChars, 'G')
	}
	if flags&RTF_HOST != 0 {
		flagChars = append(flagChars, 'H')
	}
	if flags&RTF_REINSTATE != 0 {
		flagChars = append(flagChars, 'R')
	}
	if flags&RTF_DYNAMIC != 0 {
		flagChars = append(flagChars, 'D')
	}
	if flags&RTF_MODIFIED != 0 {
		flagChars = append(flagChars, 'M')
	}
	if flags&RTF_REJECT != 0 {
		flagChars = append(flagChars, '!')
	}

	return string(flagChars)
}

func HexToIPv6(hexStr string) (string, error) {
	if len(hexStr) != 32 {
		return "", fmt.Errorf("invalid IPv6 hex string length")
	}

	var ipBytes []byte
	for i := 0; i < 32; i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return "", err
		}
		ipBytes = append(ipBytes, b)
	}

	ip := net.IP(ipBytes)
	return ip.String(), nil
}

func ParseIPv6RouteFlags(flagsStr string) string {
	var flags uint32
	_, err := fmt.Sscanf(flagsStr, "%x", &flags)
	if err != nil {
		return ""
	}

	const (
		RTF_UP        = 0x0001
		RTF_GATEWAY   = 0x0002
		RTF_HOST      = 0x0004
		RTF_REJECT    = 0x0200
		RTF_DYNAMIC   = 0x0010
		RTF_MODIFIED  = 0x0020
		RTF_DEFAULT   = 0x10000
	)

	var flagChars []byte
	
	if flags&RTF_UP != 0 {
		flagChars = append(flagChars, 'U')
	}
	if flags&RTF_GATEWAY != 0 {
		flagChars = append(flagChars, 'G')
	}
	if flags&RTF_HOST != 0 {
		flagChars = append(flagChars, 'H')
	}
	if flags&RTF_REJECT != 0 {
		flagChars = append(flagChars, '!')
	}
	if flags&RTF_DYNAMIC != 0 {
		flagChars = append(flagChars, 'D')
	}
	if flags&RTF_MODIFIED != 0 {
		flagChars = append(flagChars, 'M')
	}
	if flags&RTF_DEFAULT != 0 {
		flagChars = append(flagChars, 'd')
	}

	return string(flagChars)
}