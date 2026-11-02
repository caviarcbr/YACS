package setup

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
)

// ServiceSelection identifies which services the installer deploys.
type ServiceSelection int

const (
	OCISOnly ServiceSelection = iota
	OCISWithOnlyOffice
)

// String returns the human-readable name of the selection.
func (s ServiceSelection) String() string {
	if s == OCISWithOnlyOffice {
		return "ownCloud Infinite Scale + ONLYOFFICE"
	}

	return "ownCloud Infinite Scale"
}

// ErrInputClosed is returned when the input ends before an answer is given.
var ErrInputClosed = errors.New("input closed before an answer was given")

// Prompter asks questions on a terminal using plain line-based input.
type Prompter struct {
	source io.Reader
	in     *bufio.Reader
	out    io.Writer
}

// NewPrompter creates a Prompter that reads answers from in and writes
// questions to out.
func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{
		source: in,
		in:     bufio.NewReader(in),
		out:    out,
	}
}

// Installation holds the answers collected by the installation prompts.
type Installation struct {
	Services      ServiceSelection
	Network       NetworkConfig
	AdminPassword string
}

// PromptInstallation asks which services to install, which address devices
// will use to reach the server, and the password for the admin account.
func PromptInstallation(p *Prompter, defaultHost string) (Installation, error) {
	fmt.Fprintln(p.out, "")
	fmt.Fprintln(p.out, "Installation setup")

	index, err := p.Choose("Choose the services to install:", []string{
		"ownCloud Infinite Scale - personal cloud storage and collaboration",
		"ownCloud Infinite Scale + ONLYOFFICE - cloud storage with online document editing",
	})
	if err != nil {
		return Installation{}, err
	}

	fmt.Fprintln(p.out, "")
	fmt.Fprintln(p.out, "Enter the LAN IP address or hostname that devices will use to reach this server.")

	host, err := p.Ask("Server address", defaultHost, func(value string) error {
		_, err := NewNetworkConfig(value)
		return err
	})
	if err != nil {
		return Installation{}, err
	}

	network, err := NewNetworkConfig(host)
	if err != nil {
		return Installation{}, err
	}

	fmt.Fprintln(p.out, "")
	fmt.Fprintln(p.out, "Choose the password for the 'admin' account. You will use it to sign in.")
	fmt.Fprintln(p.out, PasswordRules)

	password, err := p.AskPassword("Admin password", ValidateAdminPassword)
	if err != nil {
		return Installation{}, err
	}

	return Installation{
		Services:      ServiceSelection(index),
		Network:       network,
		AdminPassword: password,
	}, nil
}

// PasswordRules describes the admin password requirements shown to the user.
const PasswordRules = "  12 to 128 characters: letters, digits and symbols, no spaces, and none of $ # ' \" \\ or backtick."

// ValidateAdminPassword enforces the admin password rules. The excluded
// characters would be misread by Compose when the password is written to the
// environment file, so they are rejected up front.
func ValidateAdminPassword(password string) error {
	if len(password) < 12 {
		return errors.New("the password must be at least 12 characters")
	}

	if len(password) > 128 {
		return errors.New("the password must be at most 128 characters")
	}

	for _, r := range password {
		if r < 0x21 || r > 0x7e {
			return errors.New("use letters, digits and symbols only, with no spaces")
		}

		if strings.ContainsRune("$#'\"\\`", r) {
			return fmt.Errorf("the character %q is not allowed", r)
		}
	}

	return nil
}

// Choose shows a numbered list and returns the zero-based index of the
// selected option. Pressing Enter selects the first option.
func (p *Prompter) Choose(title string, options []string) (int, error) {
	fmt.Fprintln(p.out, title)

	for i, option := range options {
		fmt.Fprintf(p.out, "  %d) %s\n", i+1, option)
	}

	for {
		answer, err := p.readLine(fmt.Sprintf("Select [1-%d, default 1]: ", len(options)))
		if err != nil {
			return 0, err
		}

		if answer == "" {
			return 0, nil
		}

		choice, err := strconv.Atoi(answer)
		if err == nil && choice >= 1 && choice <= len(options) {
			return choice - 1, nil
		}

		fmt.Fprintf(p.out, "Enter a number from 1 to %d.\n", len(options))
	}
}

