package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caviarcbr/YACS/internal/compose"
	"github.com/caviarcbr/YACS/internal/config"
	"github.com/caviarcbr/YACS/internal/docker"
	"github.com/caviarcbr/YACS/internal/setup"
)

// runInstall installs YACS interactively on the terminal.
//
// Every requirement is checked before the first question is asked, so a
// sudo password prompt appears up front and nothing is written until the
// user has confirmed the settings.
func runInstall(configPath string, in io.Reader, out io.Writer) error {
	dockerClient, err := prepareDocker(out)
	if err != nil {
		return err
	}

	configFile, err := config.ResolveConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to locate configuration: %w", err)
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	envPath := filepath.Join(paths.ConfigDir, ".env")

	if _, err := os.Stat(envPath); err == nil {
		return resumeExistingInstall(dockerClient, paths, envPath, in, out)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking existing environment file: %w", err)
	}
	if err := checkNoLeftoverVolumes(dockerClient); err != nil {
		return err
	}

	prompter := setup.NewPrompter(in, out)

	installation, err := setup.PromptInstallation(prompter, setup.DetectDefaultHost())
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Installation summary")
	fmt.Fprintf(out, "  Services:       %s\n", installation.Services)
	fmt.Fprintf(out, "  Server address: %s\n", installation.Network.PublicHost)
	fmt.Fprintf(out, "  Admin account:  admin\n")
	fmt.Fprintf(out, "  Configuration:  %s\n", paths.ConfigDir)
	fmt.Fprintln(out, "")

	confirmed, err := prompter.Confirm("Install with these settings?")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Fprintln(out, "Installation cancelled. Nothing was changed.")
		return nil
	}

	return deployInstallation(dockerClient, paths, installation, out)

}

// InstallConfigured installs YACS using settings collected by another interface.
func InstallConfigured(
	configPath string,
	installation setup.Installation,
	out io.Writer,
) error {
	return InstallConfiguredWithProgress(
		configPath,
		installation,
		out,
		nil,
	)
}

// InstallConfiguredWithProgress installs YACS and reports installation stages
// through progress when a callback is provided.
func InstallConfiguredWithProgress(
	configPath string,
	installation setup.Installation,
	out io.Writer,
	progress func(string),
) error {
	if out == nil {
		out = io.Discard
	}

	dockerClient, err := prepareDocker(out)
	if err != nil {
		return err
	}

	configFile, err := config.ResolveConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("resolving configuration: %w", err)
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	envPath := filepath.Join(paths.ConfigDir, ".env")

	if _, err := os.Stat(envPath); err == nil {
		return fmt.Errorf("an existing installation was found; use the restore workflow instead")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking existing installation: %w", err)
	}

	if err := checkNoLeftoverVolumes(dockerClient); err != nil {
		return err
	}

	return deployInstallationWithProgress(
		dockerClient,
		paths,
		installation,
		out,
		progress,
	)
}

// checkNoLeftoverVolumes stops installation when named volumes from an
// earlier YACS deployment still exist.
func checkNoLeftoverVolumes(client *docker.Client) error {
	projects := []string{"ocis", "personal-cloud"}

	var volumes []string

	for _, project := range projects {
		output, err := client.Run(
			context.Background(),
			"volume", "ls", "--quiet",
			"--filter", "label=com.docker.compose.project="+project,
		)
		if err != nil {
			return fmt.Errorf(
				"checking for leftover volumes belonging to project %q: %w",
				project,
				err,
			)
		}

		volumes = append(volumes, strings.Fields(output)...)
	}

	if len(volumes) == 0 {
		return nil
	}

	return fmt.Errorf(
		"leftover YACS data volumes from an earlier installation were found: %s\n"+
			"These volumes may contain an existing admin account and password.\n"+
			"Run `yacs uninstall --remove-data` before installing again.",
		strings.Join(volumes, ", "),
	)
}

// prepareDocker verifies the local requirements and returns a Docker client
// ready for use. It may prompt for a sudo password, so it must run before any
// installation question.
func prepareDocker(out io.Writer) (*docker.Client, error) {
	if _, err := exec.LookPath("openssl"); err != nil {
		return nil, errors.New("OpenSSL is required to generate TLS certificates")
	}

	client, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	if err := client.Preflight(context.Background(), out); err != nil {
		return nil, err
	}

	if _, err := client.ComposeVersion(context.Background()); err != nil {
		return nil, fmt.Errorf("Docker Compose v2 is required: %w", err)
	}

	fmt.Fprintln(out, "[OK] Docker and Docker Compose are available.")

	return client, nil
}

// CheckDockerAccessForTUI checks whether the TUI needs to request sudo
// authentication before starting an installation.
func CheckDockerAccessForTUI() (bool, error) {
	if _, err := exec.LookPath("openssl"); err != nil {
		return false, errors.New("OpenSSL is required to generate TLS certificates")
	}

	client, err := docker.NewClient()
	if err != nil {
		return false, err
	}

	err = client.PreflightTUI(context.Background())

	if errors.Is(err, docker.ErrSudoAuthenticationRequired) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	if _, err := client.ComposeVersion(context.Background()); err != nil {
		return false, fmt.Errorf("Docker Compose v2 is required: %w", err)
	}

	return false, nil
}

