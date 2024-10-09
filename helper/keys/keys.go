package keys

const (
	Up         = "\x1b[A"
	ShiftUp    = "\x1b[1;2A"
	Down       = "\x1b[B"
	ShiftDown  = "\x1b[1;2B"
	Right      = "\x1b[C"
	Left       = "\x1b[D"
	UpArrow    = "\x1bOA"
	DownArrow  = "\x1bOB"
	RightArrow = "\x1bOC"
	LeftArrow  = "\x1bOD"
	Escape     = "\x1b"
	Delete     = "\x1b[3~"
	Backspace  = "\x7f"
	Enter      = "\n"
	Space      = " "
	Tab        = "\t"
	ShiftTab   = "\x1b[Z"

	/**
	 * Cancel/SIGINT
	 */
	CtrlC = "\x03"

	/**
	 * Previous/Up
	 */
	CtrlP = "\x10"

	/**
	 * Next/Down
	 */
	CtrlN = "\x0E"

	/**
	 * Forward/Right
	 */
	CtrlF = "\x06"

	/**
	 * Back/Left
	 */
	CtrlB = "\x02"

	/**
	 * Backspace
	 */
	CtrlH = "\x08"

	/**
	 * Home
	 */
	CtrlA = "\x01"

	/**
	 * EOF
	 */
	CtrlD = "\x04"

	/**
	 * End
	 */
	CtrlE = "\x05"

	/**
	 * Negative affirmation
	 */
	CtrlU = "\x15"
)

var Home []string = []string{"\x1b[1~", "\x1bOH", "\x1b[H", "\x1b[7~"}
var End []string = []string{"\x1b[4~", "\x1bOF", "\x1b[F", "\x1b[8~"}

func Is(key string, keys ...string) bool {
	for _, x := range keys {
		if x == key {
			return true
		}
	}

	if len(key) > 1 {
		return Is(string(key[0]), keys...)
	}

	return false
}
