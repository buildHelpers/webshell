package agent

import "testing"

func TestValidateDomainUpdate(t *testing.T) {
	if err := validateDomainUpdate(controlPlaneUpdate{Domain: "api.example.com", Port: 443}); err != nil {
		t.Fatalf("expected domain to be valid: %v", err)
	}
	if err := validateDomainUpdate(controlPlaneUpdate{Domain: "https://example.com", Port: 443}); err == nil {
		t.Fatal("expected URL to be rejected")
	}
	if err := validateDomainUpdate(controlPlaneUpdate{Domain: "api.example.com", Port: 0}); err == nil {
		t.Fatal("expected port to be rejected")
	}
}
