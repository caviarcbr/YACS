package compose

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/caviarcbr/YACS/internal/docker"
)

type mockDockerRunner struct {
	args      []string
	output    string
	err       error
	calls     [][]string
	callCount int
}

func (m *mockDockerRunner) Run(
	ctx context.Context,
	args ...string,
) (string, error) {
	m.args = args
	m.callCount++
	m.calls = append(m.calls, append([]string(nil), args...))

	return m.output, m.err
}

func TestNewClient(t *testing.T) {
	dockerClient := &docker.Client{
		Binary: "/usr/bin/docker",
	}

	client, err := NewClient(
		dockerClient,
		"/tmp/docker-compose.yaml",
		"yacs-test",
	)

	if err != nil {
		t.Fatalf("NewClient() returned an error: %v", err)
	}

	if client.ComposeFile != "/tmp/docker-compose.yaml" {
		t.Errorf("unexpected Compose file: %s", client.ComposeFile)
	}

	if client.ProjectName != "yacs-test" {
		t.Errorf("unexpected project name: %s", client.ProjectName)
	}
}

func TestNewClientRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name        string
		docker      DockerRunner
		composeFile string
		projectName string
	}{
		{"nil Docker client", nil, "/tmp/compose.yaml", "yacs"},
		{"empty Compose file", &docker.Client{}, "", "yacs"},
		{"empty project name", &docker.Client{}, "/tmp/compose.yaml", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.docker, tt.composeFile, tt.projectName)
			if err == nil {
				t.Fatal("expected an error for invalid arguments")
			}
		})
	}
}

func TestValidateMissingComposeFile(t *testing.T) {
	dockerClient := &docker.Client{
		Binary: "/usr/bin/docker",
	}

	client, err := NewClient(
		dockerClient,
		filepath.Join(t.TempDir(), "missing.yaml"),
		"yacs-test",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Validate(context.Background())
	if err == nil {
		t.Fatal("expected an error for a missing Compose file")
	}
}

func TestValidateExistingComposeFile(t *testing.T) {
	dir := t.TempDir()
	composeFile := filepath.Join(dir, "compose.yaml")

	content := []byte(
		"services:\n" +
			"  test:\n" +
			"    image: hello-world\n",
	)

	if err := os.WriteFile(composeFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	dockerClient := &docker.Client{
		Binary: "/usr/bin/docker",
	}

	client, err := NewClient(dockerClient, composeFile, "yacs-test")
	if err != nil {
		t.Fatal(err)
	}

	// This test only verifies that the file exists.
	// Actual Docker Compose validation will be tested separately.
	if _, err := os.Stat(client.ComposeFile); err != nil {
		t.Fatalf("Compose file should exist: %v", err)
	}
}

func TestComposeUninstall(t *testing.T) {
	tests := []struct {
		name       string
		removeData bool
		wantArgs   []string
	}{
		{
			name:       "preserve volumes",
			removeData: false,
			wantArgs: []string{
				"compose",
				"-f", "/tmp/compose.yaml",
				"-p", "yacs-test",
				"down", "--remove-orphans",
			},
		},
		{
			name:       "remove volumes",
			removeData: true,
			wantArgs: []string{
				"compose",
				"-f", "/tmp/compose.yaml",
				"-p", "yacs-test",
				"down", "--remove-orphans", "--volumes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDockerRunner{}

			client, err := NewClient(
				mock,
				"/tmp/compose.yaml",
				"yacs-test",
			)
			if err != nil {
				t.Fatal(err)
			}

			if err := client.Uninstall(context.Background(), tt.removeData); err != nil {
				t.Fatalf("Uninstall() returned an error: %v", err)
			}

			if !reflect.DeepEqual(mock.args, tt.wantArgs) {
				t.Errorf(
					"expected arguments %v, got %v",
					tt.wantArgs,
					mock.args,
				)
			}

			if mock.callCount != 1 {
				t.Errorf("expected 1 Docker call, got %d", mock.callCount)
			}
		})
	}
}

func TestComposeValidate(t *testing.T) {
	mock := &mockDockerRunner{}

	dir := t.TempDir()
	composeFile := filepath.Join(dir, "compose.yaml")

	if err := os.WriteFile(composeFile, []byte("services: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(mock, composeFile, "yacs-test")
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"compose",
		"-f", composeFile,
		"-p", "yacs-test",
		"config", "--quiet",
	}

	if !reflect.DeepEqual(mock.args, expected) {
		t.Errorf("expected %v, got %v", expected, mock.args)
	}
}
