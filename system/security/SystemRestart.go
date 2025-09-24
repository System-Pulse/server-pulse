package security

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func (sm *SecurityManager) checkSystemRestart() SecurityCheck {
	checks := []func() (bool, string){
		checkRebootRequired,
		checkKernelUpdate,
		checkSystemdRestart,
	}

	for _, check := range checks {
		if required, reason := check(); required {
			return SecurityCheck{
				Name:    "System Restart",
				Status:  "Required",
				Details: reason,
			}
		}
	}

	return SecurityCheck{
		Name:    "System Restart",
		Status:  "Not Required",
		Details: "No pending system restart required",
	}
}

func checkRebootRequired() (bool, string) {
	if _, err := os.Stat("/var/run/reboot-required"); err == nil {
		if content, err := os.ReadFile("/var/run/reboot-required.pkgs"); err == nil {
			packages := strings.TrimSpace(string(content))
			return true, fmt.Sprintf("Reboot required due to package updates: %s", packages)
		}
		return true, "Reboot required flag detected"
	}
	return false, ""
}

func checkKernelUpdate() (bool, string) {

	cmd := exec.Command("uname", "-r")
	currentKernel, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	currentVersion := strings.TrimSpace(string(currentKernel))

	cmd = exec.Command("ls", "/boot")
	bootFiles, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	bootFilesList := string(bootFiles)
	newerKernelPattern := regexp.MustCompile(`vmlinuz-(\d+\.\d+\.\d+-\d+)`)
	matches := newerKernelPattern.FindAllStringSubmatch(bootFilesList, -1)

	for _, match := range matches {
		if len(match) > 1 && match[1] != currentVersion {
			return true, fmt.Sprintf("Newer kernel available: %s (current: %s)", match[1], currentVersion)
		}
	}

	return false, ""
}

func checkSystemdRestart() (bool, string) {

	cmd := exec.Command("systemctl", "list-units", "--state=failed", "--no-pager")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		failedServices := 0
		for _, line := range lines {
			if strings.Contains(line, "failed") && strings.Contains(line, ".service") {
				failedServices++
			}
		}
		if failedServices > 0 {
			return true, fmt.Sprintf("%d failed services detected - restart may be needed", failedServices)
		}
	}

	return false, ""
}
