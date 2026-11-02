package setup

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

type NetworkConfig struct {
	PublicHost  string
	BindAddress string
}

var hostnamePattern = regexp.MustCompile(
	`^([A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?)(\.([A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?))*$`,
)

// DetectDefaultHost finds the preferred LAN IPv4 address.
// It falls back to localhost if detection fails.
func DetectDefaultHost() string {
	output, err := exec.Command(
		"ip",
		"-4",
		"route",
		"get",
		"1.1.1.1",
	).Output()
	if err != nil {
		return "localhost"
	}

	fields := strings.Fields(string(output))

	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "src" {
			ip := net.ParseIP(fields[i+1])
			if ip != nil && ip.To4() != nil {
				return ip.String()
			}
		}
	}

	return "localhost"
}

// DetectDefaultBindAddress finds the preferred local IPv4 address.
// It falls back to loopback if detection fails.
func DetectDefaultBindAddress() string {
	host := DetectDefaultHost()

	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil {
		return ip.String()
	}

	return "127.0.0.1"
}

// ValidateHost accepts IPv4 addresses and simple hostnames.
func ValidateHost(value string) bool {
	value = strings.TrimSpace(value)

	if ip := net.ParseIP(value); ip != nil && ip.To4() != nil {
		return true
	}

	if len(value) == 0 || len(value) > 253 {
		return false
	}

	// Reject invalid IPv4-looking addresses instead of treating
	// them as hostnames.
	if strings.Count(value, ".") == 3 {
		parts := strings.Split(value, ".")
		allNumeric := true

		for _, part := range parts {
			if part == "" {
				allNumeric = false
				break
			}

			for _, char := range part {
				if char < '0' || char > '9' {
					allNumeric = false
					break
				}
			}
		}

		if allNumeric {
			return false
		}
	}

	return hostnamePattern.MatchString(value)
}

// NewNetworkConfig creates a network configuration from a user-selected host.
func NewNetworkConfig(host string) (NetworkConfig, error) {
	host = strings.TrimSpace(host)

	if !ValidateHost(host) {
		return NetworkConfig{}, fmt.Errorf(
			"enter a valid IPv4 address or hostname",
		)
	}

	bindAddress := host

	if net.ParseIP(host) == nil {
		bindAddress = DetectDefaultBindAddress()
	}

	return NetworkConfig{
		PublicHost:  host,
		BindAddress: bindAddress,
	}, nil
}
