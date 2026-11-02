package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestParseGlobalOptions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantConfig  string
		wantCommand []string
		wantErr     bool
	}{
		{
			name:        "no options",
			args:        []string{"install"},
			wantCommand: []string{"install"},
		},
		{
			name:        "config with separate argument",
			args:        []string{"--config", "/tmp/yacs.yaml", "install"},
			wantConfig:  "/tmp/yacs.yaml",
			wantCommand: []string{"install"},
		},
		{
			name:        "config with equals",
			args:        []string{"--config=/tmp/yacs.yaml", "uninstall"},
			wantConfig:  "/tmp/yacs.yaml",
			wantCommand: []string{"uninstall"},
		},
		{
			name:    "missing config path",
			args:    []string{"--config"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, gotCommand, err := parseGlobalOptions(tt.args)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotConfig != tt.wantConfig {
				t.Errorf("expected config %q, got %q", tt.wantConfig, gotConfig)
			}

			if strings.Join(gotCommand, " ") != strings.Join(tt.wantCommand, " ") {
				t.Errorf("expected command %v, got %v", tt.wantCommand, gotCommand)
			}
		})
	}
}

func TestRunVersion(t *testing.T) {
	output := captureOutput(t, func() {
		if err := Run([]string{"version"}); err != nil {
			t.Fatal(err)
		}
	})

	expected := AppName + " version " + AppVersion

	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestRunHelp(t *testing.T) {
	output := captureOutput(t, func() {
		if err := Run([]string{"help"}); err != nil {
			t.Fatal(err)
		}
	})

	for _, command := range []string{"install", "uninstall", "version"} {
		if !strings.Contains(output, command) {
			t.Errorf("help output should contain %q", command)
		}
	}

	for _, command := range []string{"doctor", "paths", "compose"} {
		if strings.Contains(output, command) {
			t.Errorf("help output should not contain %q", command)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	err := Run([]string{"doctor"})

	if err == nil {
		t.Fatal("expected an error for an unknown command")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("unexpected error: %v", err)
	}
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer
	fn()

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	os.Stdout = old

	var output bytes.Buffer
	if _, err := output.ReadFrom(reader); err != nil {
		t.Fatal(err)
	}

	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}

	return output.String()
}
