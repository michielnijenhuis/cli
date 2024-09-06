package cli

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/michielnijenhuis/cli/helper"
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
func Pad(text string, length int, char string) string {
	stripped := StripEscapeSequences(text)
	count := max(0, length-helper.Width(stripped))
	rightPadding := strings.Repeat(char, count)
	return text + rightPadding
}

func PadStart(text string, length int, char string) string {
	stripped := StripEscapeSequences(text)
	count := max(0, length-helper.Width(stripped))
	return strings.Repeat(char, count) + text
}

func PadEnd(text string, length int, char string) string {
	stripped := StripEscapeSequences(text)
	count := max(0, length-helper.Width(stripped))
	return text + strings.Repeat(char, count)
}

func PadCenter(text string, length int, char string) string {
	stripped := StripEscapeSequences(text)
	inputLen := utf8.RuneCountInString(stripped)
	if inputLen >= length {
		return stripped
	}

	padLen := utf8.RuneCountInString(char)
	if padLen == 0 {
		char = " "
		padLen = 1
	}

	totalPad := length - inputLen
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	leftPadStr := strings.Repeat(char, (leftPad+padLen-1)/padLen)[:leftPad]
	rightPadStr := strings.Repeat(char, (rightPad+padLen-1)/padLen)[:rightPad]

	return leftPadStr + text + rightPadStr
}

func TruncateStart(text string, width int) string {
	if width <= 0 || helper.Width(text) <= width {
		return text
	}

	return "…" + text[width-1:]
}

func StripEscapeSequences(text string) string {
	re1 := regexp.MustCompile(`\x1B[^m]*m`)
	tags := GetStyleTags()
	re2 := regexp.MustCompile(fmt.Sprintf("<(%s)>(.*?)<\\/([a-z]+)>", strings.Join(tags, "|")))
	re3 := regexp.MustCompile(`<(?:fg|bg|options)=[^;>]+(?:;(?:fg|bg|options)=[^;>]+)*>([^<]+)</>`)

	text = re1.ReplaceAllString(text, "")
	text = re2.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re2.FindStringSubmatch(match)
		if len(submatches) > 0 {
			if submatches[1] == submatches[3] {
				return submatches[2]
			}
		}
		return match
	})
	text = re3.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re3.FindStringSubmatch(match)
		if len(submatches) > 1 {
			return submatches[1]
		}
		return match
	})

	return text
}

func Longest(lines []string, minWidth int, padding int) int {
	longest := minWidth

	for _, line := range lines {
		w := helper.Width(StripEscapeSequences(line)) + padding
		longest = max(longest, w)
	}

	return longest
}

func MbSplit(s string, length int) []string {
	if length < 1 {
		length = 1
	}
	var result []string
	runeCount := 0
	start := 0

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size == 1 {
			// Handle invalid UTF-8 encoding
			return nil
		}

		runeCount++
		if runeCount == length {
			result = append(result, s[:start+size])
			s = s[start+size:]
			start = 0
			runeCount = 0
		} else {
			start += size
		}
	}

	if start > 0 {
		result = append(result, s)
	}

	return result
}

func MbSubstr(s string, start int, length int) string {
	runes := []rune(s)
	runeCount := len(runes)

	if start < 0 {
		start = runeCount + start
		if start < 0 {
			start = 0
		}
	}

	if start > runeCount {
		return ""
	}

	if length < 0 {
		length = runeCount - start + length
	}

	end := start + length
	if end > runeCount {
		end = runeCount
	}

	return string(runes[start:end])
}
