package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/caviarcbr/YACS/internal/setup"
)

type installStep int

const (
	installService installStep = iota
	installHost
	installPassword
	installConfirmPassword
	installReview
)

type installModel struct {
	step installStep

	serviceCursor int
	reviewCursor  int
	confirmed     bool
	services      setup.ServiceSelection
	network       setup.NetworkConfig

	hostInput            textinput.Model
	passwordInput        textinput.Model
	confirmPasswordInput textinput.Model

	password string
	err      error
	status   string
}

func newInstallModel() installModel {
	host := textinput.New()
	host.Prompt = "Server address: "
	host.SetValue(setup.DetectDefaultHost())

	password := textinput.New()
	password.Prompt = "Admin password: "
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 128

	confirmPassword := textinput.New()
	confirmPassword.Prompt = "Confirm password: "
	confirmPassword.EchoMode = textinput.EchoPassword
	confirmPassword.CharLimit = 128

	return installModel{
		step:                 installService,
		services:             setup.OCISOnly,
		hostInput:            host,
		passwordInput:        password,
		confirmPasswordInput: confirmPassword,
	}
}

func (m installModel) Update(msg tea.Msg) (installModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.step == installService && m.serviceCursor > 0 {
				m.serviceCursor--
			}

			if m.step == installReview && m.reviewCursor > 0 {
				m.reviewCursor--
			}

		case "down", "j":
			if m.step == installService && m.serviceCursor < 1 {
				m.serviceCursor++
			}

			if m.step == installReview && m.reviewCursor < 1 {
				m.reviewCursor++
			}

		case "enter":
			switch m.step {
			case installService:
				if m.serviceCursor == 0 {
					m.services = setup.OCISOnly
				} else {
					m.services = setup.OCISWithOnlyOffice
				}

				m.step = installHost
				return m, m.hostInput.Focus()

			case installHost:
				network, err := setup.NewNetworkConfig(
					strings.TrimSpace(m.hostInput.Value()),
				)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.network = network
				m.err = nil
				m.step = installPassword

				m.hostInput.Blur()
				return m, m.passwordInput.Focus()

			case installPassword:
				password := m.passwordInput.Value()

				if err := setup.ValidateAdminPassword(password); err != nil {
					m.err = err
					return m, nil
				}

				m.password = password
				m.err = nil
				m.step = installConfirmPassword

				m.passwordInput.Blur()
				return m, m.confirmPasswordInput.Focus()

			case installConfirmPassword:
				if m.confirmPasswordInput.Value() != m.password {
					m.err = fmt.Errorf("passwords do not match")
					return m, nil
				}

				m.err = nil
				m.step = installReview

				m.confirmPasswordInput.Blur()
				return m, nil

			case installReview:
				if m.reviewCursor == 0 {
					m.confirmed = true
					return m, nil
				}

				m.step = installConfirmPassword
				m.err = nil
				m.confirmPasswordInput.SetValue("")
				return m, m.confirmPasswordInput.Focus()
			}

		case "esc":
			if m.step > installService {
				m.step--
				m.status = ""
				m.err = nil

				switch m.step {
				case installService:
					return m, nil
				case installHost:
					return m, m.hostInput.Focus()
				case installPassword:
					return m, m.passwordInput.Focus()
				case installConfirmPassword:
					return m, m.confirmPasswordInput.Focus()
				case installReview:
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		// The parent TUI handles terminal dimensions.
	}

	switch m.step {
	case installHost:
		var cmd tea.Cmd
		m.hostInput, cmd = m.hostInput.Update(msg)
		return m, cmd

	case installPassword:
		var cmd tea.Cmd
		m.passwordInput, cmd = m.passwordInput.Update(msg)
		return m, cmd

	case installConfirmPassword:
		var cmd tea.Cmd
		m.confirmPasswordInput, cmd = m.confirmPasswordInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m installModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Install YACS"))
	b.WriteString("\n\n")

	switch m.step {
	case installService:
		b.WriteString("Choose the services you want to install.\n\n")

		options := []string{
			"ownCloud Infinite Scale",
			"ownCloud Infinite Scale + ONLYOFFICE",
		}

		for i, option := range options {
			cursor := "  "
			style := menuStyle

			if m.serviceCursor == i {
				cursor = "› "
				style = selectedMenuStyle
			}

			b.WriteString(style.Render(cursor + option))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Use ↑/↓ to choose a service and Enter to continue.",
		))

	case installHost:
		b.WriteString("Configure your server address.\n\n")
		b.WriteString(m.hostInput.View())
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Enter a hostname or IP address. Press Enter to continue.",
		))

	case installPassword:
		b.WriteString("Set your administrator password.\n\n")
		b.WriteString(m.passwordInput.View())
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Use 12–128 characters. Press Enter to continue.",
		))

	case installConfirmPassword:
		b.WriteString("Confirm your administrator password.\n\n")
		b.WriteString(m.confirmPasswordInput.View())
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Re-enter your password to confirm it.",
		))

	case installReview:
		b.WriteString("Review your installation settings.\n\n")

		b.WriteString(descriptionStyle.Render("SELECTED SERVICES"))
		b.WriteString("\n")
		b.WriteString(m.services.String())
		b.WriteString("\n\n")

		b.WriteString(descriptionStyle.Render("SERVER ADDRESS"))
		b.WriteString("\n")
		b.WriteString(m.network.PublicHost)
		b.WriteString("\n\n")

		b.WriteString(descriptionStyle.Render("SERVER URL"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf(
			"https://%s:9200",
			m.network.PublicHost,
		))
		b.WriteString("\n\n")

		b.WriteString(descriptionStyle.Render("ADMINISTRATOR"))
		b.WriteString("\n")
		b.WriteString("Username: admin\n")
		b.WriteString("Password: validated and saved for installation\n\n")

		b.WriteString(descriptionStyle.Render(
			"Your password will not be displayed.",
		))
		b.WriteString("\n\n")

		options := []string{
			"Install",
			"Go Back",
		}

		for i, option := range options {
			cursor := "  "
			style := menuStyle

			if m.reviewCursor == i {
				cursor = "› "
				style = selectedMenuStyle
			}

			b.WriteString(style.Render(cursor + option))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Review your settings before starting the installation.",
		))
	}
	b.WriteString(fmt.Sprintf("\n\n%s",
		footerStyle.Render("↑/↓ Navigate  •  Enter Select  •  Esc Back"),
	))

	return b.String()
}
