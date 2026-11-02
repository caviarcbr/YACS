package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caviarcbr/YACS/internal/config"
)

func TestValidateCleanupPaths(t *testing.T) {
	root := t.TempDir()

	installDir := filepath.Join(root, "install")
	composeDir := filepath.Join(installDir, "compose")
	configDir := filepath.Join(root, "config")
	dataDir := filepath.Join(root, "data")

	for _, dir := range []string{
		composeDir,
		configDir,
		dataDir,
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	paths := config.Paths{
		InstallDir: installDir,
		ComposeDir: composeDir,
		ConfigDir:  configDir,
		DataDir:    dataDir,
	}

	got, err := validateCleanupPaths(paths)
	if err != nil {
		t.Fatalf("expected valid cleanup paths, got error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 cleanup paths, got %d: %v", len(got), got)
	}

	for _, path := range got {
		if path == composeDir {
			t.Errorf("ComposeDir should be covered by InstallDir, but was returned separately")
		}
	}
}

func TestValidateCleanupPathsRejectsSymlink(t *testing.T) {
	root := t.TempDir()

	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "config-link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	paths := config.Paths{
		InstallDir: filepath.Join(root, "install"),
		ComposeDir: filepath.Join(root, "compose"),
		ConfigDir:  link,
		DataDir:    filepath.Join(root, "data"),
	}

	_, err := validateCleanupPaths(paths)
	if err == nil {
		t.Fatal("expected cleanup validation to reject a symlink")
	}

	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("expected symlink error, got: %v", err)
	}
}
func TestRunUninstallCancellation(t *testing.T) {
	var output strings.Builder

	err := runUninstall(
		filepath.Join(t.TempDir(), "missing-config.yaml"),
		false,
		strings.NewReader("n\n"),
		&output,
	)
	if err != nil {
		t.Fatalf("expected cancellation without error, got: %v", err)
	}

	if !strings.Contains(output.String(), "Uninstallation cancelled.") {
		t.Errorf("expected cancellation message, got:\n%s", output.String())
	}
}

func TestRunUninstallRemoveDataRequiresSecondConfirmation(t *testing.T) {
	var output strings.Builder

	err := runUninstall(
		filepath.Join(t.TempDir(), "missing-config.yaml"),
		true,
		strings.NewReader("y\nno\n"),
		&output,
	)
	if err != nil {
		t.Fatalf("expected cancellation without error, got: %v", err)
	}

	if !strings.Contains(output.String(), "Uninstallation cancelled.") {
		t.Errorf("expected cancellation message, got:\n%s", output.String())
	}

	if !strings.Contains(output.String(), "Type DELETE") {
		t.Errorf("expected destructive confirmation prompt, got:\n%s", output.String())
	}
}
func TestValidateCleanupPathsRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()

	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Join(target, "config"), 0755); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "config-link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	paths := config.Paths{
		InstallDir: filepath.Join(root, "install"),
		ComposeDir: filepath.Join(root, "compose"),
		ConfigDir:  filepath.Join(link, "config"),
		DataDir:    filepath.Join(root, "data"),
	}

	_, err := validateCleanupPaths(paths)
	if err == nil {
		t.Fatal("expected cleanup validation to reject a symlink parent")
	}

	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("expected symlink error, got: %v", err)
	}
}

func TestValidateCleanupPathsAllowsManagedDirectoriesInHome(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("determining home directory: %v", err)
	}

	root, err := os.MkdirTemp(homeDir, ".yacs-uninstall-test-")
	if err != nil {
		t.Fatalf("creating temporary home directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })

	installDir := filepath.Join(root, ".local", "share", "yacs")
	composeDir := filepath.Join(installDir, "compose")
	configDir := filepath.Join(root, ".config", "yacs")
	dataDir := filepath.Join(installDir, "data")

	for _, dir := range []string{composeDir, configDir, dataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	paths := config.Paths{
		InstallDir: installDir,
		ComposeDir: composeDir,
		ConfigDir:  configDir,
		DataDir:    dataDir,
	}

	if _, err := validateCleanupPaths(paths); err != nil {
		t.Fatalf("expected managed directories beneath home to be allowed, got: %v", err)
	}
}
