package agent

import "testing"

func TestValidateDomainUpdateAcceptsDNSHostname(t *testing.T) {
	if err := validateDomainUpdate(controlPlaneUpdate{Domain: "panel.example.com", Port: 443}); err != nil {
		t.Fatalf("expected valid hostname: %v", err)
	}
}

func TestValidateDomainUpdateRejectsURL(t *testing.T) {
	if err := validateDomainUpdate(controlPlaneUpdate{Domain: "https://panel.example.com", Port: 443}); err == nil {
		t.Fatal("expected URL to be rejected")
	}
}
