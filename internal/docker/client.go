package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Client represents the Docker command-line client.
type Client struct {
	Binary string

	// Sudo routes every Docker command through sudo. Preflight enables it
	// when the current user cannot reach the Docker daemon directly.
	Sudo bool
}

// NewClient locates the Docker executable.
func NewClient() (*Client, error) {
	binary, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("Docker CLI not found: %w", err)
	}

	return &Client{
		Binary: binary,
	}, nil
}

// Preflight confirms the Docker daemon is reachable before any work starts.
//
// If the current user lacks permission to reach the daemon, Preflight runs
// `sudo -v` on the terminal so the password prompt appears now, before the
// installer asks any questions. On success, later commands run through sudo.
// notice receives the message shown above the password prompt.
func (c *Client) Preflight(ctx context.Context, notice io.Writer) error {
	_, err := c.ServerVersion(ctx)
	if err == nil {
		return nil
	}

	if !isPermissionDenied(err) {
		return fmt.Errorf("Docker daemon is not reachable: %w", err)
	}

	if _, err := exec.LookPath("sudo"); err != nil {
		return errors.New(
			"Docker permission denied and sudo is not installed; " +
				"install sudo or add your user to the docker group",
		)
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return errors.New(
			"Docker permission denied and sudo is not installed; " +
				"install sudo or add your user to the docker group",
		)
	}

	// Check whether sudo credentials are already cached.
	// Suppress output because this is only a preliminary check.
	checkCredentials := exec.CommandContext(ctx, sudoPath, "-n", "-v")
	checkCredentials.Stdout = io.Discard
	checkCredentials.Stderr = io.Discard

	if err := checkCredentials.Run(); err != nil {
		if notice != nil {
			fmt.Fprintln(notice, "Docker needs elevated privileges. Enter your password for sudo.")
		}

		authenticate := exec.CommandContext(ctx, sudoPath, "-v")
		authenticate.Stdin = os.Stdin
		authenticate.Stdout = notice
		authenticate.Stderr = os.Stderr

		if err := authenticate.Run(); err != nil {
			return fmt.Errorf(
				"sudo authentication failed (this step needs an interactive terminal): %w",
				err,
			)
		}
	}

	c.Sudo = true

	if _, err := c.ServerVersion(ctx); err != nil {
		c.Sudo = false
		return fmt.Errorf("Docker daemon is not reachable with sudo: %w", err)
	}

	return nil
}

// command builds a Docker invocation, wrapped in sudo when Sudo is enabled.
func (c *Client) command(ctx context.Context, args []string) (*exec.Cmd, error) {
	if !c.Sudo {
		return exec.CommandContext(ctx, c.Binary, args...), nil
	}

	sudo, err := exec.LookPath("sudo")
	if err != nil {
		return nil, fmt.Errorf("sudo is required for Docker access but is not installed: %w", err)
	}

	return exec.CommandContext(ctx, sudo, append([]string{c.Binary}, args...)...), nil
}

// run executes a Docker command and returns its trimmed combined output.
func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	cmd, err := c.command(ctx, args)
	if err != nil {
		return "", err
	}

	if c.Sudo {
		cmd.Stdin = os.Stdin
	}

	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))

	if err != nil {
		return "", fmt.Errorf("%w: %s", err, text)
	}

	return text, nil
}

// RunStream executes a Docker command and streams its output as it arrives.
func (c *Client) RunStream(
	ctx context.Context,
	output io.Writer,
	args ...string,
) error {
	if output == nil {
		return fmt.Errorf("output writer cannot be nil")
	}

	cmd, err := c.command(ctx, args)
	if err != nil {
		return err
	}

	cmd.Stdout = output
	cmd.Stderr = output

	if c.Sudo {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker command failed: %w", err)
	}

	return nil
}

// ServerVersion returns the Docker daemon version.
func (c *Client) ServerVersion(ctx context.Context) (string, error) {
	return c.run(ctx, "version", "--format", "{{.Server.Version}}")
}

// ComposeVersion returns the Docker Compose plugin version.
func (c *Client) ComposeVersion(ctx context.Context) (string, error) {
	return c.run(ctx, "compose", "version", "--short")
}

// Run executes a Docker CLI command using the configured sudo mode.
func (c *Client) Run(ctx context.Context, args ...string) (string, error) {
	return c.run(ctx, args...)
}

func isPermissionDenied(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "permission denied")
}
