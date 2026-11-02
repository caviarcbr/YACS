package cli

import (
	"fmt"
	"os"
)

func Run(args []string) error {
	configPath, commandArgs, err := parseGlobalOptions(args)
	if err != nil {
		return err
	}

	if len(commandArgs) == 0 {
		printHelp()
		return nil
	}

	switch commandArgs[0] {
	case "install":
		if len(commandArgs) != 1 {
			return fmt.Errorf("usage: yacs install")
		}
		return runInstall(configPath, os.Stdin, os.Stdout)

	case "uninstall":
		removeData := false

		if len(commandArgs) > 2 {
			return fmt.Errorf("usage: yacs uninstall [--remove-data]")
		}

		if len(commandArgs) == 2 {
			if commandArgs[1] != "--remove-data" {
				return fmt.Errorf("unknown uninstall option: %s", commandArgs[1])
			}
			removeData = true
		}

		return runUninstall(configPath, removeData, os.Stdin, os.Stdout)

	case "version":
		fmt.Printf("%s version %s\n", AppName, AppVersion)

	case "help", "--help", "-h":
		printHelp()

	default:
		return fmt.Errorf("unknown command: %s", commandArgs[0])
	}

	return nil
}

func parseGlobalOptions(args []string) (string, []string, error) {
	var configPath string

	for len(args) > 0 {
		switch {
		case args[0] == "--config":
			if len(args) < 2 || args[1] == "" {
				return "", nil, fmt.Errorf("--config requires a file path")
			}

			configPath = args[1]
			args = args[2:]

		case len(args[0]) > len("--config=") &&
			args[0][:len("--config=")] == "--config=":
			configPath = args[0][len("--config="):]
			args = args[1:]

		default:
			return configPath, args, nil
		}
	}

	return configPath, args, nil
}

func printHelp() {
	fmt.Printf(`%s - Yet Another Cloud Server

Usage:
  yacs [--config <path>] <command>

Global Options:
  --config <path>    Specify a configuration file

Commands:
  install            Install ownCloud Infinite Scale
  uninstall          Remove the YACS deployment
  version            Display version information
  help               Display this help message

`, AppName)
}
