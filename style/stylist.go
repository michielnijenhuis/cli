package style

import "fmt"

func Reset(text string) string {
	return fmt.Sprintf("\x1b[0m%s\x1b[0m", text)
}

func Bold(text string) string {
	return fmt.Sprintf("\x1b[1m%s\x1b[22m", text)
}

func Dim(text string) string {
	return fmt.Sprintf("\x1b[2m%s\x1b[22m", text)
}

func Italic(text string) string {
	return fmt.Sprintf("\x1b[3m%s\x1b[23m", text)
}

func Underline(text string) string {
	return fmt.Sprintf("\x1b[4m%s\x1b[24m", text)
}

func Inverse(text string) string {
	return fmt.Sprintf("\x1b[7m%s\x1b[27m", text)
}

func Hidden(text string) string {
	return fmt.Sprintf("\x1b[8m%s\x1b[28m", text)
}

func Strikethrough(text string) string {
	// return fmt.Sprintf("\x1b[9m%s\x1b[0m", text);
	return fmt.Sprintf("\x1b[9m%s\x1b[29m", text)
}

func Black(text string) string {
	return fmt.Sprintf("\x1b[30m%s\x1b[39m", text)
}

func Red(text string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[39m", text)
}

func Green(text string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[39m", text)
}

func Yellow(text string) string {
	return fmt.Sprintf("\x1b[33m%s\x1b[39m", text)
}

func Blue(text string) string {
	return fmt.Sprintf("\x1b[34m%s\x1b[39m", text)
}

func Magenta(text string) string {
	return fmt.Sprintf("\x1b[35m%s\x1b[39m", text)
}

func Cyan(text string) string {
	return fmt.Sprintf("\x1b[36m%s\x1b[39m", text)
}

func White(text string) string {
	return fmt.Sprintf("\x1b[37m%s\x1b[39m", text)
}

func BgBlack(text string) string {
	return fmt.Sprintf("\x1b[40m%s\x1b[49m", text)
}

func BgRed(text string) string {
	return fmt.Sprintf("\x1b[41m%s\x1b[49m", text)
}

func BgGreen(text string) string {
	return fmt.Sprintf("\x1b[42m%s\x1b[49m", text)
}

func BgYellow(text string) string {
	return fmt.Sprintf("\x1b[43m%s\x1b[49m", text)
}

func BgBlue(text string) string {
	return fmt.Sprintf("\x1b[44m%s\x1b[49m", text)
}

func BgMagenta(text string) string {
	return fmt.Sprintf("\x1b[45m%s\x1b[49m", text)
}

func BgCyan(text string) string {
	return fmt.Sprintf("\x1b[46m%s\x1b[49m", text)
}

func BgWhite(text string) string {
	return fmt.Sprintf("\x1b[47m%s\x1b[49m", text)
}

func Gray(text string) string {
	return fmt.Sprintf("\x1b[90m%s\x1b[39m", text)
}
