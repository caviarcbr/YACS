package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caviarcbr/YACS/internal/compose"
	"github.com/caviarcbr/YACS/internal/config"
	"github.com/caviarcbr/YACS/internal/docker"
)

func newDeploymentClient(
	explicitConfigPath string,
	deployment *compose.Deployment,
) (*compose.Client, error) {
	configFile, err := config.ResolveConfigFile(explicitConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate configuration: %w", err)
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("initializing Docker client: %w", err)
	}

	if err := dockerClient.Preflight(context.Background(), os.Stderr); err != nil {
		return nil, fmt.Errorf("Docker preflight failed: %w", err)
	}

	return newDeploymentClientWithDocker(paths, deployment, dockerClient)
}

func newDeploymentClientWithDocker(
	paths config.Paths,
	deployment *compose.Deployment,
	dockerClient *docker.Client,
) (*compose.Client, error) {
	composeClient, err := compose.NewClient(
		dockerClient,
		deployment.ComposeFile,
		deployment.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("initializing Compose client: %w", err)
	}

	envPath := filepath.Join(paths.ConfigDir, ".env")

	if _, err := os.Stat(envPath); err == nil {
		composeClient.SetEnvFile(envPath)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("checking environment file: %w", err)
	}

	return composeClient, nil
}
