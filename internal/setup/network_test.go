package setup

import (
	"strings"
	"testing"
)

func TestValidateHost(t *testing.T) {
	tests := []struct {
		name  string
		host  string
		valid bool
	}{
		{"valid IPv4", "192.168.1.100", true},
		{"valid hostname", "cloudbox", true},
		{"valid domain", "cloud.example.com", true},
		{"valid localhost", "localhost", true},
		{"empty host", "", false},
		{"invalid IPv4", "192.168.1.999", false},
		{"invalid characters", "cloud box", false},
		{"invalid punctuation", "-cloudbox", false},
		{"trailing hyphen", "cloudbox-", false},
		{"URL instead of hostname", "https://cloud.example.com", false},
		{"consecutive dots", "cloud..example.com", false},
		{"leading dot", ".cloudbox", false},
		{"trailing dot", "cloudbox.", false},
		{"hyphen at start of label", "cloud.-server", false},
		{"hyphen at end of label", "cloud-.server", false},
		{"valid internal hyphen", "my-cloudbox.local", true},
		{"valid single-label hostname", "myserver", true},
		{"valid maximum-length label", strings.Repeat("a", 63), true},
		{"label too long", strings.Repeat("a", 64), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateHost(tt.host)

			if got != tt.valid {
				t.Errorf(
					"ValidateHost(%q) = %v, want %v",
					tt.host,
					got,
					tt.valid,
				)
			}
		})
	}
}

func TestNewNetworkConfig(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		wantHost  string
		wantError bool
	}{
		{
			name:     "IPv4 address",
			host:     "192.168.1.100",
			wantHost: "192.168.1.100",
		},
		{
			name:     "hostname",
			host:     "cloudbox",
			wantHost: "cloudbox",
		},
		{
			name:      "invalid host",
			host:      "invalid host",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNetworkConfig(tt.host)

			if tt.wantError {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.PublicHost != tt.wantHost {
				t.Errorf(
					"PublicHost = %q, want %q",
					got.PublicHost,
					tt.wantHost,
				)
			}

			if got.BindAddress == "" {
				t.Error("BindAddress should not be empty")
			}
		})
	}
}
