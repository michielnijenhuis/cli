package cli

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/michielnijenhuis/cli/helper/keys"
)

// May be a map[string]string or []string
type SearchResult interface{}

type SearchPrompt struct {
	*Prompt
	Label         string
	Options       func(string) SearchResult
	Placeholder   string
	Scroll        int
	Hint          string
	matchedValues []string
	matchedLabels []string
}

const reservedLines = 7

func NewSearchPrompt(i *Input, o *Output, label string, options func(string) SearchResult, placeholder string) *SearchPrompt {
	if options == nil {
		panic("options cannot be nil")
	}

	p := &SearchPrompt{
		Prompt:      NewPrompt(i, o),
		Label:       label,
		Options:     options,
		Placeholder: placeholder,
	}
	p.Prompt.Required = true
	p.Prompt.Scroll = 5

	p.GetValue = func() string {
		if p.Highlighted < 0 || len(p.Matches()) == 0 {
			return ""
		}

		index := p.Highlighted
		if index < 0 || index >= len(p.Matches()) {
			return ""
		}

		return p.matchedValues[index]
	}

	p.trackTypedValue("", false, func(key string) bool {
		return (keys.Is(key, keys.CtrlA, keys.Home...) || keys.Is(key, keys.CtrlE, keys.End...)) && p.Highlighted >= 0
	}, false)

	p.InitializeScrolling(-1, reservedLines)

	p.on("key", func(key string) {
		if keys.Is(key, keys.Up, keys.UpArrow, keys.ShiftTab, keys.CtrlP) {
			p.HighlightPrevious(len(p.matchedValues))
		} else if keys.Is(key, keys.Down, keys.DownArrow, keys.Tab, keys.CtrlN) {
			p.HighlightNext(len(p.matchedValues))
		} else if keys.Is(key, keys.CtrlA, keys.Home...) {
			if p.Highlighted >= 0 {
				p.Highlight(0)
			}
		} else if keys.Is(key, keys.CtrlE, keys.End...) {
			if p.Highlighted >= 0 {
				p.Highlight(len(p.matchedValues) - 1)
			}
		} else if keys.Is(key, keys.Enter) {
			if p.Highlighted >= 0 {
				p.submit()
			} else {
				p.search()
			}
		} else if keys.Is(key, keys.Left, keys.LeftArrow, keys.Right, keys.RightArrow, keys.CtrlB, keys.CtrlF) {
			p.Prompt.Highlighted = -1
		} else {
			p.search()
		}
	})

	return p
}

func (p *SearchPrompt) search() {
	p.Prompt.State = PromptStateSearching
	p.Prompt.Highlighted = -1
	p.render(p.View)
	p.matchedLabels = nil
	p.matchedValues = nil
	p.Prompt.FirstVisible = 0
	p.Prompt.State = PromptStateActive
}

func (p *SearchPrompt) View() string {
	renderer := NewRenderer()
	terminalWidth, _ := TerminalWidth()
	maxWidth := terminalWidth - 6
	state := p.State

	label := Truncate(p.Label, maxWidth, "")

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", Dim(label), Truncate(p.SelectedLabel(), maxWidth, "")), false)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", Dim(label), p.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", label, Truncate(p.Error, maxWidth, "")), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", p.ValueWithCursor(maxWidth)), true)
		renderer.Line(p.renderOptions(), true)
		if p.Hint != "" {
			renderer.Line(fmt.Sprintf("<fg=gray>%s</>", p.Hint), true)
		}
	} else if state == PromptStateSearching {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> (searching)", label), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", p.ValueWithCursorAndSearchcon(maxWidth)), true)
		renderer.Line(p.renderOptions(), true)
		if p.Hint != "" {
			renderer.Line(fmt.Sprintf("<fg=gray>%s</>", p.Hint), true)
		}
	} else {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", label), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", p.ValueWithCursor(maxWidth)), true)
		renderer.Line(p.renderOptions(), true)
		if p.Hint != "" {
			renderer.Line(fmt.Sprintf("<fg=gray>%s</>", p.Hint), true)
		} else {
			renderer.NewLine(1)
		}
		renderer.Line(p.spaceForDropdown(), true)
	}

	return renderer.ToString(state)
}

