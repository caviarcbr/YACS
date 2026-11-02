package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDeploymentOCISOnly(t *testing.T) {
	root := t.TempDir()
	composeDir := filepath.Join(root, "compose")
	configDir := filepath.Join(root, "config")

	deployment, err := GenerateDeployment(
		composeDir,
		"ocis",
		configDir,
		false,
	)
	if err != nil {
		t.Fatalf("GenerateDeployment() error = %v", err)
	}

	if deployment.Name != "ocis" {
		t.Errorf("deployment name = %q, want %q", deployment.Name, "ocis")
	}

	content, err := os.ReadFile(deployment.ComposeFile)
	if err != nil {
		t.Fatalf("reading generated Compose file: %v", err)
	}

	rendered := string(content)

	if !strings.Contains(rendered, "  ocis:") {
		t.Error("generated Compose file is missing the oCIS service")
	}

	if strings.Contains(rendered, "  onlyoffice:") {
		t.Error("oCIS-only deployment unexpectedly includes ONLYOFFICE")
	}

	if strings.Contains(rendered, "  collaboration:") {
		t.Error("oCIS-only deployment unexpectedly includes collaboration")
	}

	if !strings.Contains(rendered, filepath.Join(configDir, "certs")) {
		t.Error("generated Compose file is missing the certificate directory")
	}

	cspPath := filepath.Join(deployment.Directory, "csp.yaml")
	if _, err := os.Stat(cspPath); err != nil {
		t.Errorf("CSP configuration was not generated: %v", err)
	}
}

func TestGenerateDeploymentWithOnlyOffice(t *testing.T) {
	root := t.TempDir()

	deployment, err := GenerateDeployment(
		filepath.Join(root, "compose"),
		"personal-cloud",
		filepath.Join(root, "config"),
		true,
	)
	if err != nil {
		t.Fatalf("GenerateDeployment() error = %v", err)
	}

	content, err := os.ReadFile(deployment.ComposeFile)
	if err != nil {
		t.Fatalf("reading generated Compose file: %v", err)
	}

	rendered := string(content)

	for _, service := range []string{
		"  ocis:",
		"  onlyoffice:",
		"  collaboration:",
	} {
		if !strings.Contains(rendered, service) {
			t.Errorf("generated Compose file is missing %q", service)
		}
	}
}

func TestGenerateDeploymentRefusesExistingDirectory(t *testing.T) {
	root := t.TempDir()
	composeDir := filepath.Join(root, "compose")
	deploymentDir := filepath.Join(composeDir, "ocis")

	if err := os.MkdirAll(deploymentDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingFile := filepath.Join(deploymentDir, "keep.txt")
	if err := os.WriteFile(existingFile, []byte("preserve"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := GenerateDeployment(
		composeDir,
		"ocis",
		filepath.Join(root, "config"),
		false,
	)
	if err == nil {
		t.Fatal("GenerateDeployment() succeeded; want existing-directory error")
	}

	content, readErr := os.ReadFile(existingFile)
	if readErr != nil {
		t.Fatalf("existing file was not preserved: %v", readErr)
	}

	if string(content) != "preserve" {
		t.Errorf("existing file content = %q, want %q", content, "preserve")
	}
}

func TestGenerateDeploymentRejectsInvalidName(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{
		"",
		".",
		"..",
		"../outside",
		"nested/deployment",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := GenerateDeployment(
				filepath.Join(root, "compose"),
				name,
				filepath.Join(root, "config"),
				false,
			)
			if err == nil {
				t.Errorf("GenerateDeployment(%q) succeeded; want error", name)
			}
		})
	}
}
