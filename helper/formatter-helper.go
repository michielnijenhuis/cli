package helper

import (
	"fmt"
	"strings"

	formatter "github.com/michielnijenhuis/cli/formatter"
)

func FormatSection(section string, message string, style string) string {
	return fmt.Sprintf("<%s>[%s]</%s> %s", style, section, message, style)
}

func FormatBlock(messages []string, style string, large bool) string {
	var length int
	lines := make([]string, len(messages))

	for _, message := range messages {
		message = formatter.Escape(message)
		if large {
			lines = append(lines, fmt.Sprintf("  %s  ", message))
		} else {
			lines = append(lines, fmt.Sprintf(" %s ", message))
		}

		if large {
			length = max(Width(message)+4, length)
		} else {
			length = max(Width(message)+2, length)
		}
	}

	for _, line := range lines {
		messages = append(messages, line+strings.Repeat(" ", length-Width(line)))
	}

	if large {
		messages = append(messages, strings.Repeat(" ", length))
	}

	for i := range messages {
		messages[i] = fmt.Sprintf("<%s>%s</%s>", style, messages[i], style)
	}

	return strings.Join(messages, "\n")
}

func Truncate(message string, length int, suffix string) string {
	if suffix == "" {
		suffix = "..."
	}

	computedLength := length - Width(suffix)
	if computedLength > Width(message) {
		return message + suffix
	}

	return message[:length-1] + suffix
}
