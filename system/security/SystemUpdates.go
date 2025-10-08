package security

import (
	"fmt"
	"os/exec"
	"strings"
)

func (sm *SecurityManager) checkSystemUpdates() SecurityCheck {
	managers := sm.getPackageManagers()

	for _, mgr := range managers {
		output, err := sm.runPackageManagerCommand(mgr)

		if mgr.name == "yum" || mgr.name == "dnf" {
			return sm.handleYumDnfResult(mgr, output, err)
		} else if err == nil {
			return sm.handleAptResult(mgr, output)
		}
	}

	return SecurityCheck{
		Name:    "System Updates",
		Status:  "Unknown",
		Details: "Could not determine update status - package manager not detected",
	}
}

func (sm *SecurityManager) getPackageManagers() []struct {
	name      string
	command   []string
	parseFunc func(string) (int, error)
} {
	return []struct {
		name      string
		command   []string
		parseFunc func(string) (int, error)
	}{
		{
			name:    "apt",
			command: []string{"apt", "list", "--upgradable"},
			parseFunc: func(output string) (int, error) {
				lines := strings.Split(output, "\n")
				count := 0
				for _, line := range lines {
					if strings.Contains(line, "upgradable") {
						count++
					}
				}
				return count, nil
			},
		},
		{
			name:    "yum",
			command: []string{"yum", "check-update", "--quiet"},
			parseFunc: func(output string) (int, error) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) == 1 && lines[0] == "" {
					return 0, nil
				}
				return len(lines), nil
			},
		},
		{
			name:    "dnf",
			command: []string{"dnf", "check-update", "--quiet"},
			parseFunc: func(output string) (int, error) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) == 1 && lines[0] == "" {
					return 0, nil
				}
				return len(lines), nil
			},
		},
	}
}

func (sm *SecurityManager) runPackageManagerCommand(mgr struct {
	name      string
	command   []string
	parseFunc func(string) (int, error)
}) ([]byte, error) {
	var cmd *exec.Cmd

	// Use sudo if authenticated and not running as root
	if sm.CanUseSudo && !sm.IsRoot {
		args := append([]string{"-S"}, mgr.command...)
		cmd = exec.Command("sudo", args...)
		if sm.SudoPassword != "" {
			cmd.Stdin = strings.NewReader(sm.SudoPassword + "\n")
		}
	} else {
		cmd = exec.Command(mgr.command[0], mgr.command[1:]...)
	}

	return cmd.Output()
}

func (sm *SecurityManager) handleYumDnfResult(mgr struct {
	name      string
	command   []string
	parseFunc func(string) (int, error)
}, output []byte, err error) SecurityCheck {
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 100 {
		count, _ := mgr.parseFunc(string(output))
		return SecurityCheck{
			Name:    "System Updates",
			Status:  "Updates Available",
			Details: fmt.Sprintf("%d updates available via %s", count, mgr.name),
		}
	} else if err == nil {
		return SecurityCheck{
			Name:    "System Updates",
			Status:  "Up to date",
			Details: fmt.Sprintf("All packages are up to date (%s)", mgr.name),
		}
	}

	// If we get here, there was an error but not exit code 100
	// Continue to next package manager
	return SecurityCheck{}
}

func (sm *SecurityManager) handleAptResult(mgr struct {
	name      string
	command   []string
	parseFunc func(string) (int, error)
}, output []byte) SecurityCheck {
	count, _ := mgr.parseFunc(string(output))
	if count > 0 {
		return SecurityCheck{
			Name:    "System Updates",
			Status:  "Updates Available",
			Details: fmt.Sprintf("%d updates available via %s", count, mgr.name),
		}
	}
	return SecurityCheck{
		Name:    "System Updates",
		Status:  "Up to date",
		Details: fmt.Sprintf("All packages are up to date (%s)", mgr.name),
	}
}
