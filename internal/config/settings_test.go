package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadSettingsMissingFile(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := LoadSettings(configFile)
	if err == nil {
		t.Fatal("expected an error for a missing configuration file")
	}
}

func TestLoadSettingsCustomPaths(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	customDataDir := filepath.Join(dir, "custom-data")

	customComposeDir := filepath.Join(dir, "custom-compose")

	content := []byte(
		"paths:\n" +
			"  data: " + customDataDir + "\n" +
			"  compose: " + customComposeDir + "\n",
	)

	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatalf("could not write test configuration: %v", err)
	}

	paths, err := LoadSettings(configFile)
	if err != nil {
		t.Fatalf("LoadSettings() returned an error: %v", err)
	}

	if paths.DataDir != customDataDir {
		t.Errorf("expected data directory %s, got %s",
			customDataDir, paths.DataDir)
	}

	if paths.ComposeDir != customComposeDir {
		t.Errorf("expected Compose directory %s, got %s",
			customComposeDir, paths.ComposeDir)
	}

	defaults, err := DefaultPaths()
	if err != nil {
		t.Fatal(err)
	}

	if paths.InstallDir != defaults.InstallDir {
		t.Error("install directory should retain its default")
	}
	if paths.ComposeDir == "" {
		t.Error("Compose directory should not be empty")
	}
}

func TestLoadSettingsRejectsRelativePaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relative path validation test uses Unix-style paths")
	}

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	content := []byte("paths:\n  data: relative/data\n")

	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSettings(configFile)
	if err == nil {
		t.Fatal("expected an error for a relative path")
	}
}

func TestLoadSettingsRejectsInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	content := []byte("paths: [this is not valid")

	if err := os.WriteFile(configFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSettings(configFile)
	if err == nil {
		t.Fatal("expected an error for invalid YAML")
	}
}
