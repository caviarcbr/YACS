package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPaths(t *testing.T) {
	paths, err := DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths() returned an error: %v", err)
	}

	if paths.InstallDir == "" {
		t.Error("InstallDir should not be empty")
	}

	if paths.ConfigDir == "" {
		t.Error("ConfigDir should not be empty")
	}

	if paths.DataDir == "" {
		t.Error("DataDir should not be empty")
	}

	if runtime.GOOS == "linux" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("determining user home directory: %v", err)
		}

		dataRoot := os.Getenv("XDG_DATA_HOME")
		if dataRoot == "" {
			dataRoot = filepath.Join(homeDir, ".local", "share")
		}

		dataRoot = filepath.Join(dataRoot, "yacs")

		expected := Paths{
			InstallDir: dataRoot,
			ConfigDir:  filepath.Join(dataRoot, "config"),
			DataDir:    filepath.Join(dataRoot, "data"),
			ComposeDir: filepath.Join(dataRoot, "compose"),
		}

		if paths.InstallDir != expected.InstallDir {
			t.Errorf("expected InstallDir %s, got %s", expected.InstallDir, paths.InstallDir)
		}

		if paths.ConfigDir != expected.ConfigDir {
			t.Errorf("expected ConfigDir %s, got %s", expected.ConfigDir, paths.ConfigDir)
		}

		if paths.DataDir != expected.DataDir {
			t.Errorf("expected DataDir %s, got %s", expected.DataDir, paths.DataDir)
		}

		if paths.ComposeDir != expected.ComposeDir {
			t.Errorf("expected ComposeDir %s, got %s", expected.ComposeDir, paths.ComposeDir)
		}
	}
}
