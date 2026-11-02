package config

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareCertificates(t *testing.T) {
	directory := t.TempDir()
	host := "cloud.example.test"

	paths, err := PrepareCertificates(directory, host)
	if err != nil {
		t.Fatalf("PrepareCertificates() error = %v", err)
	}

	certData, err := os.ReadFile(paths.Certificate)
	if err != nil {
		t.Fatalf("reading certificate: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		t.Fatal("certificate is not valid PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parsing certificate: %v", err)
	}

	if cert.Subject.CommonName != host {
		t.Errorf("certificate CN = %q, want %q", cert.Subject.CommonName, host)
	}

	if err := cert.VerifyHostname(host); err != nil {
		t.Errorf("certificate does not cover configured host: %v", err)
	}

	if err := cert.VerifyHostname("localhost"); err != nil {
		t.Errorf("certificate does not cover localhost: %v", err)
	}

	if err := cert.VerifyHostname("127.0.0.1"); err != nil {
		t.Errorf("certificate does not cover 127.0.0.1: %v", err)
	}

	keyInfo, err := os.Stat(paths.PrivateKey)
	if err != nil {
		t.Fatalf("reading private key metadata: %v", err)
	}

	if keyInfo.Mode().Perm() != 0640 {
		t.Errorf("private key permissions = %04o, want 0640", keyInfo.Mode().Perm())
	}

	certInfo, err := os.Stat(paths.Certificate)
	if err != nil {
		t.Fatalf("reading certificate metadata: %v", err)
	}

	if certInfo.Mode().Perm() != 0644 {
		t.Errorf("certificate permissions = %04o, want 0644", certInfo.Mode().Perm())
	}
}

func TestPrepareCertificatesPreservesExistingPair(t *testing.T) {
	directory := t.TempDir()

	certPath := filepath.Join(directory, "server.crt")
	keyPath := filepath.Join(directory, "server.key")

	originalCert := []byte("existing certificate")
	originalKey := []byte("existing private key")

	if err := os.WriteFile(certPath, originalCert, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(keyPath, originalKey, 0600); err != nil {
		t.Fatal(err)
	}

	_, err := PrepareCertificates(directory, "cloud.example.test")
	if err != nil {
		t.Fatalf("PrepareCertificates() error = %v", err)
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(certData) != string(originalCert) {
		t.Error("existing certificate was modified")
	}

	if string(keyData) != string(originalKey) {
		t.Error("existing private key was modified")
	}
}

func TestPrepareCertificatesRejectsIncompletePair(t *testing.T) {
	directory := t.TempDir()

	certPath := filepath.Join(directory, "server.crt")

	if err := os.WriteFile(certPath, []byte("existing certificate"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := PrepareCertificates(directory, "cloud.example.test")
	if err == nil {
		t.Fatal("expected error for incomplete certificate pair")
	}
}

func TestPrepareCertificatesRejectsEmptyHost(t *testing.T) {
	_, err := PrepareCertificates(t.TempDir(), "")
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}
