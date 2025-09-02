package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ErrMsg error
type TickMsg time.Time

// Nouvelle fonction pour formater les pourcentages avec couleurs
func FormatPercentage(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

// Fonction pour obtenir l'icône selon le niveau d'usage
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

// Version compacte du formatage des bytes
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
			if group.Name == "docker" {
				return true, "The user has permissions to run Docker."
			}
		}

		return false, "The user is not in the 'docker' group. To add them, run 'sudo usermod -aG docker <your_username>' and then log out and log back in."

	case "darwin": // macOS (Docker Desktop)
		return true, "Assuming the user has Docker permissions (Docker Desktop manages this)."

	case "windows": // Windows (Docker Desktop)
		return true, "Assuming the user has Docker permissions (Docker Desktop manages this)."

	default:
		return false, fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS)
	}
}
