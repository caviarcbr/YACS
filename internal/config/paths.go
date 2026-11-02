package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Paths struct {
	InstallDir string
	ConfigDir  string
	DataDir    string
	ComposeDir string
	// ModulesDir string
}

func DefaultPaths() (Paths, error) {
	var paths Paths

	switch runtime.GOOS {
	case "linux":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return Paths{}, fmt.Errorf("determining home directory: %w", err)
			}

			dataHome = filepath.Join(homeDir, ".local", "share")
		}

		installDir := filepath.Join(dataHome, "yacs")

		paths = Paths{
			InstallDir: installDir,
			ConfigDir:  filepath.Join(installDir, "config"),
			DataDir:    filepath.Join(installDir, "data"),
			ComposeDir: filepath.Join(installDir, "compose"),
			// ModulesDir: filepath.Join(installDir, "modules"),
		}

	case "windows":
		programFiles := os.Getenv("ProgramFiles")
		programData := os.Getenv("ProgramData")

		if programFiles == "" || programData == "" {
			return Paths{}, fmt.Errorf(
				"could not determine Windows system directories",
			)
		}

		paths = Paths{
			InstallDir: filepath.Join(programFiles, "YACS"),
			ComposeDir: filepath.Join(programFiles, "YACS", "compose"),
			ConfigDir:  filepath.Join(programData, "YACS", "config"),
			DataDir:    filepath.Join(programData, "YACS", "data"),
			// ModulesDir: filepath.Join(programFiles, "YACS", "modules"),
		}

	default:
		return Paths{}, fmt.Errorf(
			"unsupported operating system: %s",
			runtime.GOOS,
		)
	}

	return paths, nil
}
