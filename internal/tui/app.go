package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/caviarcbr/YACS/internal/cli"
	"github.com/caviarcbr/YACS/internal/setup"
)

var installStages = []string{
	"Checking prerequisites...",
	"Preparing TLS certificates...",
	"Saving installation configuration...",
	"Generating Docker Compose deployment...",
	"Validating Compose configuration...",
	"Pulling Docker images...",
	"Starting services...",
}

type model struct {
	cursor              int
	width               int
	height              int
	status              string
	spinner             spinner.Model
	screen              string
	installer           installModel
	installing          bool
	installStage        string
	installOutput       string
	installErr          error
	uninstaller         uninstallModel
	uninstalling        bool
	uninstallErr        error
	uninstallRemoveData bool
	sudoForUninstall    bool
	uninstallOutput     string
	restoring           bool
	restoreStage        string
	restoreOutput       string
	restoreInfo         cli.RestoreInfo
	restoreErr          error
	sudoForRestore      bool
	sudo                sudoModel
	pendingInstallation setup.Installation
}

func Run() error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	program := tea.NewProgram(model{
		spinner: s,
	})

	_, err := program.Run()
	return err
}
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case spinner.TickMsg:
		if !m.installing &&
			!m.uninstalling &&
			!m.restoring &&
			m.screen != "checking-restore" {
			return m, nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case dockerAccessMsg:
		if msg.err != nil {
			m.installStage = ""
			m.installErr = msg.err
			m.screen = "install-result"
			return m, nil
		}

		if msg.needsSudo {
			m.sudoForUninstall = false
			m.sudo = newSudoModel()
			m.screen = "sudo"
			return m, nil
		}

		m.installing = true
		m.screen = "installing"
		m.installOutput = ""
		m.installErr = nil
		m.installStage = ""

		return m, tea.Batch(
			m.spinner.Tick,
			installConfiguredCmd(m.pendingInstallation),
		)

	case sudoAuthMsg:
		if msg.err != nil {
			m.installStage = ""
			m.sudo.SetError(msg.err)
			m.screen = "sudo"
			return m, m.sudo.Focus()
		}

		m.sudo.ClearPassword()
		m.sudo.Blur()
		m.installing = true
		m.screen = "installing"
		m.installOutput = ""
		m.installErr = nil
		m.installStage = ""

		return m, tea.Batch(
			m.spinner.Tick,
			installConfiguredCmd(m.pendingInstallation),
		)
	case installProgressMsg:
		m.installStage = msg.stage
		return m, waitForInstallEvent(msg.events)

	case installFinishedMsg:
		m.installing = false
		m.installStage = ""
		m.installOutput = msg.output
		m.installErr = msg.err
		m.screen = "install-result"
		return m, nil

	case restoreProgressMsg:
		m.restoreStage = msg.stage
		return m, waitForRestoreEvent(msg.events)

	case restoreFinishedMsg:
		m.restoring = false
		m.restoreStage = ""
		m.restoreOutput = msg.output
		m.restoreErr = msg.err
		m.screen = "restore-result"
		return m, nil

	case restoreInspectMsg:
		if msg.err != nil {
			m.restoreErr = msg.err
			m.screen = "restore-result"
			return m, nil
		}

		m.restoreInfo = msg.info
		m.restoreErr = nil
		m.screen = "restore-review"
		return m, nil

	case restoreDockerAccessMsg:
		if msg.err != nil {
			m.restoreErr = msg.err
			m.screen = "restore-result"
			return m, nil
		}

		if msg.needsSudo {
			m.sudoForRestore = true
			m.sudoForUninstall = false
			m.sudo = newSudoModel()
			m.screen = "sudo"
			return m, nil
		}

		m.restoring = true
		m.restoreOutput = ""
		m.restoreErr = nil
		m.restoreStage = ""
		m.screen = "restoring"

		return m, tea.Batch(
			m.spinner.Tick,
			startRestoreCmd(),
		)

	case uninstallDockerAccessMsg:
		if msg.err != nil {
			m.uninstallErr = msg.err
			m.screen = "uninstall-result"
			return m, nil
		}

		if msg.needsSudo {
			m.sudoForUninstall = true
			m.sudo = newSudoModel()
			m.screen = "sudo"
			return m, nil
		}

		m.uninstalling = true
		m.uninstallOutput = ""
		m.uninstallErr = nil
		m.screen = "uninstalling"

		return m, tea.Batch(
			m.spinner.Tick,
			uninstallConfiguredCmd(m.uninstallRemoveData),
		)

	case restoreSudoAuthMsg:
		if msg.err != nil {
			m.sudo.SetError(msg.err)
			m.screen = "sudo"
			return m, m.sudo.Focus()
		}

		m.sudo.ClearPassword()
		m.sudo.Blur()

		m.restoring = true
		m.restoreOutput = ""
		m.restoreErr = nil
		m.restoreStage = ""
		m.screen = "restoring"

		return m, tea.Batch(
			m.spinner.Tick,
			startRestoreCmd(),
		)

	case uninstallSudoAuthMsg:
		if msg.err != nil {
			m.sudo.SetError(msg.err)
			m.screen = "sudo"
			return m, m.sudo.Focus()
		}

		m.sudo.ClearPassword()
		m.sudo.Blur()
		m.uninstalling = true
		m.uninstallOutput = ""
		m.uninstallErr = nil
		m.screen = "uninstalling"

		return m, tea.Batch(
			m.spinner.Tick,
			uninstallConfiguredCmd(m.uninstallRemoveData),
		)

	case uninstallFinishedMsg:
		m.uninstalling = false
		m.uninstallOutput = msg.output
		m.uninstallErr = msg.err
		m.screen = "uninstall-result"
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:

		if m.screen == "sudo" {

			switch msg.String() {

			case "esc":
				m.sudo.ClearPassword()
				m.sudo.Blur()

				if m.sudoForRestore {
					m.screen = "restore-review"
				} else if m.sudoForUninstall {
					m.screen = "uninstall"
				} else {
					m.screen = "install"
				}

				return m, nil

			case "enter":
				if m.sudo.Empty() {
					m.sudo.SetError(fmt.Errorf("please enter your sudo password"))
					return m, nil
				}

				password := m.sudo.Password()
				m.sudo.ClearPassword()
				m.sudo.Blur()
				m.screen = "sudo-authenticating"

				if m.sudoForRestore {
					return m, authenticateRestoreSudoCmd(password)
				} else if m.sudoForUninstall {
					return m, authenticateUninstallSudoCmd(password)
				}

				return m, authenticateSudoCmd(password)

			default:
				var cmd tea.Cmd
				m.sudo, cmd = m.sudo.Update(msg)
				return m, cmd
			}
		}

		if m.screen == "sudo-authenticating" ||
			m.screen == "checking-docker" ||
			m.screen == "checking-uninstall-docker" ||
			m.screen == "checking-restore" ||
			m.screen == "checking-restore-docker" {
			return m, nil
		}

		if m.screen == "installing" ||
			m.screen == "uninstalling" ||
			m.screen == "restoring" {
			// Don't allow navigation while deployment is running.
			return m, nil
		}

		if m.screen == "install-result" ||
			m.screen == "uninstall-result" ||
			m.screen == "restore-result" {
			switch msg.String() {
			case "enter", "esc":
				m.screen = ""
				m.status = ""
				return m, nil
			}

			return m, nil
		}

		if m.screen == "restore-review" {
			switch msg.String() {
			case "esc":
				m.screen = ""
				m.restoreErr = nil
				return m, nil

			case "enter":
				m.restoreErr = nil
				m.screen = "checking-restore-docker"
				return m, checkRestoreDockerAccessCmd()
			}

			return m, nil
		}

		if m.screen == "install" {
			if msg.String() == "esc" && m.installer.step == installService {
				m.screen = ""
				return m, nil
			}
			var cmd tea.Cmd
			m.installer, cmd = m.installer.Update(msg)

			if m.installer.confirmed {
				m.installer.confirmed = false

				m.pendingInstallation = setup.Installation{
					Services:      m.installer.services,
					Network:       m.installer.network,
					AdminPassword: m.installer.password,
				}

				m.screen = "checking-docker"
				return m, checkDockerAccessCmd()
			}

			return m, cmd
		}

		if m.screen == "uninstall" {
			if msg.String() == "esc" &&
				m.uninstaller.step == uninstallMode {
				m.screen = ""
				return m, nil
			}

			var cmd tea.Cmd
			m.uninstaller, cmd = m.uninstaller.Update(msg)

			if m.uninstaller.confirmed {
				m.uninstaller.confirmed = false
				m.uninstallRemoveData = m.uninstaller.removeData
				m.screen = "checking-uninstall-docker"
				return m, checkUninstallDockerAccessCmd()
			}

			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}

		case "enter":
			switch m.cursor {
			case 0:
				m.screen = "install"
				m.installer = newInstallModel()
				return m, nil
			case 1:
				m.restoreInfo = cli.RestoreInfo{}
				m.restoreErr = nil
				m.screen = "checking-restore"
				m.status = ""
				return m, inspectRestoreCmd()
			case 2:
				m.screen = "uninstall"
				m.uninstaller = newUninstallModel()
				m.status = ""
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {

	if m.screen == "sudo" {
		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.sudo.View(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "sudo-authenticating" || m.screen == "checking-docker" {
		message := "Checking Docker access..."

		if m.screen == "checking-uninstall-docker" {
			message = "Checking Docker accesss for uninstall..."
		}

		if m.screen == "sudo-authenticating" {
			message = "Verifying sudo credentials..."
		}

		content := titleStyle.Render("Docker access required") +
			"\n\n" + message +
			"\n\n" + descriptionStyle.Render("Please wait.")

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "checking-restore" {
		content := titleStyle.Render("Inspecting existing installation") +
			"\n\n" +
			m.spinner.View() +
			" Looking for an existing YACS deployment..." +
			"\n\n" +
			descriptionStyle.Render("Please wait.")

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "restore-review" {
		var b strings.Builder

		b.WriteString(titleStyle.Render("Restore existing YACS installation"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("Deployment: %s\n", m.restoreInfo.DeploymentName))
		b.WriteString(fmt.Sprintf("Server URL: %s\n", m.restoreInfo.ServerURL))
		b.WriteString("\n")
		b.WriteString("Existing credentials, certificates, and data will be preserved.\n")
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("Enter to continue  •  Esc to cancel"))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "restoring" {
		var b strings.Builder

		b.WriteString(titleStyle.Render("Restoring YACS"))
		b.WriteString("\n\n")

		b.WriteString(m.spinner.View())
		b.WriteString(" Starting your existing deployment...")
		b.WriteString("\n\n")

		if m.restoreStage != "" {
			b.WriteString(descriptionStyle.Render(m.restoreStage))
			b.WriteString("\n\n")
		}

		b.WriteString("Your existing configuration, credentials, certificates, and persistent data are preserved.")
		b.WriteString("\n\n")

		b.WriteString(descriptionStyle.Render("Please wait. Do not close the application."))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "installing" {
		var b strings.Builder

		b.WriteString(titleStyle.Render("Installing YACS"))
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" Working on your personal cloud deployment...")
		b.WriteString("\n\n")
		b.WriteString("Preparing your personal cloud deployment...\n\n")

		currentIndex := -1

		for i, stage := range installStages {
			if stage == m.installStage {
				currentIndex = i
				break
			}
		}

		for i, stage := range installStages {
			marker := "○"
			style := descriptionStyle

			switch {
			case i < currentIndex:
				marker = "✓"
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

			case i == currentIndex:
				marker = "›"
				style = selectedMenuStyle
			}

			b.WriteString(style.Render(marker + " " + strings.TrimSuffix(stage, "...")))
			b.WriteString("\n")
		}

		if currentIndex == 5 {
			b.WriteString("\n")
			b.WriteString(descriptionStyle.Render(
				"Downloading Docker images may take several minutes.",
			))
		}

		b.WriteString("\n")
		b.WriteString(descriptionStyle.Render(
			"Please wait. Do not close the application.",
		))

		content := b.String()

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		))

		view.AltScreen = true
		return view
	}

	if m.screen == "uninstalling" {
		var b strings.Builder

		b.WriteString(titleStyle.Render("Uninstalling YACS"))
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" Removing your deployment...")
		b.WriteString("\n\n")

		if m.uninstallRemoveData {
			b.WriteString("Removing the deployment and its persistent data.\n")
			b.WriteString("This includes YACS configuration and certificates.\n")
		} else {
			b.WriteString("Removing the running Docker deployment.\n")
			b.WriteString("Your configuration and persistent data will be preserved.\n")
		}

		b.WriteString("\n")
		b.WriteString(descriptionStyle.Render(
			"Please wait. Do not close the application.",
		))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "restore-result" {
		var b strings.Builder

		if m.restoreErr != nil {
			b.WriteString(errorStyle.Render("Restoration failed"))
			b.WriteString("\n\n")
			b.WriteString(m.restoreErr.Error())
		} else {
			b.WriteString(titleStyle.Render("Restoration completed!"))
			b.WriteString("\n\n")
			b.WriteString("Your existing YACS deployment is running again.")
			b.WriteString("\n\n")

			b.WriteString(fmt.Sprintf("Deployment: %s\n", m.restoreInfo.DeploymentName))
			b.WriteString(fmt.Sprintf("Server URL: %s\n", m.restoreInfo.ServerURL))
			b.WriteString("Admin username: admin\n")
			b.WriteString("Admin password: unchanged from the original installation\n")

		}

		b.WriteString("\n\n")
		b.WriteString(descriptionStyle.Render("Press Enter or Esc to return to the menu."))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "install-result" {
		var b strings.Builder

		if m.installErr != nil {
			b.WriteString(errorStyle.Render("Installation failed"))
			b.WriteString("\n\n")
			b.WriteString(m.installErr.Error())
		} else {
			b.WriteString(titleStyle.Render("Installation completed!"))
			b.WriteString("\n\n")

			serverURL := fmt.Sprintf(
				"https://%s:9200",
				m.pendingInstallation.Network.PublicHost,
			)

			details := []string{
				"Your server is ready!",
				"",
				// fmt.Sprintf("Server name: %s", m.pendingInstallation.Network.PublicHost),
				fmt.Sprintf("Server URL: %s", serverURL),
				"Admin username: admin",
				"Admin password: The password you chose during setup",
			}

			for _, line := range details {
				b.WriteString(
					lipgloss.NewStyle().
						Width(70).
						Align(lipgloss.Center).
						Render(line),
				)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n\n")
		b.WriteString(footerStyle.Render("Enter or Esc to return to the main menu"))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "uninstall-result" {
		var b strings.Builder

		if m.uninstallErr != nil {
			b.WriteString(errorStyle.Render("Uninstallation failed"))
			b.WriteString("\n\n")
			b.WriteString(m.uninstallErr.Error())
		} else {
			b.WriteString(titleStyle.Render("Uninstallation completed!"))
			b.WriteString("\n\n")

			if m.uninstallRemoveData {
				b.WriteString("The deployment and its persistent data have been removed.\n")
				b.WriteString("YACS configuration files and certificates have also been removed.\n")
			} else {
				b.WriteString("The Docker deployment has been removed.\n")
				b.WriteString("Your configuration and persistent data have been preserved.\n")
				b.WriteString("You can restore the deployment later using YACS.\n")
			}

			if m.uninstallOutput != "" {
				b.WriteString("\n")
				b.WriteString(descriptionStyle.Render(strings.TrimSpace(m.uninstallOutput)))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n\n")
		b.WriteString(footerStyle.Render("Enter or Esc to return to the main menu"))

		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			b.String(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "install" {
		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.installer.View(),
		))
		view.AltScreen = true
		return view
	}

	if m.screen == "uninstall" {
		view := tea.NewView(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.uninstaller.View(),
		))
		view.AltScreen = true
		return view
	}
	var b strings.Builder

	b.WriteString(titleStyle.Render("YACS"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Yet Another Cloud Server"))
	b.WriteString("\n\n")

	b.WriteString("Your personal cloud, made simple.\n")
	b.WriteString(subtitleStyle.Render("Install and manage your ownCloud Infinite Scale server."))
	b.WriteString("\n\n")

	b.WriteString("MAIN MENU\n\n")

	contentWidth := 56
	if m.width > 0 && m.width < contentWidth+8 {
		contentWidth = m.width - 8
	}
	if contentWidth < 20 {
		contentWidth = 20
	}

	for i, item := range menuItems {
		prefix := "  "
		style := menuStyle

		if m.cursor == i {
			prefix = "› "
			style = selectedMenuStyle
		}

		content := fmt.Sprintf(
			"%s%s\n%s",
			prefix,
			item.title,
			descriptionStyle.Render(item.description),
		)

		b.WriteString(style.Width(contentWidth).Render(content))
		b.WriteString("\n\n")
	}

	if m.status != "" {
		b.WriteString(statusStyle.Render(m.status))
		b.WriteString("\n\n")
	}

	b.WriteString(footerStyle.Render("↑/↓ Navigate  •  Enter Select  •  q Quit"))

	view := tea.NewView(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		b.String(),
	))
	view.AltScreen = true

	return view
}
