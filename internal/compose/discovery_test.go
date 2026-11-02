package compose

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverDeployments(t *testing.T) {
	root := t.TempDir()

	deployments := []string{
		"ocis",
		"onlyoffice",
		"proxy",
	}

	for _, name := range deployments {
		dir := filepath.Join(root, name)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}

		composeFile := filepath.Join(dir, "compose.yaml")

		if err := os.WriteFile(composeFile, []byte("services: {}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// This directory should not be discovered.
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0755); err != nil {
		t.Fatal(err)
	}

	// Hidden directories should be ignored.
	if err := os.MkdirAll(filepath.Join(root, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, ".hidden", "compose.yaml"),
		[]byte("services: {}\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	found, err := DiscoverDeployments(root)
	if err != nil {
		t.Fatalf("DiscoverDeployments() returned an error: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 deployments, got %d", len(found))
	}

	expected := []string{"ocis", "onlyoffice", "proxy"}

	for i, deployment := range found {
		if deployment.Name != expected[i] {
			t.Errorf("expected deployment %q, got %q",
				expected[i], deployment.Name)
		}

		if deployment.ComposeFile != filepath.Join(
			root, expected[i], "compose.yaml",
		) {
			t.Errorf("unexpected Compose file path: %s",
				deployment.ComposeFile)
		}
	}
}

func TestDiscoverDeploymentsMissingDirectory(t *testing.T) {
	_, err := DiscoverDeployments(
		filepath.Join(t.TempDir(), "missing"),
	)

	if err == nil {
		t.Fatal("expected an error for a missing Compose directory")
	}
}

func TestDiscoverDeploymentsEmptyDirectory(t *testing.T) {
	root := t.TempDir()

	found, err := DiscoverDeployments(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("expected no deployments, got %d", len(found))
	}
}
