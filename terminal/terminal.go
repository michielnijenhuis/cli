package terminal

import (
	"os"

	"golang.org/x/term"
)

func GetWidth() (int, error) {
	var width int
	var err error
	width, _, err = term.GetSize(0)

	if err != nil {
		return 80, err
	}

	return width, nil
}

func GetHeight() (int, error) {
	var height int
	var err error
	_, height, err = term.GetSize(0)

	if err != nil {
		return 80, err
	}

	return height, nil
}

func GetSize() (int, int, error) {
	return term.GetSize(0)
}

func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
