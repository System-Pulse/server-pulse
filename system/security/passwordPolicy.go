package security

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func (sm *SecurityManager) checkPasswordPolicy() SecurityCheck {
	// Vérifier la configuration PAM pour les politiques de mots de passe
	pamFiles := []string{
		"/etc/pam.d/common-password",
		"/etc/pam.d/system-auth",
		"/etc/security/pwquality.conf",
	}

	var details []string
	hasPolicy := false

	for _, file := range pamFiles {
		if content, err := os.ReadFile(file); err == nil {
			contentStr := string(content)

			// Rechercher les règles de complexité
			if strings.Contains(contentStr, "pam_pwquality") ||
				strings.Contains(contentStr, "pam_cracklib") {
				hasPolicy = true
				details = append(details, fmt.Sprintf("Password policy found in %s", file))

				// Analyser les paramètres spécifiques
				if strings.Contains(contentStr, "minlen") {
					re := regexp.MustCompile(`minlen=(\d+)`)
					if matches := re.FindStringSubmatch(contentStr); len(matches) > 1 {
						details = append(details, fmt.Sprintf("Minimum length: %s", matches[1]))
					}
				}

				if strings.Contains(contentStr, "minclass") {
					re := regexp.MustCompile(`minclass=(\d+)`)
					if matches := re.FindStringSubmatch(contentStr); len(matches) > 1 {
						details = append(details, fmt.Sprintf("Minimum character classes: %s", matches[1]))
					}
				}
			}
		}
	}

	if hasPolicy {
		return SecurityCheck{
			Name:    "Password Policy",
			Status:  "Enabled",
			Details: strings.Join(details, ", "),
		}
	}

	return SecurityCheck{
		Name:    "Password Policy",
		Status:  "Disabled",
		Details: "No password complexity policy found",
	}
}
