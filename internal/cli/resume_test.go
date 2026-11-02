package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRestoreFixture(t *testing.T) (string, string, string) {
	t.Helper()

	root := t.TempDir()

	configDir := filepath.Join(root, "config")
	composeDir := filepath.Join(root, "compose")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(composeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(root, "config.yaml")

	content := "paths:\n" +
		"  config: " + configDir + "\n" +
		"  compose: " + composeDir + "\n"

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	return configFile, configDir, composeDir
}

func addRestoreDeployment(t *testing.T, composeDir, name string) {
	t.Helper()

	deploymentDir := filepath.Join(composeDir, name)

	if err := os.MkdirAll(deploymentDir, 0755); err != nil {
		t.Fatal(err)
	}

	composeFile := filepath.Join(deploymentDir, "compose.yaml")

	if err := os.WriteFile(composeFile, []byte("services: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestInspectRestore(t *testing.T) {
	configFile, configDir, composeDir := setupRestoreFixture(t)

	if err := os.WriteFile(
		filepath.Join(configDir, ".env"),
		[]byte("PUBLIC_HOST=192.168.1.100\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	addRestoreDeployment(t, composeDir, "personal-cloud")

	info, err := InspectRestore(configFile)
	if err != nil {
		t.Fatalf("InspectRestore() returned an error: %v", err)
	}

	if info.DeploymentName != "personal-cloud" {
		t.Errorf("expected deployment name personal-cloud, got %q", info.DeploymentName)
	}

	if info.ServerURL != "https://192.168.1.100:9200" {
		t.Errorf("unexpected server URL: %q", info.ServerURL)
	}
}

func TestInspectRestoreMissingConfiguration(t *testing.T) {
	configFile, _, _ := setupRestoreFixture(t)

	_, err := InspectRestore(configFile)
	if err == nil {
		t.Fatal("expected an error for missing environment configuration")
	}

	if !strings.Contains(err.Error(), "no existing YACS configuration") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInspectRestoreNoDeployments(t *testing.T) {
	configFile, configDir, _ := setupRestoreFixture(t)

	if err := os.WriteFile(
		filepath.Join(configDir, ".env"),
		[]byte("PUBLIC_HOST=192.168.1.100\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	_, err := InspectRestore(configFile)
	if err == nil {
		t.Fatal("expected an error when no Compose deployment exists")
	}

	if !strings.Contains(err.Error(), "no existing Compose deployment") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInspectRestoreMultipleDeployments(t *testing.T) {
	configFile, configDir, composeDir := setupRestoreFixture(t)

	if err := os.WriteFile(
		filepath.Join(configDir, ".env"),
		[]byte("PUBLIC_HOST=192.168.1.100\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	addRestoreDeployment(t, composeDir, "deployment-one")
	addRestoreDeployment(t, composeDir, "deployment-two")

	_, err := InspectRestore(configFile)
	if err == nil {
		t.Fatal("expected an error for multiple deployments")
	}

	if !strings.Contains(err.Error(), "refusing to guess") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInspectRestoreMissingPublicHost(t *testing.T) {
	configFile, configDir, composeDir := setupRestoreFixture(t)

	if err := os.WriteFile(
		filepath.Join(configDir, ".env"),
		[]byte("OTHER_SETTING=value\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	addRestoreDeployment(t, composeDir, "personal-cloud")

	_, err := InspectRestore(configFile)
	if err == nil {
		t.Fatal("expected an error when PUBLIC_HOST is missing")
	}

	if !strings.Contains(err.Error(), "PUBLIC_HOST") {
		t.Errorf("unexpected error: %v", err)
	}
}
