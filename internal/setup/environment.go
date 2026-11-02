package setup

// EnvironmentValues converts the wizard configuration into environment
// variables consumed by the Compose deployments.
func EnvironmentValues(
	selection ServiceSelection,
	network NetworkConfig,
	security SecurityConfig,
) map[string]string {
	values := map[string]string{
		"PUBLIC_HOST":        network.PublicHost,
		"BIND_ADDRESS":       network.BindAddress,
		"OCIS_JWT_SECRET":    security.OCISJWTSecret,
		"IDM_ADMIN_PASSWORD": security.AdminPassword,
	}

	if selection == OCISWithOnlyOffice {
		values["ONLYOFFICE_JWT_SECRET"] = security.OnlyOfficeJWTSecret
		values["COLLABORATION_WOPI_SECRET"] = security.CollaborationWOPISecret
	}

	return values
}
