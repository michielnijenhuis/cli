package terminal

import (
	"os"

	Terminal "golang.org/x/term"
)

func GetWidth() (int, error) {
	var width int
	var err error
	width, _, err = Terminal.GetSize(0)

	if err != nil {
		return 80, err
	}

	return width, nil
}

func GetHeight() (int, error) {
	var height int
	var err error
	_, height, err = Terminal.GetSize(0)

	if err != nil {
		return 80, err
	}

	return height, nil
}

func GetSize() (int, int, error) {
	return Terminal.GetSize(0)
}

func IsInteractive() bool {
	return Terminal.IsTerminal(int(os.Stdout.Fd()))
}
