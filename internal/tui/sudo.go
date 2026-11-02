package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type sudoModel struct {
	input textinput.Model
	err   error
}

func newSudoModel() sudoModel {
	input := textinput.New()
	input.Prompt = "Sudo password: "
	input.EchoMode = textinput.EchoPassword
	input.CharLimit = 256
	input.Focus()

	return sudoModel{
		input: input,
	}
}

func (m sudoModel) Update(msg tea.Msg) (sudoModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m sudoModel) View() string {
	content := titleStyle.Render("Docker access required") +
		"\n\n" +
		"YACS needs administrator permission to access Docker\n" +
		"and install your personal cloud services.\n\n" +
		"Enter your sudo password to continue.\n\n" +
		m.input.View()

	if m.err != nil {
		content += "\n\n" + errorStyle.Render(m.err.Error())
	}

	content += "\n\n" +
		descriptionStyle.Render("Enter to continue  •  Esc to cancel") +
		"\n" +
		descriptionStyle.Render("Your password is not saved by YACS.")

	return content
}

func (m *sudoModel) ClearPassword() {
	m.input.SetValue("")
}

func (m *sudoModel) SetError(err error) {
	m.err = err
}

func (m *sudoModel) Password() string {
	return m.input.Value()
}

func (m *sudoModel) Focus() tea.Cmd {
	return m.input.Focus()
}

func (m *sudoModel) Blur() {
	m.input.Blur()
}

func (m *sudoModel) Empty() bool {
	return m.input.Value() == ""
}

func (m *sudoModel) String() string {
	return fmt.Sprint(m.input.Value())
}
