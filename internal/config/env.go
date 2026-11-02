package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var envKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// WriteEnvFile creates a new environment file with restrictive permissions.
// It refuses to overwrite an existing file to protect stored secrets.
func WriteEnvFile(path string, values map[string]string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("environment file path cannot be empty")
	}

	if len(values) == 0 {
		return fmt.Errorf("environment configuration cannot be empty")
	}

	for key, value := range values {
		if !envKeyPattern.MatchString(key) {
			return fmt.Errorf("invalid environment variable name: %q", key)
		}

		if strings.ContainsAny(value, "\r\n\x00") {
			return fmt.Errorf(
				"environment variable %s contains invalid characters",
				key,
			)
		}
	}

	directory := filepath.Dir(path)

	if err := os.MkdirAll(directory, 0700); err != nil {
		return fmt.Errorf("creating environment directory: %w", err)
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var content strings.Builder

	for _, key := range keys {
		fmt.Fprintf(&content, "%s=%s\n", key, values[key])
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0600,
	)
	if err != nil {
		return fmt.Errorf("creating environment file: %w", err)
	}

	_, writeErr := file.WriteString(content.String())
	closeErr := file.Close()

	if writeErr != nil {
		os.Remove(path)
		return fmt.Errorf("writing environment file: %w", writeErr)
	}

	if closeErr != nil {
		os.Remove(path)
		return fmt.Errorf("closing environment file: %w", closeErr)
	}

	if err := os.Chmod(path, 0600); err != nil {
		os.Remove(path)
		return fmt.Errorf("setting environment file permissions: %w", err)
	}

	return nil
}

// ReadEnvValue retrieves a single value from a generated environment file.
func ReadEnvValue(path, key string) (string, error) {
	if !envKeyPattern.MatchString(key) {
		return "", fmt.Errorf("invalid environment variable name: %q", key)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening environment file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var value string
	found := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, currentValue, ok := strings.Cut(line, "=")
		if !ok {
			return "", fmt.Errorf("invalid environment file entry")
		}

		if name != key {
			continue
		}

		if found {
			return "", fmt.Errorf("duplicate environment variable: %s", key)
		}

		value = currentValue
		found = true
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading environment file: %w", err)
	}

	if !found || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("environment variable %s is missing or empty", key)
	}

	return value, nil
}
