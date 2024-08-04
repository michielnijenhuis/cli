package cli

import (
	"flag"
	"os"

	"golang.org/x/term"
)

func TerminalWidth() (int, error) {
	var width int
	var err error
	width, _, err = term.GetSize(0)

	if err != nil {
		return 80, err
	}

	return width, nil
}

func TerminalHeight() (int, error) {
	var height int
	var err error
	_, height, err = term.GetSize(0)

	if err != nil {
		return 80, err
	}

	return height, nil
}

func TerminalSize() (int, int, error) {
	return term.GetSize(0)
}

func TerminalIsInteractive() bool {
	if flag.Lookup("test.v") != nil {
		// running `go test`
		return false
	}

	return term.IsTerminal(int(os.Stdout.Fd()))
}