// CheckDockerAccessForUninstallTUI checks Docker access before the TUI
// starts an uninstall operation.
func CheckDockerAccessForUninstallTUI() (bool, error) {
	client, err := docker.NewClient()
	if err != nil {
		return false, err
	}

	err = client.PreflightTUI(context.Background())

	if errors.Is(err, docker.ErrSudoAuthenticationRequired) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	if _, err := client.ComposeVersion(context.Background()); err != nil {
		return false, fmt.Errorf("Docker Compose v2 is required: %w", err)
	}

	return false, nil
}

// CheckDockerAccessForRestoreTUI checks Docker access before the TUI
// starts a restore operation.
func CheckDockerAccessForRestoreTUI() (bool, error) {
	client, err := docker.NewClient()
	if err != nil {
		return false, err
	}

	err = client.PreflightTUI(context.Background())

	if errors.Is(err, docker.ErrSudoAuthenticationRequired) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	if _, err := client.ComposeVersion(context.Background()); err != nil {
		return false, fmt.Errorf("Docker Compose v2 is required: %w", err)
	}

	return false, nil
}

// AuthenticateDockerForTUI validates sudo credentials for the TUI installer.
func AuthenticateDockerForTUI(password string) error {
	client, err := docker.NewClient()
	if err != nil {
		return err
	}

	if err := client.AuthenticateSudo(context.Background(), password); err != nil {
		return err
	}

	// Confirm that Docker is accessible with the newly authenticated credentials.
	client.Sudo = true

	if _, err := client.ServerVersion(context.Background()); err != nil {
		return fmt.Errorf("Docker is not reachable with sudo: %w", err)
	}

	if _, err := client.ComposeVersion(context.Background()); err != nil {
		return fmt.Errorf("Docker Compose v2 is required: %w", err)
	}

	return nil
}

func deployInstallation(
	dockerClient *docker.Client,
	paths config.Paths,
	installation setup.Installation,
	out io.Writer,
) error {
	return deployInstallationWithProgress(
		dockerClient,
		paths,
		installation,
		out,
		nil,
	)
}

func deployInstallationWithProgress(
	dockerClient *docker.Client,
	paths config.Paths,
	installation setup.Installation,
	out io.Writer,
	progress func(string),
) error {
	reportProgress := func(message string) {
		fmt.Fprintln(out, message)

		if progress != nil {
			progress(message)
		}
	}

	withOnlyOffice := installation.Services == setup.OCISWithOnlyOffice

	security, err := setup.NewSecurityConfig(withOnlyOffice)
	if err != nil {
		return fmt.Errorf("generating security configuration: %w", err)
	}

	security.AdminPassword = installation.AdminPassword

	envPath := filepath.Join(paths.ConfigDir, ".env")
	certDir := filepath.Join(paths.ConfigDir, "certs")

	certDirExisted := false
	if info, err := os.Stat(certDir); err == nil {
		certDirExisted = info.IsDir()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking certificate directory: %w", err)
	}

	envCreated := false
	deploymentCreated := false

	defer func() {
		if deploymentCreated {
			return
		}

		if envCreated {
			_ = os.Remove(envPath)
		}

		if !certDirExisted {
			_ = os.RemoveAll(certDir)
		}
	}()

	reportProgress("Preparing TLS certificates...")

	if _, err := config.PrepareCertificates(
		certDir,
		installation.Network.PublicHost,
	); err != nil {
		return fmt.Errorf("preparing TLS certificates: %w", err)
	}

	reportProgress("Saving installation configuration...")

	values := setup.EnvironmentValues(
		installation.Services,
		installation.Network,
		security,
	)

	if err := config.WriteEnvFile(envPath, values); err != nil {
		return fmt.Errorf("saving installation configuration: %w", err)
	}

	envCreated = true

	deploymentName := "ocis"
	if withOnlyOffice {
		deploymentName = "personal-cloud"
	}

	reportProgress("Generating Docker Compose deployment...")

	deployment, err := compose.GenerateDeployment(
		paths.ComposeDir,
		deploymentName,
		paths.ConfigDir,
		withOnlyOffice,
	)
	if err != nil {
		return fmt.Errorf("generating Compose deployment: %w", err)
	}

	deploymentCreated = true

	composeClient, err := compose.NewClient(
		dockerClient,
		deployment.ComposeFile,
		deployment.Name,
	)
	if err != nil {
		return fmt.Errorf("initializing Compose client: %w", err)
	}

	composeClient.SetEnvFile(envPath)

	composeProgress := func(message string) {
		reportProgress(message)
	}

	if err := composeClient.InstallQuiet(
		context.Background(),
		composeProgress,
	); err != nil {
		return fmt.Errorf(
			"%w\nThe configuration was kept at %s. Run `yacs uninstall --remove-data` or delete that file before trying again.",
			err,
			envPath,
		)
	}

	serverURL := fmt.Sprintf(
		"https://%s:9200",
		installation.Network.PublicHost,
	)

	fmt.Fprintf(out, "\n[OK] %s is running.\n\n", installation.Services)
	fmt.Fprintf(out, "Server URL:      %s\n", serverURL)
	fmt.Fprintln(out, "Admin username:  admin")
	fmt.Fprintln(out, "Admin password:  the password you just chose")

	return nil
}
