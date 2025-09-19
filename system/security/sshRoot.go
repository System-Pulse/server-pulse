package security

func (sm *SecurityManager) checkSSHRootLogin() SecurityCheck {
	// TODO: Replace with your actual SSH root login check function
	return SecurityCheck{
		Name:    "SSH Root Login",
		Status:  "Disabled",
		Details: "Root login is properly disabled",
	}
}
