package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SecurityManager struct {
	Certificate *x509.Certificate
	Hostname    string
}

func NewSecurityManager() *SecurityManager {
	return &SecurityManager{}
}

type SecurityCheck struct {
	Name    string
	Status  string
	Details string
}

type SecurityCheckResult struct {
	Checks     []SecurityCheck
	LastUpdate time.Time
}

type CertificateInfos struct {
	Subject            string
	Issuer             string
	SerialNumber       string
	Version            int
	ValidityPeriodFrom string
	ValidityPeriodTo   string
	DaysUntilExpiry    int
	HostnameVerified   bool
	AlternativeNames   []string
	Algorithm          string
	SignatureAlgorithm string
}

type SecurityMsg []SecurityCheck

type CertificateDisplayMsg CertificateInfos

func (sm *SecurityManager) RunSecurityChecks() tea.Cmd {
	return func() tea.Msg {
		checks := []SecurityCheck{
			sm.checkSSLCertificate(),
			sm.checkSSHRootLogin(),
			sm.checkOpenPorts(),
			sm.checkFirewallStatus(),
			sm.checkSystemUpdates(),
		}

		return SecurityMsg(checks)
	}
}

func (sm *SecurityManager) checkSSLCertificate() SecurityCheck {
	//TODO: find the domaine name linked to ip address of the server and replace it with the correct domain name

	domaine := "www.arndofficiel.com"
	port := "443"
	conn, err := tls.Dial("tcp", domaine+":"+port, &tls.Config{
		ServerName: domaine,
	})
	if err != nil {
		return SecurityCheck{
			Name:    "SSL Certificate",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to connect to server: %v", err),
		}
	}
	defer conn.Close()

	state := conn.ConnectionState()
	now := time.Now()

	if len(state.PeerCertificates) == 0 {
		return SecurityCheck{
			Name:    "SSL Certificate",
			Status:  "Invalid",
			Details: "No SSL certificate found",
		}
	}

	cert := state.PeerCertificates[0]
	// Store certificate and hostname for detailed display
	sm.Certificate = cert
	sm.Hostname = domaine

	if cert.NotAfter.Before(now) {
		return SecurityCheck{
			Name:    "SSL Certificate",
			Status:  "Invalid",
			Details: "Certificate has expired",
		}
	}

	if cert.NotAfter.Before(now.Add(30 * 24 * time.Hour)) {
		daysUntilExpiration := cert.NotAfter.Sub(now).Hours() / 24
		return SecurityCheck{
			Name:    "SSL Certificate",
			Status:  "Warning",
			Details: fmt.Sprintf("Certificate expires soon in %.0f days", daysUntilExpiration),
		}
	}

	daysUntilExpiration := time.Until(cert.NotAfter).Hours() / 24
	return SecurityCheck{
		Name:    "SSL Certificate",
		Status:  "Valid",
		Details: fmt.Sprintf("Certificate expires in %.0f days", daysUntilExpiration),
	}
}

func DisplayCertificateInfo(cert *x509.Certificate, hostname string) CertificateInfos {
	valideHost := true
	if err := cert.VerifyHostname(hostname); err != nil {
		valideHost = false
	}

	otherNames := cert.DNSNames
	if len(otherNames) == 0 {
		otherNames = []string{"No alternative names"}
	}
	return CertificateInfos{
		Subject:            cert.Subject.CommonName,
		Issuer:             cert.Issuer.CommonName,
		SerialNumber:       cert.SerialNumber.String(),
		Version:            cert.Version,
		ValidityPeriodFrom: cert.NotBefore.Format("2006-01-02 15:04:05 MST"),
		ValidityPeriodTo:   cert.NotAfter.Format("2006-01-02 15:04:05 MST"),
		DaysUntilExpiry:    int(time.Until(cert.NotAfter).Hours() / 24),
		HostnameVerified:   valideHost,
		AlternativeNames:   otherNames,
		Algorithm:          cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
	}
}

func (sm *SecurityManager) RunCertificateDisplay() tea.Cmd {
	return func() tea.Msg {
		if sm.Certificate == nil {
			// If no certificate is available, return an error or run SSL check first
			return SecurityMsg([]SecurityCheck{{
				Name:    "Certificate Display",
				Status:  "Error",
				Details: "No certificate available. Run security checks first.",
			}})
		}
		info := DisplayCertificateInfo(sm.Certificate, sm.Hostname)
		return CertificateDisplayMsg(info)
	}
}

func (sm *SecurityManager) checkSSHRootLogin() SecurityCheck {
	// TODO: Replace with your actual SSH root login check function
	return SecurityCheck{
		Name:    "SSH Root Login",
		Status:  "Disabled",
		Details: "Root login is properly disabled",
	}
}

func (sm *SecurityManager) checkOpenPorts() SecurityCheck {
	// TODO: Replace with your actual open ports check function
	return SecurityCheck{
		Name:    "Open Ports",
		Status:  "Secure",
		Details: "Only necessary ports are open (22, 80, 443)",
	}
}

func (sm *SecurityManager) checkFirewallStatus() SecurityCheck {
	// TODO: Replace with your actual firewall status check function
	return SecurityCheck{
		Name:    "Firewall Status",
		Status:  "Active",
		Details: "UFW is active and properly configured",
	}
}

func (sm *SecurityManager) checkSystemUpdates() SecurityCheck {
	// TODO: Replace with your actual system updates check function
	return SecurityCheck{
		Name:    "System Updates",
		Status:  "Up to date", // or "Updates available"
		Details: "All security patches are applied",
	}
}
