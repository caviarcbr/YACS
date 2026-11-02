package compose

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DockerRunner interface {
	Run(ctx context.Context, args ...string) (string, error)
}

type StreamingDockerRunner interface {
	RunStream(ctx context.Context, output io.Writer, args ...string) error
}

// Client manages Docker Compose deployments.

type Client struct {
	Docker      DockerRunner
	ComposeFile string
	ProjectName string
	EnvFile     string
}

// NewClient creates a Compose client.
func NewClient(dockerClient DockerRunner, composeFile, projectName string) (*Client, error) {
	if dockerClient == nil {
		return nil, fmt.Errorf("Docker client cannot be nil")
	}

	if strings.TrimSpace(composeFile) == "" {
		return nil, fmt.Errorf("Compose file path cannot be empty")
	}

	if strings.TrimSpace(projectName) == "" {
		return nil, fmt.Errorf("Compose project name cannot be empty")
	}

	return &Client{
		Docker:      dockerClient,
		ComposeFile: filepath.Clean(composeFile),
		ProjectName: projectName,
	}, nil
}

// SetEnvFile configures an optional environment file for Compose.
func (c *Client) SetEnvFile(path string) {
	if strings.TrimSpace(path) == "" {
		c.EnvFile = ""
		return
	}

	c.EnvFile = filepath.Clean(path)
}

// Validate checks whether the Compose file is valid.
func (c *Client) Validate(ctx context.Context) error {
	if _, err := os.Stat(c.ComposeFile); err != nil {
		return fmt.Errorf("accessing Compose file: %w", err)
	}

	if c.EnvFile != "" {
		if _, err := os.Stat(c.EnvFile); err != nil {
			return fmt.Errorf("accessing environment file: %w", err)
		}
	}

	_, err := c.run(ctx, "config", "--quiet")
	if err != nil {
		return fmt.Errorf("validating Compose configuration: %w", err)
	}

	return nil
}

// PullStream downloads Compose images and streams Docker output.
func (c *Client) PullStream(ctx context.Context, output io.Writer) error {
	if runner, ok := c.Docker.(StreamingDockerRunner); ok {
		composeArgs := []string{
			"compose", "-f", c.ComposeFile, "-p", c.ProjectName,
		}

		if c.EnvFile != "" {
			composeArgs = append(composeArgs, "--env-file", c.EnvFile)
		}

		composeArgs = append(composeArgs, "pull")

		if err := runner.RunStream(ctx, output, composeArgs...); err != nil {
			return fmt.Errorf("pulling Compose images: %w", err)
		}

		return nil
	}

	result, err := c.run(ctx, "pull")
	if result != "" {
		_, _ = io.WriteString(output, result+"\n")
	}
	if err != nil {
		return fmt.Errorf("pulling Compose images: %w", err)
	}

	return nil
}

func (c *Client) InstallStream(
	ctx context.Context,
	output io.Writer,
	progress func(string),
) error {
	if progress != nil {
		progress("Validating Compose configuration...")
	}

	if err := c.Validate(ctx); err != nil {
		return fmt.Errorf("validating Compose configuration: %w", err)
	}

	if progress != nil {
		progress("Pulling Docker images...")
	}

	if err := c.PullStream(ctx, output); err != nil {
		return fmt.Errorf("pulling Docker images: %w", err)
	}

	if progress != nil {
		progress("Starting services...")
	}

	if err := c.UpStream(ctx, output); err != nil {
		return fmt.Errorf("starting services: %w", err)
	}

	return nil
}

// UpStream starts the Compose deployment and streams Docker output.
func (c *Client) UpStream(ctx context.Context, output io.Writer) error {
	if runner, ok := c.Docker.(StreamingDockerRunner); ok {
		composeArgs := []string{
			"compose", "-f", c.ComposeFile, "-p", c.ProjectName,
		}

		if c.EnvFile != "" {
			composeArgs = append(composeArgs, "--env-file", c.EnvFile)
		}

		composeArgs = append(composeArgs, "up", "-d")

		if err := runner.RunStream(ctx, output, composeArgs...); err != nil {
			return fmt.Errorf("starting Compose deployment: %w", err)
		}

		return nil
	}

	result, err := c.run(ctx, "up", "-d")
	if result != "" {
		_, _ = io.WriteString(output, result+"\n")
	}
	if err != nil {
		return fmt.Errorf("starting Compose deployment: %w", err)
	}

	return nil
}

// Uninstall stops and removes the Compose deployment.
// When removeData is true, named volumes are also removed.
func (c *Client) Uninstall(ctx context.Context, removeData bool) error {
	args := []string{"down", "--remove-orphans"}

	if removeData {
		args = append(args, "--volumes")
	}

	_, err := c.run(ctx, args...)
	if err != nil {
		return fmt.Errorf("uninstalling Compose deployment: %w", err)
	}

	return nil
}

// run executes a Docker Compose command.
func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	composeArgs := []string{"compose", "-f", c.ComposeFile, "-p", c.ProjectName}

	if c.EnvFile != "" {
		composeArgs = append(composeArgs, "--env-file", c.EnvFile)
	}

	composeArgs = append(composeArgs, args...)

	// Reuse the Docker client's command execution and sudo fallback.
	return c.Docker.Run(ctx, composeArgs...)
}
