package helper

import (
	"regexp"
	"strings"
)

func StrimWidth(str string, start int, width int, trimMarker string) string {
	var result string
	var currentWidth int
	var i int

	for i < len(str) && currentWidth < start+width {
		char := str[i]
		charCode := []rune(string(char))[0]
		charWidth := 1
		if charCode > 127 {
			charWidth = 2
		}

		if currentWidth+charWidth > start+width {
			break
		}

		if currentWidth >= start {
			result += string(char)
		}

		currentWidth += charWidth
		i++
	}

	if currentWidth > start+width {
		result += trimMarker
	}

	return result
}

func TrimWidthBackwards(str string, start int, width int) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	reversed := string(runes)
	trimmed := reversed[start:width]

	runes = []rune(trimmed)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return trimmed
}

// Pads while ignoring escape sequences
func Pad(text string, length int, char byte) string {
	c := max(0, length-Width(StripEscapeSequences(text)))
	rightPadding := strings.Repeat(string(char), c)
	return text + rightPadding
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

func TruncateStart(text string, width int) string {
	if width <= 0 || Width(text) <= width {
		return text
	}

	return "…" + text[width-1:]
}

func StripEscapeSequences(text string) string {
	re1 := regexp.MustCompile(`\x1B[^m]*m`)
	re2 := regexp.MustCompile(`<(info|comment|question|error|header|highlight)>(.*?)<\/\\1>`)
	re3 := regexp.MustCompile(`(i?)<(?:(?:[fb]g|options)=[a-z,;]+)+>(.*?)<\/>`)

	text = re1.ReplaceAllString(text, "")
	text = re2.ReplaceAllString(text, "$2")
	text = re3.ReplaceAllString(text, "$1")

	return text
}

func Longest(lines []string, minWidth int, padding int) int {
	longest := minWidth

	for _, line := range lines {
		w := Width(StripEscapeSequences(line)) + padding
		longest = max(longest, w)
	}

	return longest
}
