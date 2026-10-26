package terminal

import (
	"os"

	"golang.org/x/term"
)

func Columns() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		return 80
	}

	return width
}

func Lines() int {
	_, height, err := term.GetSize(0)
	if err != nil {
		return 80
	}

	return height
}

func Size() (int, int) {
	w, h, err := term.GetSize(0)
	if err != nil {
		return 80, 80
	}

	return w, h
}

func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
