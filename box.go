package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

const defaultBoxColor = "gray"

func Box(title string, body string, footer string, color string, info string) string {
	if color == "" {
		color = defaultBoxColor
	}

	var output strings.Builder

	minWidth := 60
	terminalWidth, _ := TerminalWidth()
	minWidth = min(minWidth, terminalWidth-6)

	bodyLines := strings.Split(body, Eol)
	var footerLines []string
	if footer != "" {
		footerLines = strings.Split(footer, Eol)
	}

	lines := make([]string, 0, len(bodyLines)+len(footerLines)+1)
	lines = append(lines, title)
	lines = append(lines, bodyLines...)
	lines = append(lines, footerLines...)
	width := Longest(lines, minWidth, 0)

	titleLength := helper.Width(StripEscapeSequences(title))
	titleLabel := ""
	if titleLength > 0 {
		titleLabel = " " + title + " "
	}
	topBorderLength := width - titleLength
	if titleLength == 0 {
		topBorderLength += 2
	}
	topBorder := strings.Repeat("─", topBorderLength)

	output.WriteString(fmt.Sprintf("<fg=%s>┌</>%s<fg=%s>%s┐</>", color, titleLabel, color, topBorder) + Eol)

	for _, line := range bodyLines {
		output.WriteString(fmt.Sprintf("<fg=%s>│</> %s <fg=%s>│</>", color, Pad(line, width, " "), color) + Eol)
	}

	if len(footerLines) > 0 {
		output.WriteString(fmt.Sprintf("<fg=%s>├%s┤</>", color, strings.Repeat("─", width+2)) + Eol)

		for _, line := range footerLines {
			output.WriteString(fmt.Sprintf("<fg=%s>│</> %s <fg=%s>│</>", color, Pad(line, width, " "), color) + Eol)
		}
	}

	bottomBorderLength := width
	if info != "" {
		bottomBorderLength -= helper.Width(StripEscapeSequences(info))
	} else {
		bottomBorderLength += 2
	}

	bottomBorder := strings.Repeat("─", bottomBorderLength)
	if info != "" {
		info = " " + strings.TrimSpace(info) + " "
	}

	output.WriteString(fmt.Sprintf("<fg=%s>└%s%s┘</>", color, bottomBorder, info) + Eol)

	return strings.TrimSpace(output.String()) + Eol
}
