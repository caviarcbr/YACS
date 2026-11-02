package setup

import (
	"encoding/hex"
	"testing"
)

func TestNewSecurityConfigOCISOnly(t *testing.T) {
	config, err := NewSecurityConfig(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.OCISJWTSecret == "" {
		t.Error("OCISJWTSecret should not be empty")
	}

	if config.OnlyOfficeJWTSecret != "" {
		t.Error("OnlyOfficeJWTSecret should be empty when ONLYOFFICE is disabled")
	}

	if config.CollaborationWOPISecret != "" {
		t.Error("CollaborationWOPISecret should be empty when ONLYOFFICE is disabled")
	}
}

func TestNewSecurityConfigWithOnlyOffice(t *testing.T) {
	config, err := NewSecurityConfig(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets := []string{
		config.OCISJWTSecret,
		config.OnlyOfficeJWTSecret,
		config.CollaborationWOPISecret,
	}

	for i, secret := range secrets {
		if len(secret) != 64 {
			t.Errorf("secret %d has length %d, want 64", i, len(secret))
		}

		if _, err := hex.DecodeString(secret); err != nil {
			t.Errorf("secret %d is not valid hexadecimal: %v", i, err)
		}
	}

	if config.OCISJWTSecret == config.OnlyOfficeJWTSecret ||
		config.OCISJWTSecret == config.CollaborationWOPISecret ||
		config.OnlyOfficeJWTSecret == config.CollaborationWOPISecret {
		t.Error("generated secrets should be unique")
	}
}

func TestGenerateSecretProducesUniqueValues(t *testing.T) {
	first, err := generateSecret()
	if err != nil {
		t.Fatalf("first generation failed: %v", err)
	}

	second, err := generateSecret()
	if err != nil {
		t.Fatalf("second generation failed: %v", err)
	}

	if first == second {
		t.Error("generated secrets should be different")
	}
}
