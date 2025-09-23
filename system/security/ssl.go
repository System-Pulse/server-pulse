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

type CertificateDisplayMsg CertificateInfos

func (sm *SecurityManager) checkSSLCertificate() SecurityCheck {

	// cmd := exec.Command("curl", "-s", "https://api.ipify.org")
	// pubIp, err := cmd.Output()
	// if err != nil {
	// 	return SecurityCheck{
	// 		Name:    "SSL Certificate",
	// 		Status:  "Error",
	// 		Details: fmt.Sprintf("Failed to get public IP: %v", err),
	// 	}
	// }

	// names, err := exec.Command("dig", "+short", "-x", string(pubIp)).Output()
	// if err != nil {
	// 	return SecurityCheck{
	// 		Name:    "SSL Certificate",
	// 		Status:  "Error",
	// 		Details: fmt.Sprintf("Failed to get domain name: %v", err),
	// 	}
	// }

	// domaine := strings.TrimSpace(string(names))
	// domaine = strings.TrimSuffix(domaine, ".")
	domaine := "arndofficiel.com"

	port := "443"
	conn, err := tls.Dial("tcp", domaine+":"+port, &tls.Config{
		ServerName: domaine,
	})
	if err != nil {
		return SecurityCheck{
			Name:    "SSL Certificate",
			Status:  "Error",
			Details: fmt.Sprintf("Failed to connect to %v: %v", domaine, err),
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
	if sm.Certificate == nil {
		return nil
	}

	return func() tea.Msg {
		info := DisplayCertificateInfo(sm.Certificate, sm.Hostname)
		return CertificateDisplayMsg(info)
	}
}
