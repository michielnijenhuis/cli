package helper

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	tagPattern = regexp.MustCompile(`</?[^>]+>`)
	urlRegex   = regexp.MustCompile(`https?://\S+`)
)

type itemType int

const (
	itemText itemType = iota
	itemTag
	itemURL
)

type item struct {
	typ   itemType
	value string
}

// Wrap wraps text to a given width and returns a single string joined by breakChar.
func Wrap(text string, width int, breakChar string, allowCutUrls bool) string {
	return strings.Join(WrapToSlice(text, width, breakChar, allowCutUrls), breakChar)
}

// WrapToSlice breaks text into wrapped line segments while preserving and carrying over tags.
func WrapToSlice(text string, width int, breakChar string, allowCutUrls bool) []string {
	if width <= 0 {
		return []string{text}
	}

	var parts []string
	rawLines := strings.Split(text, "\n")

	for _, rawLine := range rawLines {
		wrapSingleLine(rawLine, width, &parts, allowCutUrls)
	}

	return parts
}

func wrapSingleLine(line string, width int, parts *[]string, allowCutUrls bool) {
	if line == "" {
		*parts = append(*parts, "")
		return
	}

	items := tokenize(line, allowCutUrls)
	var activeTags []string

	var currentLine []item
	currentWidth := 0

	flushLine := func() {
		// Build line string and append closing tags if needed
		var sb strings.Builder
		for _, tag := range activeTags {
			sb.WriteString(tag)
		}
		for _, it := range currentLine {
			sb.WriteString(it.value)
		}
		*parts = append(*parts, sb.String())
		currentLine = nil
		currentWidth = 0
	}

	// Process items into words/tokens
	for _, it := range items {
		switch it.typ {
		case itemTag:
			updateTagStack(&activeTags, it.value)
			currentLine = append(currentLine, it)
		case itemURL:
			urlWidth := utf8.RuneCountInString(it.value)
			if !allowCutUrls && currentWidth > 0 && currentWidth+urlWidth > width {
				flushLine()
			}
			currentLine = append(currentLine, it)
			currentWidth += urlWidth
		case itemText:
			runes := []rune(it.value)
			i := 0
			for i < len(runes) {
				// Find next word or whitespace sequence
				start := i
				if unicode.IsSpace(runes[i]) {
					for i < len(runes) && unicode.IsSpace(runes[i]) {
						i++
					}
				} else {
					for i < len(runes) && !unicode.IsSpace(runes[i]) {
						i++
					}
				}
				chunk := string(runes[start:i])
				chunkWidth := utf8.RuneCountInString(chunk)

				if unicode.IsSpace(runes[start]) {
					// Whitespace handling
					if currentWidth+chunkWidth > width {
						flushLine()
						// Skip leading spaces on wrapped lines
						continue
					}
					currentLine = append(currentLine, item{typ: itemText, value: chunk})
					currentWidth += chunkWidth
				} else {
					// Word handling
					if currentWidth > 0 && currentWidth+chunkWidth > width {
						flushLine()
					}
					// Handle overly long words that exceed total width
					for chunkWidth > width {
						if currentWidth > 0 {
							flushLine()
						}
						// Split long word
						splitIdx := width - currentWidth
						if splitIdx <= 0 {
							splitIdx = width
						}
						subRunes := []rune(chunk)
						if splitIdx > len(subRunes) {
							splitIdx = len(subRunes)
						}
						subChunk := string(subRunes[:splitIdx])
						currentLine = append(currentLine, item{typ: itemText, value: subChunk})
						flushLine()
						chunk = string(subRunes[splitIdx:])
						chunkWidth = utf8.RuneCountInString(chunk)
					}
					if chunkWidth > 0 {
						currentLine = append(currentLine, item{typ: itemText, value: chunk})
						currentWidth += chunkWidth
					}
				}
			}
		}
	}

	if len(currentLine) > 0 {
		var sb strings.Builder
		for _, tag := range activeTags {
			sb.WriteString(tag)
		}
		for _, it := range currentLine {
			sb.WriteString(it.value)
		}
		*parts = append(*parts, sb.String())
	}
}

func tokenize(line string, allowCutUrls bool) []item {
	matches := tagPattern.FindAllStringIndex(line, -1)
	if !allowCutUrls {
		matches = append(matches, urlRegex.FindAllStringIndex(line, -1)...)
	}

	// Sort matches by starting index
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i][0] > matches[j][0] {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	var items []item
	lastIndex := 0

	for _, match := range matches {
		if match[0] < lastIndex {
			continue
		}
		if match[0] > lastIndex {
			items = append(items, item{typ: itemText, value: line[lastIndex:match[0]]})
		}

		val := line[match[0]:match[1]]
		if strings.HasPrefix(val, "</") {
			items = append(items, item{typ: itemTag, value: val})
		} else if tagPattern.MatchString(val) {
			items = append(items, item{typ: itemTag, value: val})
		} else {
			items = append(items, item{typ: itemURL, value: val})
		}
		lastIndex = match[1]
	}

	if lastIndex < len(line) {
		items = append(items, item{typ: itemText, value: line[lastIndex:]})
	}

	return items
}

func updateTagStack(stack *[]string, tag string) {
	if tag == "</>" || strings.HasPrefix(tag, "</") {
		if len(*stack) > 0 {
			*stack = (*stack)[:len(*stack)-1]
		}
	} else {
		*stack = append(*stack, tag)
	}
}
