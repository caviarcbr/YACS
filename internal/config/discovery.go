package config

import (
	"os"
	"path/filepath"
	"strings"
)

const ConfigEnvVar = "YACS_CONFIG"

// DefaultConfigFile returns the operating system's default configuration file.
func DefaultConfigFile() string {
	paths, err := DefaultPaths()
	if err != nil {
		return ""
	}

	return filepath.Join(paths.ConfigDir, "config.yaml")
}

// ResolveConfigFile determines which configuration file YACS should use.
func ResolveConfigFile(explicitPath string) (string, error) {
	// 1. Explicit command-line path takes priority.
	if path := strings.TrimSpace(explicitPath); path != "" {
		return filepath.Clean(path), nil
	}

	// 2. Check the YACS_CONFIG environment variable.
	if path := strings.TrimSpace(os.Getenv(ConfigEnvVar)); path != "" {
		return filepath.Clean(path), nil
	}

	// 3. Fall back to the operating system default.
	return DefaultConfigFile(), nil
}
