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

var originalState *term.State
var terminalState int

func TerminalMakeRaw() error {
	if terminalState > 0 {
		terminalState++
		return nil
	}

	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	terminalState++
	originalState = state

	return nil
}

func TerminalRestore() error {
	if terminalState <= 0 || terminalState > 1 {
		if terminalState > 1 {
			terminalState--
		}

		return nil
	}

	terminalState = 0
	err := term.Restore(int(os.Stdin.Fd()), originalState)
	originalState = nil

	return err
}
