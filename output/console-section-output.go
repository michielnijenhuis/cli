package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/terminal"
)

type ConsoleSectionOutput struct {
	content   []string
	lines     int
	sections  []*ConsoleSectionOutput
	maxHeight int
	StreamOutput
}

func NewConsoleSectionOutput(stream *os.File, sections []*ConsoleSectionOutput, verbosity uint, decorated bool, formatter formatter.OutputFormatterInferface) *ConsoleSectionOutput {
	streamOutput := NewStreamOutput(stream, verbosity, decorated, formatter)

	consoleSectionOutput := &ConsoleSectionOutput{
		StreamOutput: *streamOutput,
		content:      make([]string, 0),
		lines:        0,
		maxHeight:    0,
		sections:     sections,
	}

	helper.Unshift(&consoleSectionOutput.sections, consoleSectionOutput)

	return consoleSectionOutput
}

func (o *ConsoleSectionOutput) SetMaxHeight(height int) {
	previousMaxHeight := o.maxHeight
	o.maxHeight = height

	var existingContent string
	if previousMaxHeight > 0 {
		existingContent = o.popStreamContentUntilCurrentSection(min(previousMaxHeight, o.lines))
	} else {
		existingContent = o.popStreamContentUntilCurrentSection(o.lines)
	}

	o.StreamOutput.DoWrite(o.VisibleContent(), false)
	o.StreamOutput.DoWrite(existingContent, false)
}

func (o *ConsoleSectionOutput) Clear(lines int) {
	if len(o.content) == 0 || !o.IsDecorated() {
		return
	}

	if lines > 0 {
		// splice
	} else {
		lines = o.lines
		o.content = make([]string, 0)
	}

	o.lines -= lines

	if o.maxHeight > 0 {
		o.StreamOutput.DoWrite(o.popStreamContentUntilCurrentSection(min(o.maxHeight, lines)), false)
	} else {
		o.StreamOutput.DoWrite(o.popStreamContentUntilCurrentSection(lines), false)
	}
}

func (o *ConsoleSectionOutput) Overwrite(message string) {
	o.Clear(0)
	o.Writeln(message, 0)
}

func (o *ConsoleSectionOutput) Content() string {
	return strings.Join(o.content, "")
}

func (o *ConsoleSectionOutput) VisibleContent() string {
	if o.maxHeight == 0 {
		return o.Content()
	}

	return strings.Join(o.content[-o.maxHeight:], "")
}

func (o *ConsoleSectionOutput) AddContent(input string, newLine bool) int {
	width, _ := terminal.Width()
	lines := strings.Split(input, "\n")
	var linesAdded int
	count := len(lines) - 1

	for i := 0; i < len(lines); i++ {
		lineContent := lines[i]

		// re-add the line break (that has been removed in the above `explode()` for
		// - every line that is not the last line
		// - if newline is required, also add it to the last line
		if i < count || newLine {
			lineContent += "\n"
		}

		// skip line if there is no text (or newline for that matter)
		if lineContent == "" {
			continue
		}

		// For the first line, check if the previous line (last entry of `this.content`)
		// needs to be continued (i.e. does not end with a line break).
		lastLine := o.content[len(o.content)-1]
		if i == 0 && lastLine != "" && !strings.HasSuffix(lastLine, "\n") {
			// deduct the line count of the previous line
			deduction := o.displayLength(lastLine) / width
			if deduction == 0 {
				deduction = 1
			}
			o.lines -= int(deduction)

			// concatenate previous and new line
			lineContent = lastLine + lineContent

			// replace last entry of `this.content` with the new expanded line
			o.content = append(o.content[:len(o.content)-1], lineContent)
		} else {
			// otherwise just add the new content
			o.content = append(o.content, lineContent)
		}
	}

	o.lines += linesAdded

	return linesAdded
}

func (o *ConsoleSectionOutput) AddNewLineOfInputSubmit() {
	o.content = append(o.content, "\n")
	o.lines++
}

func (o *ConsoleSectionOutput) DoWrite(message string, newLine bool) {
	// Simulate newline behavior for consistent output formatting, avoiding extra logic
	if !newLine && strings.HasSuffix(message, "\n") {
		message = message[0 : len(message)-1]
		newLine = true
	}

	if !o.IsDecorated() {
		o.StreamOutput.DoWrite(message, newLine)
		return
	}

	// Check if the previous line (last entry of `this.content`) needs to be continued
	// (i.e. does not end with a line break). In which case, it needs to be erased first.
	var lastLine string
	if len(o.content) > 0 {
		lastLine = o.content[len(o.content)-1]
	}

	deleteLastLine := lastLine != "" && !strings.HasSuffix(lastLine, "\n")

	var linesToClear int
	if deleteLastLine {
		linesToClear = 1
	}

	linesAdded := o.AddContent(message, newLine)
	lineOverflow := o.maxHeight > 0

	if lineOverflow && o.lines > o.maxHeight {
		// on overflow, clear the whole section and redraw again (to remove the first lines)
		linesToClear = o.maxHeight
	}

	erasedContent := o.popStreamContentUntilCurrentSection(linesToClear)

	if lineOverflow {
		// redraw existing lines of the section
		previousLinesOfSection := o.content[o.lines-o.maxHeight : o.maxHeight-linesAdded]
		o.StreamOutput.DoWrite(strings.Join(previousLinesOfSection, ""), false)
	}

	if deleteLastLine {
		o.StreamOutput.DoWrite(lastLine+message, true)
	} else {
		o.StreamOutput.DoWrite(message, true)
	}

	o.StreamOutput.DoWrite(erasedContent, true)
}

func (o *ConsoleSectionOutput) popStreamContentUntilCurrentSection(numberOfLinesToClearFromCurrentSection int) string {
	numberOfLinesToClear := numberOfLinesToClearFromCurrentSection
	erasedContent := make([]string, 0)

	for _, section := range o.sections {
		if section == o {
			break
		}

		if section.maxHeight > 0 {
			numberOfLinesToClear += min(section.lines, section.maxHeight)
		} else {
			numberOfLinesToClear += section.lines
		}

		sectionContent := section.VisibleContent()
		if sectionContent != "" {
			if !strings.HasSuffix(sectionContent, "\n") {
				sectionContent += "\n"
			}

			erasedContent = append(erasedContent, sectionContent)
		}
	}

	if numberOfLinesToClear > 0 {
		// move cursor up n lines
		o.StreamOutput.DoWrite(fmt.Sprintf("\x1b[%dA", numberOfLinesToClear), false)

		o.StreamOutput.DoWrite("\x1b[0J", false)
	}

	reversed := make([]string, 0, len(erasedContent))
	for i := len(erasedContent) - 1; i >= 0; i-- {
		reversed = append(reversed, erasedContent[i])
	}

	return strings.Join(reversed, "")
}

func (o *ConsoleSectionOutput) displayLength(text string) int {
	return helper.Width(formatter.RemoveDecoration(o.Formatter(), strings.Replace(text, "\t", "        ", 1)))
}
