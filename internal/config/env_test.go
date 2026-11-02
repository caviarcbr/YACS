package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")

	values := map[string]string{
		"OCIS_JWT_SECRET": "test-secret",
		"PUBLIC_HOST":     "cloudbox",
		"BIND_ADDRESS":    "192.168.1.100",
	}

	if err := WriteEnvFile(path, values); err != nil {
		t.Fatalf("WriteEnvFile() returned an error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %04o", info.Mode().Perm())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range values {
		expected := key + "=" + value

		if !strings.Contains(string(content), expected) {
			t.Errorf("environment file is missing %q", expected)
		}
	}
}

func TestWriteEnvFileDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")

	original := []byte("OCIS_JWT_SECRET=original-secret\n")

	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	err := WriteEnvFile(path, map[string]string{
		"OCIS_JWT_SECRET": "replacement-secret",
	})

	if err == nil {
		t.Fatal("expected an error when environment file already exists")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != string(original) {
		t.Error("existing environment file was modified")
	}
}

func TestWriteEnvFileRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name:   "invalid key",
			values: map[string]string{"invalid-key": "value"},
		},
		{
			name:   "multiline value",
			values: map[string]string{"PUBLIC_HOST": "cloudbox\nOTHER=value"},
		},
		{
			name:   "empty configuration",
			values: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".env")

			if err := WriteEnvFile(path, tt.values); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