// Ask prompts for a value, using defaultValue when the answer is empty, and
// repeats the question until validate accepts the answer. A nil validate
// accepts any non-empty answer.
func (p *Prompter) Ask(label, defaultValue string, validate func(string) error) (string, error) {
	prompt := label + ": "
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", label, defaultValue)
	}

	for {
		answer, err := p.readLine(prompt)
		if err != nil {
			return "", err
		}

		if answer == "" {
			answer = defaultValue
		}

		if answer == "" {
			fmt.Fprintln(p.out, "A value is required.")
			continue
		}

		if validate != nil {
			if err := validate(answer); err != nil {
				fmt.Fprintf(p.out, "  %v\n", err)
				continue
			}
		}

		return answer, nil
	}
}

// AskPassword prompts for a secret twice without echoing it, and repeats
// until validate accepts it and both entries match.
func (p *Prompter) AskPassword(label string, validate func(string) error) (string, error) {
	for {
		password, err := p.readSecret(label + ": ")
		if err != nil {
			return "", err
		}

		if validate != nil {
			if err := validate(password); err != nil {
				fmt.Fprintf(p.out, "  %v\n", err)
				continue
			}
		}

		repeated, err := p.readSecret("Repeat " + label + ": ")
		if err != nil {
			return "", err
		}

		if repeated != password {
			fmt.Fprintln(p.out, "  The passwords do not match. Try again.")
			continue
		}

		return password, nil
	}
}

// Confirm asks a yes/no question. Anything other than "y" or "yes" is no.
func (p *Prompter) Confirm(question string) (bool, error) {
	answer, err := p.readLine(question + " [y/N]: ")
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(answer)

	return answer == "y" || answer == "yes", nil
}

func (p *Prompter) readLine(prompt string) (string, error) {
	fmt.Fprint(p.out, prompt)

	line, err := p.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading answer: %w", err)
	}

	if errors.Is(err, io.EOF) && line == "" {
		fmt.Fprintln(p.out, "")
		return "", ErrInputClosed
	}

	return strings.TrimSpace(line), nil
}

// readSecret reads one line with terminal echo turned off when the input is
// a terminal. Echo is restored before returning, including on Ctrl+C.
func (p *Prompter) readSecret(prompt string) (string, error) {
	restore := disableEcho(p.source)
	defer restore()

	fmt.Fprint(p.out, prompt)

	line, err := p.in.ReadString('\n')
	fmt.Fprintln(p.out, "")

	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading answer: %w", err)
	}

	if errors.Is(err, io.EOF) && line == "" {
		return "", ErrInputClosed
	}

	return strings.TrimSpace(line), nil
}

// disableEcho turns off terminal echo for in, if in is a terminal, and
// returns a function that turns it back on. On Ctrl+C the echo is restored
// before the process exits, so the shell is not left without echo.
func disableEcho(in io.Reader) func() {
	file, ok := in.(*os.File)
	if !ok {
		return func() {}
	}

	if !runStty(file, "-g") {
		return func() {}
	}

	if !runStty(file, "-echo") {
		return func() {}
	}

	restore := func() {
		runStty(file, "echo")
	}

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)

	done := make(chan struct{})

	go func() {
		select {
		case <-interrupts:
			restore()
			os.Exit(130)
		case <-done:
		}
	}()

	return func() {
		signal.Stop(interrupts)
		close(done)
		restore()
	}
}

// runStty runs stty against the terminal behind file and reports success.
func runStty(file *os.File, args ...string) bool {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = file

	return cmd.Run() == nil
}
