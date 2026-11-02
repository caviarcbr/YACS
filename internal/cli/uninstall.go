package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/caviarcbr/YACS/internal/compose"
	"github.com/caviarcbr/YACS/internal/config"
)

func runUninstall(
	configPath string,
	removeData bool,
	input io.Reader,
	output io.Writer,
) error {
	reader := bufio.NewReader(input)

	fmt.Fprintln(output, "This will stop and remove the YACS deployment.")
	fmt.Fprintln(output, "Docker images will be kept.")

	if removeData {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "WARNING: This will also permanently delete YACS data, configuration, and named Docker volumes.")
	}

	fmt.Fprint(output, "Continue with uninstallation? [y/N]: ")

	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading confirmation: %w", err)
	}

	answer = strings.TrimSpace(answer)
	if !strings.EqualFold(answer, "y") &&
		!strings.EqualFold(answer, "yes") {
		fmt.Fprintln(output, "Uninstallation cancelled.")
		return nil
	}

	if removeData {
		fmt.Fprint(output, "Type DELETE to permanently remove YACS data: ")

		answer, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("reading destructive confirmation: %w", err)
		}

		if strings.TrimSpace(answer) != "DELETE" {
			fmt.Fprintln(output, "Uninstallation cancelled.")
			return nil
		}
	}

	configFile, err := config.ResolveConfigFile(configPath)
	if err != nil {
		return err
	}

	paths, err := config.LoadSettings(configFile)
	if err != nil {
		return err
	}

	deployments, err := compose.DiscoverDeployments(paths.ComposeDir)
	if err != nil {
		return err
	}

	if len(deployments) == 0 {
		return fmt.Errorf("no YACS deployment found in %s; nothing was removed", paths.ComposeDir)
	}

	if len(deployments) > 1 {
		return fmt.Errorf(
			"found %d deployments in %s; refusing to guess which one to uninstall",
			len(deployments),
			paths.ComposeDir,
		)
	}

	deployment := deployments[0]

	var cleanupPaths []string

	if removeData {
		cleanupPaths, err = validateCleanupPaths(paths)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(output, "\nUninstalling deployment %q...\n", deployment.Name)

	client, err := newDeploymentClient(configPath, &deployment)
	if err != nil {
		return err
	}

	if err := client.Uninstall(context.Background(), removeData); err != nil {
		return err
	}

	fmt.Fprintln(output, "Compose deployment removed.")

	if removeData {
		for _, path := range cleanupPaths {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("removing YACS directory %s: %w", path, err)
			}

			fmt.Fprintf(output, "Removed %s\n", path)
		}
	}

	return nil
}

// validateCleanupPaths returns the directories that may be removed by
// --remove-data. It refuses paths that could include unrelated system
// or user data.
func validateCleanupPaths(paths config.Paths) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}

	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		return nil, err
	}

	targets := []string{
		paths.InstallDir,
		paths.ComposeDir,
		paths.ConfigDir,
		paths.DataDir,
	}

	var cleaned []string

	for _, target := range targets {
		absolute, err := filepath.Abs(target)
		if err != nil {
			return nil, fmt.Errorf("resolving cleanup path %q: %w", target, err)
		}

		absolute = filepath.Clean(absolute)

		if absolute == string(filepath.Separator) {
			return nil, fmt.Errorf("refusing to remove filesystem root")
		}

		relative, err := filepath.Rel(string(filepath.Separator), absolute)
		if err != nil {
			return nil, err
		}

		if len(strings.Split(relative, string(filepath.Separator))) < 2 {
			return nil, fmt.Errorf("refusing to remove broad system directory %s", absolute)
		}

		// YACS now stores its managed files beneath the user's home directory.
		// Permit those paths, but never remove the home directory itself or
		// any ancestor that would contain it.
		if absolute == homeDir || isPathWithin(homeDir, absolute) {
			return nil, fmt.Errorf("refusing to remove home directory or its parent: %s", absolute)
		}

		if err := rejectSymlinkComponents(absolute); err != nil {
			return nil, err
		}

		info, err := os.Lstat(absolute)
		if err == nil {
			if !info.IsDir() {
				return nil, fmt.Errorf("cleanup path is not a directory: %s", absolute)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("checking cleanup path %s: %w", absolute, err)
		}

		cleaned = append(cleaned, absolute)

	}

	// InstallDir normally contains ComposeDir. Remove the parent once,
	// rather than attempting to remove both directories independently.
	var result []string

	for _, candidate := range cleaned {
		covered := false

		for _, other := range cleaned {
			if candidate != other && isPathWithin(candidate, other) {
				covered = true
				break
			}
		}

		if !covered {
			result = append(result, candidate)
		}
	}

	// Reject overlapping cleanup directories. The only expected nesting
	// is ComposeDir inside InstallDir, handled above.
	for i, first := range result {
		for _, second := range result[i+1:] {
			if isPathWithin(first, second) || isPathWithin(second, first) {
				return nil, fmt.Errorf(
					"cleanup directories overlap: %s and %s",
					first,
					second,
				)
			}
		}
	}

	return result, nil
}

// isPathWithin reports whether child is a directory beneath parent.
func isPathWithin(child, parent string) bool {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	return relative != "." &&
		relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) &&
		!filepath.IsAbs(relative)
}

// rejectSymlinkComponents refuses paths that contain a symlink in any
// existing component, including parent directories.
func rejectSymlinkComponents(path string) error {
	current := string(filepath.Separator)
	relative := strings.TrimPrefix(path, current)

	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" {
			continue
		}

		current = filepath.Join(current, component)

		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("checking path component %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing cleanup path containing symlink: %s", current)
		}
	}

	return nil
}
