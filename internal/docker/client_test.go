package docker

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewClientFindsDocker(t *testing.T) {
	dir := t.TempDir()

	binaryName := "docker"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binary := filepath.Join(dir, binaryName)

	if err := os.WriteFile(binary, []byte(""), 0700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir)

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() returned an error: %v", err)
	}

	if client.Binary != binary {
		t.Errorf("expected Docker binary %q, got %q", binary, client.Binary)
	}
}

func TestNewClientDockerNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	_, err := NewClient()
	if err == nil {
		t.Fatal("expected an error when Docker is not found")
	}
}
