package compose

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	composeassets "github.com/caviarcbr/YACS/compose"
)

type deploymentTemplateData struct {
	IncludeOnlyOffice bool
	CSPPath           string
	CertificateDir    string
}

// GenerateDeployment creates a Compose deployment from the embedded template.
// It refuses to overwrite an existing deployment.
func GenerateDeployment(
	composeDir string,
	name string,
	configDir string,
	includeOnlyOffice bool,
) (Deployment, error) {
	if strings.TrimSpace(name) == "" ||
		name == "." ||
		name == ".." ||
		filepath.Base(name) != name {
		return Deployment{}, fmt.Errorf("invalid deployment name: %q", name)
	}

	if strings.TrimSpace(configDir) == "" {
		return Deployment{}, fmt.Errorf("configuration directory cannot be empty")
	}

	directory := filepath.Join(composeDir, name)

	if err := os.MkdirAll(composeDir, 0755); err != nil {
		return Deployment{}, fmt.Errorf("creating Compose directory: %w", err)
	}

	if err := os.Mkdir(directory, 0755); err != nil {
		return Deployment{}, fmt.Errorf(
			"creating deployment directory: %w", err,
		)
	}

	cleanup := true
	defer func() {
		if cleanup {
			os.RemoveAll(directory)
		}
	}()

	templateContent, err := composeassets.Files.ReadFile("compose.yml")
	if err != nil {
		return Deployment{}, fmt.Errorf(
			"reading embedded Compose template: %w", err,
		)
	}

	tmpl, err := template.New("compose").Parse(string(templateContent))
	if err != nil {
		return Deployment{}, fmt.Errorf(
			"parsing Compose template: %w", err,
		)
	}

	data := deploymentTemplateData{
		IncludeOnlyOffice: includeOnlyOffice,
		CSPPath:           filepath.Join(directory, "csp.yaml"),
		CertificateDir:    filepath.Join(configDir, "certs"),
	}

	var rendered bytes.Buffer

	if err := tmpl.Execute(&rendered, data); err != nil {
		return Deployment{}, fmt.Errorf(
			"rendering Compose template: %w", err,
		)
	}

	composeFile := filepath.Join(directory, "compose.yaml")

	if err := os.WriteFile(composeFile, rendered.Bytes(), 0644); err != nil {
		return Deployment{}, fmt.Errorf(
			"writing Compose file: %w", err,
		)
	}

	cspContent, err := composeassets.Files.ReadFile("csp.yaml")
	if err != nil {
		return Deployment{}, fmt.Errorf(
			"reading embedded CSP configuration: %w", err,
		)
	}

	cspFile := filepath.Join(directory, "csp.yaml")

	if err := os.WriteFile(cspFile, cspContent, 0644); err != nil {
		return Deployment{}, fmt.Errorf(
			"writing CSP configuration: %w", err,
		)
	}

	cleanup = false

	return Deployment{
		Name:        name,
		Directory:   directory,
		ComposeFile: composeFile,
	}, nil
}
