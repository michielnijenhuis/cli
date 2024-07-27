package formatter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

func FormatSection(section string, message string, style string) string {
	return fmt.Sprintf("<%s>[%s]</%s> %s", style, section, message, style)
}

func FormatBlock(messages []string, style string, large bool) string {
	var length int
	lines := make([]string, len(messages))

	for _, message := range messages {
		message = Escape(message)
		if large {
			lines = append(lines, fmt.Sprintf("  %s  ", message))
		} else {
			lines = append(lines, fmt.Sprintf(" %s ", message))
		}

		if large {
			length = max(helper.Width(message)+4, length)
		} else {
			length = max(helper.Width(message)+2, length)
		}
	}

	for _, line := range lines {
		messages = append(messages, line+strings.Repeat(" ", length-helper.Width(line)))
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

	computedLength := length - helper.Width(suffix)
	if computedLength > helper.Width(message) {
		return message + suffix
	}

	return message[:length-1] + suffix
}

// TODO: implement
func DetectEncoding(formatter OutputFormatterInferface, str string) string {
	return ""
}

func RemoveDecoration(formatter OutputFormatterInferface, str string) string {
	isDecorated := formatter.IsDecorated()
	formatter.SetDecorated(false)

	str = formatter.Format(str)

	re1 := regexp.MustCompile(`\033\[[^m]*m`)
	str = re1.ReplaceAllString(str, "")

	re2 := regexp.MustCompile(`\\033]8;[^;]*;[^\\033]*\\033\\\\`)
	str = re2.ReplaceAllString(str, "")

	formatter.SetDecorated(isDecorated)

	return str
}
