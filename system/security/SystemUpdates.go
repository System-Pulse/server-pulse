package security

import (
	"fmt"
	"os/exec"
	"strings"
)

func (sm *SecurityManager) checkSystemUpdates() SecurityCheck {

	managers := []struct {
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

	for _, mgr := range managers {
		cmd := exec.Command(mgr.command[0], mgr.command[1:]...)
		output, err := cmd.Output()

		if mgr.name == "yum" || mgr.name == "dnf" {
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
		} else if err == nil {
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
	}

	return SecurityCheck{
		Name:    "System Updates",
		Status:  "Unknown",
		Details: "Could not determine update status - package manager not detected",
	}
}
