package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caviarcbr/YACS/internal/compose"
	"github.com/caviarcbr/YACS/internal/config"
	"github.com/caviarcbr/YACS/internal/docker"
	"github.com/caviarcbr/YACS/internal/setup"
)

// RestoreInfo describes an existing YACS installation.
type RestoreInfo struct {
	DeploymentName string
	ServerURL      string
}

// InspectRestore discovers an existing installation without modifying it.
func InspectRestore(configPath string) (RestoreInfo, error) {
	configFile, err := config.ResolveConfigFile(configPath)
	if err != nil {
		return RestoreInfo{}, fmt.Errorf("resolving configuration file: %w", err)
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return RestoreInfo{}, fmt.Errorf("loading configuration: %w", err)
	}

	envPath := filepath.Join(paths.ConfigDir, ".env")

	if _, err := os.Stat(envPath); err != nil {
		if os.IsNotExist(err) {
			return RestoreInfo{}, fmt.Errorf("no existing YACS configuration found at %s", envPath)
		}

		return RestoreInfo{}, fmt.Errorf("checking existing configuration: %w", err)
	}

	deployments, err := compose.DiscoverDeployments(paths.ComposeDir)
	if err != nil {
		return RestoreInfo{}, fmt.Errorf("discovering existing deployment: %w", err)
	}

	if len(deployments) == 0 {
		return RestoreInfo{}, fmt.Errorf("no existing Compose deployment found in %s", paths.ComposeDir)
	}

	if len(deployments) > 1 {
		return RestoreInfo{}, fmt.Errorf(
			"found %d existing deployments; refusing to guess which one to restore",
			len(deployments),
		)
	}

	host, err := config.ReadEnvValue(envPath, "PUBLIC_HOST")
	if err != nil {
		return RestoreInfo{}, fmt.Errorf("reading existing server address: %w", err)
	}

	return RestoreInfo{
		DeploymentName: deployments[0].Name,
		ServerURL:      fmt.Sprintf("https://%s:9200", host),
	}, nil
}

// RestoreConfiguredWithProgress starts an existing deployment without
// modifying its configuration, certificates, credentials, or data.
func RestoreConfiguredWithProgress(
	configPath string,
	out io.Writer,
	progress func(string),
) error {
	configFile, err := config.ResolveConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("resolving configuration file: %w", err)
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	envPath := filepath.Join(paths.ConfigDir, ".env")

	if _, err := os.Stat(envPath); err != nil {
		return fmt.Errorf("checking existing configuration: %w", err)
	}

	deployments, err := compose.DiscoverDeployments(paths.ComposeDir)
	if err != nil {
		return fmt.Errorf("discovering existing deployment: %w", err)
	}

	if len(deployments) != 1 {
		return fmt.Errorf("expected exactly one existing deployment, found %d", len(deployments))
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return err
	}

	if err := dockerClient.PreflightTUI(context.Background()); err != nil {
		return err
	}

	client, err := compose.NewClient(
		dockerClient,
		deployments[0].ComposeFile,
		deployments[0].Name,
	)
	if err != nil {
		return fmt.Errorf("initializing Compose client: %w", err)
	}

	client.SetEnvFile(envPath)

	if progress != nil {
		progress("Validating Compose configuration...")
	}

	if err := client.InstallStream(context.Background(), out, progress); err != nil {
		return fmt.Errorf(
			"restoring deployment %q: %w\nExisting configuration was preserved at %s",
			deployments[0].Name,
			err,
			envPath,
		)
	}

	fmt.Fprintln(out, "\n[OK] Existing YACS deployment restored.")

	return nil
}

func resumeExistingInstall(
	dockerClient *docker.Client,
	paths config.Paths,
	envPath string,
	in io.Reader,
	out io.Writer,
) error {
	deployments, err := compose.DiscoverDeployments(paths.ComposeDir)
	if err != nil {
		return fmt.Errorf("discovering existing deployment: %w", err)
	}

	if len(deployments) == 0 {
		return fmt.Errorf(
			"found existing configuration at %s, but no Compose deployment exists in %s; refusing to overwrite existing configuration",
			envPath,
			paths.ComposeDir,
		)
	}

	if len(deployments) > 1 {
		return fmt.Errorf(
			"found %d existing deployments; refusing to guess which one to restore",
			len(deployments),
		)
	}

	deployment := deployments[0]

	fmt.Fprintln(out, "\nAn existing YACS installation was found.")
	fmt.Fprintf(out, "  Deployment:    %s\n", deployment.Name)
	fmt.Fprintf(out, "  Configuration: %s\n", envPath)
	fmt.Fprintln(out, "\nExisting credentials, certificates, and data will be preserved.")

	prompter := setup.NewPrompter(in, out)

	confirmed, err := prompter.Confirm("Restore this installation?")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Fprintln(out, "Restoration cancelled. Nothing was changed.")
		return nil
	}

	client, err := compose.NewClient(
		dockerClient,
		deployment.ComposeFile,
		deployment.Name,
	)
	if err != nil {
		return fmt.Errorf("initializing Compose client: %w", err)
	}

	client.SetEnvFile(filepath.Join(paths.ConfigDir, ".env"))

	progress := func(message string) {
		fmt.Fprintln(out, message)
	}

	fmt.Fprintf(out, "\nRestoring deployment %q...\n", deployment.Name)

	if err := client.InstallStream(context.Background(), out, progress); err != nil {
		return fmt.Errorf(
			"restoring deployment %q: %w\nExisting configuration was preserved at %s",
			deployment.Name,
			err,
			envPath,
		)
	}

	host, err := config.ReadEnvValue(envPath, "PUBLIC_HOST")
	if err != nil {
		return fmt.Errorf(
			"deployment was restored, but could not read its server address: %w",
			err,
		)
	}

	serverURL := fmt.Sprintf("https://%s:9200", host)

	fmt.Fprintln(out, "\n[OK] Existing YACS deployment restored.")
	fmt.Fprintf(out, "Server URL:      %s\n", serverURL)
	fmt.Fprintln(out, "Admin username:  admin")
	fmt.Fprintln(out, "Admin password:  unchanged from the original installation")

	return nil
}
