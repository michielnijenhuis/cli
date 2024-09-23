package cli

import (
	"os"

	"golang.org/x/term"
)

func TerminalWidth() (int, error) {
	width, _, err := term.GetSize(0)
	if err != nil {
		return 80, err
	}

	return width, nil
}

func TerminalHeight() int {
	_, height, err := term.GetSize(0)
	if err != nil {
		return 80
	}

	return height
}

func TerminalSize() (int, int) {
	w, h, err := term.GetSize(0)
	if err != nil {
		return 80, 80
	}

	return w, h
}

func TerminalIsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
