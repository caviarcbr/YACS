package setup

import "testing"

func TestEnvironmentValuesOCISOnly(t *testing.T) {
	network := NetworkConfig{
		PublicHost:  "cloudbox",
		BindAddress: "192.168.1.100",
	}

	security := SecurityConfig{
		OCISJWTSecret: "ocis-secret",
	}

	values := EnvironmentValues(OCISOnly, network, security)

	if values["PUBLIC_HOST"] != "cloudbox" {
		t.Errorf("unexpected PUBLIC_HOST: %q", values["PUBLIC_HOST"])
	}

	if values["BIND_ADDRESS"] != "192.168.1.100" {
		t.Errorf("unexpected BIND_ADDRESS: %q", values["BIND_ADDRESS"])
	}

	if values["OCIS_JWT_SECRET"] != "ocis-secret" {
		t.Error("OCIS_JWT_SECRET was not included")
	}

	if _, exists := values["ONLYOFFICE_JWT_SECRET"]; exists {
		t.Error("ONLYOFFICE_JWT_SECRET should not be included")
	}

	if _, exists := values["COLLABORATION_WOPI_SECRET"]; exists {
		t.Error("COLLABORATION_WOPI_SECRET should not be included")
	}
}

func TestEnvironmentValuesWithOnlyOffice(t *testing.T) {
	security := SecurityConfig{
		OCISJWTSecret:           "ocis-secret",
		OnlyOfficeJWTSecret:     "onlyoffice-secret",
		CollaborationWOPISecret: "wopi-secret",
	}

	values := EnvironmentValues(
		OCISWithOnlyOffice,
		NetworkConfig{
			PublicHost:  "cloudbox",
			BindAddress: "192.168.1.100",
		},
		security,
	)

	expected := map[string]string{
		"OCIS_JWT_SECRET":           "ocis-secret",
		"ONLYOFFICE_JWT_SECRET":     "onlyoffice-secret",
		"COLLABORATION_WOPI_SECRET": "wopi-secret",
	}

	for key, want := range expected {
		if values[key] != want {
			t.Errorf("%s = %q, want %q", key, values[key], want)
		}
	}
}
