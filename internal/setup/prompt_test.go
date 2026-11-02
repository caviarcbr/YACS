package setup

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestChooseDefaultsToFirstOption(t *testing.T) {
	var out bytes.Buffer
	p := NewPrompter(strings.NewReader("\n"), &out)

	index, err := p.Choose("Pick one:", []string{"a", "b"})
	if err != nil {
		t.Fatalf("Choose() error: %v", err)
	}

	if index != 0 {
		t.Errorf("expected index 0, got %d", index)
	}
}

func TestChooseRetriesInvalidInput(t *testing.T) {
	var out bytes.Buffer
	p := NewPrompter(strings.NewReader("9\nx\n2\n"), &out)

	index, err := p.Choose("Pick one:", []string{"a", "b"})
	if err != nil {
		t.Fatalf("Choose() error: %v", err)
	}

	if index != 1 {
		t.Errorf("expected index 1, got %d", index)
	}

	if strings.Count(out.String(), "Enter a number from 1 to 2.") != 2 {
		t.Errorf("expected two retry messages, got output:\n%s", out.String())
	}
}

func TestAskUsesDefaultAndRetriesInvalidValue(t *testing.T) {
	var out bytes.Buffer
	p := NewPrompter(strings.NewReader("bad host!\n\n"), &out)

	value, err := p.Ask("Server address", "192.168.1.10", func(v string) error {
		if !ValidateHost(v) {
			return errors.New("invalid host")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Ask() error: %v", err)
	}

	if value != "192.168.1.10" {
		t.Errorf("expected default value, got %q", value)
	}

	if !strings.Contains(out.String(), "invalid host") {
		t.Errorf("expected validation message in output:\n%s", out.String())
	}
}

func TestConfirmDefaultsToNo(t *testing.T) {
	cases := map[string]bool{
		"\n":    false,
		"n\n":   false,
		"y\n":   true,
		"YES\n": true,
	}

	for input, want := range cases {
		p := NewPrompter(strings.NewReader(input), &bytes.Buffer{})

		got, err := p.Confirm("Continue?")
		if err != nil {
			t.Fatalf("Confirm(%q) error: %v", input, err)
		}

		if got != want {
			t.Errorf("Confirm(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestPromptsReturnErrorWhenInputEnds(t *testing.T) {
	p := NewPrompter(strings.NewReader(""), &bytes.Buffer{})

	if _, err := p.Confirm("Continue?"); !errors.Is(err, ErrInputClosed) {
		t.Errorf("expected ErrInputClosed, got %v", err)
	}
}

func TestPromptInstallationCollectsAnswers(t *testing.T) {
	p := NewPrompter(strings.NewReader("2\n10.0.0.5\nCorrect-Horse-42\nCorrect-Horse-42\n"), &bytes.Buffer{})

	installation, err := PromptInstallation(p, "192.168.1.1")
	if err != nil {
		t.Fatalf("PromptInstallation() error: %v", err)
	}

	if installation.Services != OCISWithOnlyOffice {
		t.Errorf("expected ONLYOFFICE selection, got %v", installation.Services)
	}

	if installation.Network.PublicHost != "10.0.0.5" {
		t.Errorf("expected host 10.0.0.5, got %q", installation.Network.PublicHost)
	}

	if installation.AdminPassword != "Correct-Horse-42" {
		t.Errorf("unexpected admin password %q", installation.AdminPassword)
	}
}

func TestAskPasswordRetriesMismatch(t *testing.T) {
	var out bytes.Buffer
	p := NewPrompter(strings.NewReader("Correct-Horse-42\nCorrect-Horse-43\nCorrect-Horse-42\nCorrect-Horse-42\n"), &out)

	password, err := p.AskPassword("Admin password", ValidateAdminPassword)
	if err != nil {
		t.Fatalf("AskPassword() error: %v", err)
	}

	if password != "Correct-Horse-42" {
		t.Errorf("unexpected password %q", password)
	}

	if !strings.Contains(out.String(), "The passwords do not match") {
		t.Errorf("expected mismatch message in output:\n%s", out.String())
	}
}

func TestValidateAdminPassword(t *testing.T) {
	cases := map[string]bool{
		"Correct-Horse-42":       true,
		"short1":                 false,
		"has space in it 12":     false,
		"dollar$sign-12345":      false,
		"hash#sign-1234567":      false,
		"quote'sign-1234567":     false,
		"back\\slash-12345":      false,
		"tick`mark-1234567":      false,
		strings.Repeat("a", 129): false,
	}

	for password, valid := range cases {
		err := ValidateAdminPassword(password)
		if valid && err != nil {
			t.Errorf("expected %q to be valid, got %v", password, err)
		}

		if !valid && err == nil {
			t.Errorf("expected %q to be rejected", password)
		}
	}
}
