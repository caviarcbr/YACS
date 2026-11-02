package main

import (
	"fmt"
	"os"

	"github.com/caviarcbr/YACS/internal/cli"
	"github.com/caviarcbr/YACS/internal/tui"
)

func main() {
	if len(os.Args) == 1 {
		if err := tui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