func (p *SearchPrompt) ValueWithCursor(maxWidth int) string {
	if p.Highlighted >= 0 {
		if p.typedValue == "" {
			return Dim(Truncate(p.Placeholder, maxWidth, ""))
		} else {
			return Truncate(p.typedValue, maxWidth, "")
		}
	}

	if p.typedValue == "" {
		return Dim(p.AddCursor(p.Placeholder, 0, maxWidth))
	}

	return p.AddCursor(p.typedValue, p.CursorPosition(), maxWidth)
}

func (p *SearchPrompt) ValueWithCursorAndSearchcon(maxWidth int) string {
	matches := p.Matches()
	longest := Longest(matches, maxWidth, 2)
	w := min(maxWidth, longest)
	padded := Pad(p.ValueWithCursor(maxWidth-1)+"  ", w, "")
	re := regexp.MustCompile(`\s$`)

	return re.ReplaceAllString(padded, "<fg=cyan>…</>")
}

func (p *SearchPrompt) Matches() []string {
	if p.matchedLabels == nil {
		result := p.Options(p.typedValue)
		if result == nil {
			p.matchedLabels = []string{}
			p.matchedValues = []string{}
			return p.matchedLabels
		}

		switch r := result.(type) {
		case []string:
			p.matchedLabels = r
			p.matchedValues = r
		case map[string]string:
			p.matchedLabels = make([]string, 0, len(r))
			p.matchedValues = make([]string, 0, len(r))
			for key, val := range r {
				p.matchedValues = append(p.matchedValues, key)
				p.matchedLabels = append(p.matchedLabels, val)
			}
		default:
			panic("invalid search result")
		}

		sort.Strings(p.matchedLabels)
		sort.Strings(p.matchedValues)
	}

	return p.matchedLabels
}

func (p *SearchPrompt) Visible() []string {
	matches := p.Matches()
	length := len(matches)
	if length == 0 {
		return matches
	}
	firstVisible := p.Prompt.FirstVisible
	// if firstVisible < 0 {
	// 	firstVisible = 0
	// }
	scroll := p.Prompt.Scroll
	// if scroll < 0 {
	// 	scroll = 0
	// } else if scroll > length {
	// 	scroll = length
	// }

	return matches[firstVisible:scroll]
}

func (p *SearchPrompt) SearchValue() string {
	return p.typedValue
}

func (p *SearchPrompt) SelectedLabel() string {
	labels := p.matchedLabels
	length := len(labels)
	index := p.Highlighted
	if index < 0 || index >= length {
		return ""
	}
	return labels[index]
}

func (p *SearchPrompt) renderOptions() string {
	if p.SearchValue() == "" && len(p.Matches()) == 0 {
		text := "No results."
		if p.Prompt.State == PromptStateSearching {
			text = "Searching..."
		}
		return fmt.Sprintf("<fg=gray>  %s</>", text)
	}

	terminalWidth, _ := TerminalWidth()
	matches := p.Matches()
	visible := p.Visible()
	items := make([]string, len(visible))
	for i, item := range visible {
		item = Truncate(item, terminalWidth-10, "")
		if p.Prompt.Highlighted == i {
			item = fmt.Sprintf("<fg=cyan>%s</> %s ", ChevronSmall, item)
		} else {
			item = fmt.Sprintf("  %s  ", Dim(item))
		}
		items[i] = item
	}

	return strings.Join(ScrollBar(items, p.Prompt.FirstVisible, p.Prompt.Scroll, len(matches), min(Longest(matches, maxLineLength, 4), terminalWidth-6), ""), Eol)
}

func (p *SearchPrompt) spaceForDropdown() string {
	if p.SearchValue() != "" {
		return ""
	}

	terminalHeight := TerminalHeight()

	newLines := min(p.Prompt.Scroll, terminalHeight-len(p.Matches()))
	if len(p.Matches()) == 0 {
		newLines++
	}
	newLines = max(newLines, 0)

	return strings.Repeat(Eol, newLines)
}

func (p *SearchPrompt) Render() (string, error) {
	return p.Prompt.doPrompt(p.View)
}
