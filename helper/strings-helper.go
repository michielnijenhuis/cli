package helper

import "strings"

// TODO: implement
func StrimWidth(str string, start int, width int, trimMarker string) string {
	return str
}

// TODO: implement
func TrimWidthBackwards(str string, start int, width int) string {
	return str
}

// TODO: implement
func Pad(text string, length int, char string) string {
	return text
}

func PadStart(text string, length int, char byte) string {
	current := len(text)
	if length >= current {
		return text
	}
	return strings.Repeat(string(char), length-current) + text
}

func PadEnd(text string, length int, char byte) string {
	current := len(text)
	if length >= current {
		return text
	}
	return text + strings.Repeat(string(char), length-current)
}

// TODO: implement (remove 2)
func Truncate2(text string, width int) string {
	return text
}

// TODO: implement
func StripEscapeSequences(text string) string {
	return text
}

// TODO: implement
func Longest(lines []string, minWidth int, padding int) int {
	return 0
}
