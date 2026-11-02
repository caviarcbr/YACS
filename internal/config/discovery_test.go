package config

import (
	"path/filepath"
	"testing"
)

func TestResolveConfigFileExplicitPath(t *testing.T) {
	t.Setenv(ConfigEnvVar, "/tmp/environment.yaml")

	explicit := "/tmp/explicit.yaml"

	got, err := ResolveConfigFile(explicit)
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Clean(explicit) {
		t.Errorf("expected %s, got %s", explicit, got)
	}
}

func TestResolveConfigFileEnvironment(t *testing.T) {
	expected := "/tmp/environment.yaml"
	t.Setenv(ConfigEnvVar, expected)

	got, err := ResolveConfigFile("")
	if err != nil {
		t.Fatal(err)
	}

	if got != filepath.Clean(expected) {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestResolveConfigFileDefault(t *testing.T) {
	t.Setenv(ConfigEnvVar, "")

	defaults, err := DefaultPaths()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(defaults.ConfigDir, "config.yaml")

	got, err := ResolveConfigFile("")
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
