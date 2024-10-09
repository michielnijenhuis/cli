package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
	"github.com/michielnijenhuis/cli/helper/keys"
	"github.com/michielnijenhuis/cli/terminal"
)

type SelectPrompt struct {
	*Prompt
	Label        string
	Values       []string
	Labels       []string
	DefaultValue string
	Scroll       int
	Hint         string
}

func NewSelectPrompt(i *Input, o *Output, label string, values []string, labels []string, defaultValue string) *SelectPrompt {
	if labels == nil {
		labels = values
	}

	p := &SelectPrompt{
		Prompt:       NewPrompt(i, o),
		Label:        label,
		Values:       values,
		Labels:       labels,
		DefaultValue: defaultValue,
		Scroll:       5,
	}

	p.Required = true
	p.GetValue = func() string {
		if !terminal.IsInteractive() {
			return p.DefaultValue
		}

		if p.Highlighted < 0 || p.Highlighted >= len(p.Values) {
			return ""
		}

		return p.Values[p.Highlighted]
	}

	if defaultValue != "" {
		i := array.IndexOf(values, defaultValue)
		if i < 0 {
			i = 0
		}
		p.InitializeScrolling(i, 0)
		p.ScrollToHighlighted(len(values))
	} else {
		p.InitializeScrolling(0, 0)
	}

	p.on("key", func(key string) {
		total := len(p.Values)

		if keys.Is(key, keys.Up, keys.UpArrow, keys.Left, keys.LeftArrow, keys.ShiftTab, keys.CtrlP, keys.CtrlB, "k", "h") {
			p.HighlightPrevious(total)
			return
		}

		if keys.Is(key, keys.Down, keys.DownArrow, keys.Right, keys.RightArrow, keys.Tab, keys.CtrlN, keys.CtrlF, "j", "l") {
			p.HighlightNext(total)
			return
		}

		if keys.Is(key, keys.CtrlE) || keys.Is(key, keys.End...) {
			p.Highlight(total - 1)
			return
		}

		if keys.Is(key, keys.Enter) {
			p.submit()
			return
		}
	})

	return p
}

func (p *SelectPrompt) GetLabel() string {
	if p.Highlighted < 0 {
		return ""
	}

	return p.Labels[p.Highlighted]
}

func (p *SelectPrompt) Visible() []string {
	firstVisible := p.FirstVisible
	if firstVisible < 0 {
		firstVisible = 0
	} else if firstVisible >= len(p.Labels) {
		firstVisible = max(0, len(p.Labels)-1)
	}

	scroll := min(p.Scroll, len(p.Labels))

	return p.Labels[firstVisible:scroll]
}

func (p *SelectPrompt) String() string {
	renderer := NewRenderer()
	maxWidth := terminal.Columns() - 6
	state := p.State

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", Truncate(p.Label, maxWidth, ""), p.GetLabel()), false)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", Truncate(p.Label, maxWidth, ""), p.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", Truncate(p.Label, maxWidth, ""), p.Error), true)
	} else {
		if !p.isChild {
			renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", Truncate(p.Label, maxWidth, "")), true)
		}

		renderer.Line(p.renderOptions(), true)
	}

	return renderer.ToString(state)
}

func (p *SelectPrompt) renderOptions() string {
	width := terminal.Columns()
	visible := p.Visible()
	values := p.Values
	color := ColorCyan
	if p.State == PromptStateCancel {
		color = "dim"
	}

	options := make([]string, 0, len(visible))

	for _, v := range visible {
		index := array.IndexOf(visible, v)
		label := Truncate(v, width-12, "")

		if p.State == PromptStateCancel {
			if p.Highlighted == index {
				label = fmt.Sprintf("› ● %s  ", Strikethrough(label))
			} else {
				label = fmt.Sprintf("  ○ %s  ", Strikethrough(label))
			}
		} else {
			if p.Highlighted == index {
				label = fmt.Sprintf("<fg=cyan>›</> <fg=cyan>●</> %s  ", label)
			} else {
				label = fmt.Sprintf("  %s %s  ", Dim("○"), Dim(label))
			}
		}

		options = append(options, label)
	}

	return strings.Join(ScrollBar(options, p.FirstVisible, p.Scroll, len(values), min(Longest(values, -1, 6), width-6), color), Eol)
}

func (p *SelectPrompt) Render() (string, error) {
	return p.Prompt.doPrompt(p.String)
}
