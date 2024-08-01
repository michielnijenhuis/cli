package helper

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const tagOpenRegexSegment = `[a-z](?:[^\\<>]*|\\.)*`
const tagCloseRegexSegment = `[a-z][^<>]*`
const urlPattern = `https?:\/\/\S+`

func Wrap(text string, width int, breakChar string, allowCutUrls bool) string {

	tagPattern := regexp.MustCompile(fmt.Sprintf(`<(%s|/%s?)>`, tagOpenRegexSegment, tagCloseRegexSegment))
	urlRegex := regexp.MustCompile(urlPattern)
	breakPointPattern := regexp.MustCompile(fmt.Sprintf(`.{1,%d}(\\s|$)|.{1,%d}`, width, width))

	var parts []string
	lastIndex := 0

	matches := tagPattern.FindAllStringIndex(text, -1)
	if allowCutUrls {
		matches = append(matches, urlRegex.FindAllStringIndex(text, -1)...)
	}

	// Sort matches to handle interleaved tag and URL matches
	sort.Slice(matches, func(i, j int) bool {
		return matches[i][0] < matches[j][0]
	})

	for _, match := range matches {
		if match[0] > lastIndex {
			wrapLine(text[lastIndex:match[0]], breakPointPattern, &parts, breakChar)
		}

		parts = append(parts, text[match[0]:match[1]])
		lastIndex = match[1]
	}

	if lastIndex < len(text) {
		wrapLine(text[lastIndex:], breakPointPattern, &parts, breakChar)
	}

	return strings.Join(parts, "")
}

func wrapLine(line string, breakPointPattern *regexp.Regexp, parts *[]string, breakChar string) {
	matches := breakPointPattern.FindAllStringSubmatchIndex(line, -1)
	lastMatchEnd := 0

	for _, match := range matches {
		chunk := line[match[0]:match[1]]
		*parts = append(*parts, strings.TrimRight(chunk, " "))
		lastMatchEnd = match[1]
		if lastMatchEnd < len(line) {
			*parts = append(*parts, breakChar)
		}
	}

	if lastMatchEnd < len(line) {
		*parts = append(*parts, line[lastMatchEnd:])
	}
}
