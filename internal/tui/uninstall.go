package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type uninstallStep int

const (
	uninstallMode uninstallStep = iota
	uninstallReview
	uninstallDeleteConfirmation
)

type uninstallModel struct {
	step uninstallStep

	modeCursor   int
	reviewCursor int

	removeData bool
	confirmed  bool

	deleteInput textinput.Model

	err error
}

func newUninstallModel() uninstallModel {
	input := textinput.New()
	input.Prompt = "Confirmation: "
	input.CharLimit = 6

	return uninstallModel{
		step:        uninstallMode,
		deleteInput: input,
	}
}

func (m uninstallModel) Update(msg tea.Msg) (uninstallModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.step == uninstallMode && m.modeCursor > 0 {
				m.modeCursor--
			}

			if m.step == uninstallReview && m.reviewCursor > 0 {
				m.reviewCursor--
			}

		case "down", "j":
			if m.step == uninstallMode && m.modeCursor < 1 {
				m.modeCursor++
			}

			if m.step == uninstallReview && m.reviewCursor < 1 {
				m.reviewCursor++
			}

		case "enter":
			switch m.step {
			case uninstallMode:
				m.removeData = m.modeCursor == 1
				m.step = uninstallReview
				m.reviewCursor = 0
				m.err = nil
				return m, nil

			case uninstallReview:
				if m.reviewCursor == 1 {
					m.step = uninstallMode
					return m, nil
				}

				if m.removeData {
					m.step = uninstallDeleteConfirmation
					m.deleteInput.SetValue("")
					return m, m.deleteInput.Focus()
				}

				m.confirmed = true
				return m, nil

			case uninstallDeleteConfirmation:
				if m.deleteInput.Value() != "DELETE" {
					m.err = fmt.Errorf("please type DELETE exactly to confirm")
					return m, nil
				}

				m.err = nil
				m.deleteInput.Blur()
				m.confirmed = true
				return m, nil
			}

		case "esc":
			switch m.step {
			case uninstallMode:
				return m, nil

			case uninstallReview:
				m.step = uninstallMode
				m.err = nil
				return m, nil

			case uninstallDeleteConfirmation:
				m.step = uninstallReview
				m.deleteInput.Blur()
				m.deleteInput.SetValue("")
				m.err = nil
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		// The parent TUI handles terminal dimensions.
	}

	if m.step == uninstallDeleteConfirmation {
		var cmd tea.Cmd
		m.deleteInput, cmd = m.deleteInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m uninstallModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Uninstall YACS"))
	b.WriteString("\n\n")

	switch m.step {
	case uninstallMode:
		b.WriteString("Choose what you want to remove.\n\n")

		options := []string{
			"Remove deployment only",
			"Remove everything",
		}

		for i, option := range options {
			cursor := "  "
			style := menuStyle

			if m.modeCursor == i {
				cursor = "› "
				style = selectedMenuStyle
			}

			b.WriteString(style.Render(cursor + option))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Use ↑/↓ to choose an option and Enter to continue.",
		))

	case uninstallReview:
		if m.removeData {
			b.WriteString("Review full removal.\n\n")

			b.WriteString("The following will be removed:\n")
			b.WriteString("• Running YACS containers and Compose deployment\n")
			b.WriteString("• Named Docker volumes and persistent application data\n")
			b.WriteString("• YACS configuration and environment files\n")
			b.WriteString("• YACS-managed certificates and deployment files\n\n")

			b.WriteString(descriptionStyle.Render(
				"Docker images will be kept. This operation permanently deletes your cloud data.",
			))
		} else {
			b.WriteString("Review deployment removal.\n\n")

			b.WriteString("The following will be removed:\n")
			b.WriteString("• Running YACS containers and Compose deployment\n\n")

			b.WriteString("The following will be preserved:\n")
			b.WriteString("• Docker volumes and persistent application data\n")
			b.WriteString("• YACS configuration and environment files\n")
			b.WriteString("• TLS certificates and deployment files\n\n")

			b.WriteString(descriptionStyle.Render(
				"You can restore the deployment later using the existing configuration.",
			))
		}

		b.WriteString("\n\n")

		options := []string{
			"Continue",
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

	case uninstallDeleteConfirmation:
		b.WriteString(errorStyle.Render("Permanent data deletion"))
		b.WriteString("\n\n")

		b.WriteString(
			"This will permanently delete your YACS configuration, certificates, and Docker volumes.\n\n",
		)

		b.WriteString("Type DELETE below to confirm:\n\n")
		b.WriteString(m.deleteInput.View())
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(descriptionStyle.Render(
			"Press Enter to confirm or Esc to go back.",
		))
	}

	b.WriteString("\n\n")
	b.WriteString(footerStyle.Render(
		"↑/↓ Navigate  •  Enter Select  •  Esc Back",
	))

	return b.String()
}
