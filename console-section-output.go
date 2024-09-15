package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

type ConsoleSectionOutput struct {
	*Output
	content   []string
	lines     int
	sections  []*ConsoleSectionOutput
	maxHeight int
}

func NewConsoleSectionOutput(o *Output, sections []*ConsoleSectionOutput) *ConsoleSectionOutput {
	if sections == nil {
		sections = make([]*ConsoleSectionOutput, 0)
	}

	cso := &ConsoleSectionOutput{
		Output: o,
	}

	helper.Unshift(&sections, cso)

	cso.sections = sections

	return cso
}

func (c *ConsoleSectionOutput) SetMaxHeight(maxHeight int) {
	prev := c.maxHeight
	c.maxHeight = maxHeight

	var existingContent string
	if prev != 0 {
		existingContent = c.popStreamContentUntilCurrentSection(min(prev, c.lines))
	} else {
		existingContent = c.popStreamContentUntilCurrentSection(c.lines)
	}

	c.Output.DoWrite(c.VisibleContent(), false)
	c.Output.DoWrite(existingContent, false)
}

func (c *ConsoleSectionOutput) Clear(lines int) {
	if len(c.content) == 0 || !c.IsDecorated() {
		return
	}

	if lines >= 0 {
		c.content = c.content[:len(c.content)-lines-1]
	} else {
		lines = c.lines
		c.content = make([]string, 0)
	}

	c.lines -= lines

	var existingContent string
	if c.maxHeight != 0 {
		existingContent = c.popStreamContentUntilCurrentSection(min(c.maxHeight, c.lines))
	} else {
		existingContent = c.popStreamContentUntilCurrentSection(c.lines)
	}

	c.Output.DoWrite(existingContent, false)
}

func (c *ConsoleSectionOutput) Overwrite(message string) {
	c.Clear(-1)
	c.Writeln(message, 0)
}

func (c *ConsoleSectionOutput) OverwriteMany(message []string) {
	c.Clear(-1)
	c.Writelns(message, 0)
}

func (c *ConsoleSectionOutput) Content() string {
	return strings.Join(c.content, "")
}

func (c *ConsoleSectionOutput) VisibleContent() string {
	if c.maxHeight <= 0 {
		return c.Content()
	}

	return strings.Join(c.content[:len(c.content)-1-c.maxHeight], "")
}

func (c *ConsoleSectionOutput) AddContent(input string, newLine bool) int {
	width, _ := TerminalWidth()
	lines := strings.Split(input, Eol)
	linesAdded := 0
	count := len(lines) - 1

	for i, lineContent := range lines {
		if i < count || newLine {
			lineContent += Eol
		}

		if lineContent == "" {
			continue
		}

		if i == 0 && len(c.content) > 0 && !strings.HasSuffix(c.content[len(c.content)-1], Eol) {
			lastLine := c.content[len(c.content)-1]
			c.lines -= max(1, c.displayLength(lastLine)/width)
			lineContent = lastLine + lineContent
			c.content[len(c.content)-1] = lineContent
		} else {
			c.content = append(c.content, lineContent)
		}

		linesAdded += max(1, c.displayLength(lineContent)/width)
	}

	c.lines += linesAdded

	return linesAdded
}

func (c *ConsoleSectionOutput) AddNewLineOfInputSubmit() {
	c.content = append(c.content, Eol)
	c.lines++
}

func (c *ConsoleSectionOutput) DoWrite(message string, newLine bool) {
	if !newLine && strings.HasSuffix(message, Eol) {
		message = message[:len(message)-len(Eol)]
		newLine = true
	}

	if !c.Output.IsDecorated() {
		c.Output.DoWrite(message, newLine)
		return
	}

	var deleteLastLine bool
	var lastLine string
	var linesToClear int

	if len(c.content) > 0 {
		lastLine = c.content[len(c.content)-1]
	}

	deleteLastLine = lastLine != "" && !strings.HasSuffix(lastLine, Eol)
	if deleteLastLine {
		linesToClear = 1
	}

	linesAdded := c.AddContent(message, newLine)

	lineOverflow := c.maxHeight > 0 && c.lines > c.maxHeight
	if lineOverflow {
		linesToClear = c.maxHeight
	}

	erasedContent := c.popStreamContentUntilCurrentSection(linesToClear)

	if lineOverflow {
		previousLinesOfSection := c.content[c.lines-c.maxHeight : c.maxHeight-linesAdded]
		c.Output.DoWrite(strings.Join(previousLinesOfSection, ""), false)
	}

	if deleteLastLine {
		c.Output.DoWrite(lastLine+message, true)
	} else {
		c.Output.DoWrite(message, true)
	}

	c.Output.DoWrite(erasedContent, false)
}

func (c *ConsoleSectionOutput) popStreamContentUntilCurrentSection(numberOfLinesToClearFromCurrentSection int) string {
	numberOfLinesToClear := numberOfLinesToClearFromCurrentSection
	erasedContent := make([]string, 0)

	for _, section := range c.sections {
		if section == c {
			break
		}

		if section.maxHeight > 0 {
			numberOfLinesToClear += min(section.lines, section.maxHeight)
		} else {
			numberOfLinesToClear += section.lines
		}

		if sectionContent := section.VisibleContent(); sectionContent != "" {
			if !strings.HasSuffix(sectionContent, Eol) {
				sectionContent += Eol
			}

			erasedContent = append(erasedContent, sectionContent)
		}
	}

	if numberOfLinesToClear > 0 {
		c.Output.DoWrite(fmt.Sprintf("\x1b[%dA", numberOfLinesToClear), false)
		c.Output.DoWrite("\x1b[0J", false)
	}

	slices.Reverse(erasedContent)
	return strings.Join(erasedContent, "")
}

func (c *ConsoleSectionOutput) displayLength(text string) int {
	f := c.Formatter()
	return helper.Width(f.RemoveDecoration(strings.ReplaceAll(text, "\t", "        ")))
}
