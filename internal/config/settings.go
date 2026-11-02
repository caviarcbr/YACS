package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Paths PathSettings `yaml:"paths"`
}

type PathSettings struct {
	InstallDir string `yaml:"install"`
	ConfigDir  string `yaml:"config"`
	DataDir    string `yaml:"data"`
	ComposeDir string `yaml:"compose"`
}

// LoadSettings loads YACS configuration and applies any custom paths.
func LoadSettings(configFile string) (Paths, error) {
	paths, err := DefaultPaths()
	if err != nil {
		return Paths{}, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) && configFile == DefaultConfigFile() {
			return paths, nil
		}

		return Paths{}, fmt.Errorf("reading configuration file %s: %w", configFile, err)
	}

	var settings Settings

	if err := yaml.Unmarshal(data, &settings); err != nil {
		return Paths{}, fmt.Errorf("parsing configuration: %w", err)
	}

	overrides := []struct {
		name  string
		value string
		dest  *string
	}{
		{"install", settings.Paths.InstallDir, &paths.InstallDir},
		{"config", settings.Paths.ConfigDir, &paths.ConfigDir},
		{"data", settings.Paths.DataDir, &paths.DataDir},
		{"compose", settings.Paths.ComposeDir, &paths.ComposeDir},
	}

	for _, override := range overrides {
		if override.value == "" {
			continue
		}

		if !filepath.IsAbs(override.value) {
			return Paths{}, fmt.Errorf(
				"%s path must be absolute: %s",
				override.name,
				override.value,
			)
		}

		*override.dest = filepath.Clean(override.value)
	}

	return paths, nil
}
