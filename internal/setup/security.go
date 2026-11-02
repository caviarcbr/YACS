package setup

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type SecurityConfig struct {
	OCISJWTSecret           string
	OnlyOfficeJWTSecret     string
	CollaborationWOPISecret string

	// AdminPassword is the password the user chose for the oCIS admin account.
	AdminPassword string
}

// generateSecret creates a cryptographically secure random secret.
func generateSecret() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate security secret: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// NewSecurityConfig generates the secrets required by the selected services.
func NewSecurityConfig(includeOnlyOffice bool) (SecurityConfig, error) {
	config := SecurityConfig{}

	var err error

	config.OCISJWTSecret, err = generateSecret()
	if err != nil {
		return SecurityConfig{}, err
	}

	if includeOnlyOffice {
		config.OnlyOfficeJWTSecret, err = generateSecret()
		if err != nil {
			return SecurityConfig{}, err
		}

		config.CollaborationWOPISecret, err = generateSecret()
		if err != nil {
			return SecurityConfig{}, err
		}
	}

	return config, nil
}
