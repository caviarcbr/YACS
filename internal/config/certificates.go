package config

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	certificateDirMode = 0710
	certificateMode    = 0644
	privateKeyMode     = 0640
)

type CertificatePaths struct {
	Certificate string
	PrivateKey  string
}

func setCertificatePermissions(paths CertificatePaths, directory string) error {
	if err := os.Chmod(directory, certificateDirMode); err != nil {
		return fmt.Errorf("setting certificate directory permissions: %w", err)
	}

	if err := os.Chmod(paths.Certificate, certificateMode); err != nil {
		return fmt.Errorf("setting certificate permissions: %w", err)
	}

	if err := os.Chmod(paths.PrivateKey, privateKeyMode); err != nil {
		return fmt.Errorf("setting private key permissions: %w", err)
	}

	return nil
}

// PrepareCertificates generates a self-signed TLS certificate without
// overwriting an existing certificate pair.
func PrepareCertificates(directory, host string) (CertificatePaths, error) {
	if host == "" {
		return CertificatePaths{}, fmt.Errorf("certificate host cannot be empty")
	}

	paths := CertificatePaths{
		Certificate: filepath.Join(directory, "server.crt"),
		PrivateKey:  filepath.Join(directory, "server.key"),
	}

	certExists := fileExists(paths.Certificate)
	keyExists := fileExists(paths.PrivateKey)

	if certExists && keyExists {
		if err := setCertificatePermissions(paths, directory); err != nil {
			return CertificatePaths{}, err
		}

		return paths, nil
	}

	if certExists || keyExists {
		return CertificatePaths{}, fmt.Errorf(
			"incomplete TLS certificate pair: both server.crt and server.key must exist",
		)
	}

	openssl, err := exec.LookPath("openssl")
	if err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"OpenSSL is required to generate TLS certificates",
		)
	}

	if err := os.MkdirAll(directory, 0700); err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"creating certificate directory: %w", err,
		)
	}

	tempCert := filepath.Join(directory, ".server.crt.tmp")
	tempKey := filepath.Join(directory, ".server.key.tmp")

	defer os.Remove(tempCert)
	defer os.Remove(tempKey)

	san := "DNS:localhost,IP:127.0.0.1"

	if ip := net.ParseIP(host); ip != nil {
		san += ",IP:" + ip.String()
	} else {
		san += ",DNS:" + host
	}

	command := exec.Command(
		openssl,
		"req",
		"-x509",
		"-nodes",
		"-newkey", "rsa:3072",
		"-sha256",
		"-days", "397",
		"-keyout", tempKey,
		"-out", tempCert,
		"-subj", "/CN="+host,
		"-addext", "subjectAltName="+san,
	)

	output, err := command.CombinedOutput()
	if err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"generating TLS certificate: %s: %w",
			string(output),
			err,
		)
	}

	if err := os.Chmod(tempKey, 0600); err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"setting private key permissions: %w", err,
		)
	}

	if err := os.Chmod(tempCert, 0644); err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"setting certificate permissions: %w", err,
		)
	}

	if err := os.Rename(tempKey, paths.PrivateKey); err != nil {
		return CertificatePaths{}, fmt.Errorf(
			"installing private key: %w", err,
		)
	}

	if err := os.Rename(tempCert, paths.Certificate); err != nil {
		os.Remove(paths.PrivateKey)
		return CertificatePaths{}, fmt.Errorf(
			"installing certificate: %w", err,
		)
	}

	if err := setCertificatePermissions(paths, directory); err != nil {
		return CertificatePaths{}, err
	}

	return paths, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
