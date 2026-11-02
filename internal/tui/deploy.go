package tui

import (
	"bytes"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/caviarcbr/YACS/internal/cli"
	"github.com/caviarcbr/YACS/internal/setup"
)

type restoreDockerAccessMsg struct {
	needsSudo bool
	err       error
}

type restoreSudoAuthMsg struct {
	err error
}

type dockerAccessMsg struct {
	needsSudo bool
	err       error
}

type sudoAuthMsg struct {
	err error
}

type restoreProgressMsg struct {
	stage  string
	events <-chan installEvent
}

type restoreFinishedMsg struct {
	output string
	err    error
}

type installProgressMsg struct {
	stage  string
	events <-chan installEvent
}

type installFinishedMsg struct {
	output string
	err    error
}

type installEvent struct {
	stage  string
	output string
	err    error
	done   bool
}

type restoreInspectMsg struct {
	info cli.RestoreInfo
	err  error
}

type uninstallDockerAccessMsg struct {
	needsSudo bool
	err       error
}

type uninstallSudoAuthMsg struct {
	err error
}

type uninstallFinishedMsg struct {
	output string
	err    error
}

func checkDockerAccessCmd() tea.Cmd {
	return func() tea.Msg {
		needsSudo, err := cli.CheckDockerAccessForTUI()

		return dockerAccessMsg{
			needsSudo: needsSudo,
			err:       err,
		}
	}
}

func inspectRestoreCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := cli.InspectRestore("")

		return restoreInspectMsg{
			info: info,
			err:  err,
		}
	}
}

func authenticateSudoCmd(password string) tea.Cmd {
	return func() tea.Msg {
		err := cli.AuthenticateDockerForTUI(password)

		return sudoAuthMsg{err: err}
	}
}

func installConfiguredCmd(installation setup.Installation) tea.Cmd {
	events := make(chan installEvent, 16)

	return func() tea.Msg {
		go func() {
			var output bytes.Buffer

			err := cli.InstallConfiguredWithProgress(
				"",
				installation,
				&output,
				func(stage string) {
					events <- installEvent{stage: stage}
				},
			)

			events <- installEvent{
				done:   true,
				output: output.String(),
				err:    err,
			}

			close(events)
		}()

		return receiveInstallEvent(events)
	}
}

func startRestoreCmd() tea.Cmd {
	events := make(chan installEvent, 16)

	return func() tea.Msg {
		go func() {
			var output bytes.Buffer

			err := cli.RestoreConfiguredWithProgress(
				"",
				&output,
				func(stage string) {
					events <- installEvent{stage: stage}
				},
			)

			events <- installEvent{
				done:   true,
				output: output.String(),
				err:    err,
			}

			close(events)
		}()

		return receiveRestoreEvent(events)
	}
}

func waitForInstallEvent(events <-chan installEvent) tea.Cmd {
	return func() tea.Msg {
		return receiveInstallEvent(events)
	}
}

func receiveInstallEvent(events <-chan installEvent) tea.Msg {
	event, ok := <-events
	if !ok {
		return installFinishedMsg{
			err: errors.New("installation progress channel closed unexpectedly"),
		}
	}

	if event.done {
		return installFinishedMsg{
			output: event.output,
			err:    event.err,
		}
	}

	return installProgressMsg{
		stage:  event.stage,
		events: events,
	}
}

func waitForRestoreEvent(events <-chan installEvent) tea.Cmd {
	return func() tea.Msg {
		return receiveRestoreEvent(events)
	}
}

func receiveRestoreEvent(events <-chan installEvent) tea.Msg {
	event, ok := <-events
	if !ok {
		return restoreFinishedMsg{
			err: errors.New("restore progress channel closed unexpectedly"),
		}
	}

	if event.done {
		return restoreFinishedMsg{
			output: event.output,
			err:    event.err,
		}
	}

	return restoreProgressMsg{
		stage:  event.stage,
		events: events,
	}
}

func checkRestoreDockerAccessCmd() tea.Cmd {
	return func() tea.Msg {
		needsSudo, err := cli.CheckDockerAccessForRestoreTUI()

		return restoreDockerAccessMsg{
			needsSudo: needsSudo,
			err:       err,
		}
	}
}

func checkUninstallDockerAccessCmd() tea.Cmd {
	return func() tea.Msg {
		needsSudo, err := cli.CheckDockerAccessForUninstallTUI()

		return uninstallDockerAccessMsg{
			needsSudo: needsSudo,
			err:       err,
		}
	}
}

func authenticateRestoreSudoCmd(password string) tea.Cmd {
	return func() tea.Msg {
		err := cli.AuthenticateDockerForTUI(password)

		return restoreSudoAuthMsg{err: err}
	}
}

func authenticateUninstallSudoCmd(password string) tea.Cmd {
	return func() tea.Msg {
		err := cli.AuthenticateDockerForTUI(password)
		return uninstallSudoAuthMsg{err: err}
	}
}

func uninstallConfiguredCmd(removeData bool) tea.Cmd {
	return func() tea.Msg {
		var output bytes.Buffer

		err := cli.UninstallConfiguredForTUI("", removeData, &output)

		return uninstallFinishedMsg{
			output: output.String(),
			err:    err,
		}
	}
}
