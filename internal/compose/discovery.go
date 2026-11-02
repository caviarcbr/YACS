package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Deployment represents a discovered Docker Compose deployment.
type Deployment struct {
	Name        string
	Directory   string
	ComposeFile string
}

// DiscoverDeployments finds Compose deployments in the specified directory.
func DiscoverDeployments(composeDir string) ([]Deployment, error) {
	entries, err := os.ReadDir(composeDir)
	if err != nil {
		return nil, fmt.Errorf("reading Compose directory: %w", err)
	}

	var deployments []Deployment

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		directory := filepath.Join(composeDir, name)
		composeFile := filepath.Join(directory, "compose.yaml")

		info, err := os.Stat(composeFile)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		deployments = append(deployments, Deployment{
			Name:        name,
			Directory:   directory,
			ComposeFile: composeFile,
		})
	}

	sort.Slice(deployments, func(i, j int) bool {
		return deployments[i].Name < deployments[j].Name
	})

	return deployments, nil
}
