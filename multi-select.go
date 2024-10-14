package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
	"github.com/michielnijenhuis/cli/helper/keys"
	"github.com/michielnijenhuis/cli/terminal"
)

// []string or map[string]string
type MultiSelectOptions interface{}

type MultiSelectPrompt struct {
	*Prompt
	Label         string
	Hint          string
	labels        []string
	values        []string
	labelToValues map[string]string
	selected      []string
}

func NewMultiSelectPrompt(i *Input, o *Output, label string, options MultiSelectOptions, defaultValues []string) *MultiSelectPrompt {
	var labels []string
	var values []string

	switch t := options.(type) {
	case []string:
		sort.Strings(t)
		labels = t
		values = t
	case map[string]string:
		labelToValue := make(map[string]string)
		labels = make([]string, 0, len(t))
		for value, label := range t {
			labelToValue[label] = value
			labels = append(labels, label)
		}
		sort.Strings(labels)
		values = make([]string, 0, len(t))
		for _, label := range labels {
			values = append(values, labelToValue[label])
		}
	default:
		labels = []string{}
		values = []string{}
	}

	p := &MultiSelectPrompt{
		Prompt:   NewPrompt(i, o),
		Label:    label,
		labels:   labels,
		values:   values,
		selected: make([]string, 0),
	}

	p.Scroll = 5

	p.labelToValues = make(map[string]string)
	for i, v := range p.labels {
		p.labelToValues[v] = p.values[i]
	}

	if defaultValues != nil {
		p.selected = defaultValues
	}

	p.InitializeScrolling(0, 0)

	p.on("key", func(key string) {
		switch {
		case keys.Is(key, keys.Up, keys.UpArrow, keys.Left, keys.LeftArrow, keys.ShiftTab, keys.CtrlP, keys.CtrlB, "k", "h"):
			p.HighlightPrevious(len(p.values))
		case keys.Is(key, keys.Down, keys.DownArrow, keys.Right, keys.RightArrow, keys.Tab, keys.CtrlN, keys.CtrlF, "j", "l"):
			p.HighlightNext(len(p.values))
		case keys.Is(key, keys.Home...) || keys.Is(key, keys.CtrlA):
			p.Highlight(0)
		case keys.Is(key, keys.End...) || keys.Is(key, keys.CtrlE):
			p.Highlight(len(p.values) - 1)
		case keys.Is(key, keys.Space):
			p.toggleHighlighted()
		case keys.Is(key, keys.CtrlA):
			p.toggleAll()
		case keys.Is(key, keys.Enter):
			p.submit()
		}
	})

	p.GetValue = func() string {
		return "IDK"
	}

	return p
}

func (p *MultiSelectPrompt) SetRequired() {
	p.Prompt.Required = true
}

func (p *MultiSelectPrompt) toggleHighlighted() {
	if p.Highlighted < 0 {
		return
	}

	value := p.values[p.Highlighted]
	if p.IsSelected(value) {
		p.selected = array.Remove(p.selected, value)
	} else {
		p.selected = append(p.selected, value)
	}
}

func (p *MultiSelectPrompt) toggleAll() {
	if len(p.selected) == len(p.values) {
		p.selected = make([]string, 0)
	} else {
		p.selected = make([]string, len(p.values))
		copy(p.selected, p.values)
	}
}

func (p *MultiSelectPrompt) SelectedLabels() []string {
	labels := make([]string, 0, len(p.selected))
	for i, v := range p.values {
		for _, s := range p.selected {
			if v == s {
				labels = append(labels, p.labels[i])
			}
		}
	}
	return labels
}

func (p *MultiSelectPrompt) Visible() []string {
	length := len(p.labels)
	if length == 0 {
		return p.labels
	}

	start := max(0, p.FirstVisible)
	end := min(length, start+p.Scroll)
	return p.labels[start:end]
}

func (p *MultiSelectPrompt) IsHighlighted(value string) bool {
	i := array.IndexOf(p.values, value)
	if i < 0 {
		return false
	}

	return p.Highlighted == i
}

func (p *MultiSelectPrompt) IsSelected(value string) bool {
	for _, v := range p.selected {
		if v == value {
			return true
		}
	}

	return false
}

func (p *MultiSelectPrompt) View() string {
	renderer := NewRenderer()
	terminalWidth := terminal.Columns()
	maxWidth := terminalWidth - 6
	state := p.State
	label := Truncate(p.Label, maxWidth, "")

	switch state {
	case PromptStateSubmit:
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", Dim(label), Truncate(strings.Join(p.SelectedLabels(), ", "), maxWidth, "")), false)
	case PromptStateCancel:
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", Dim(label), p.CancelMessage), true)
		renderer.Line(p.renderOptions(), true)
	case PromptStateError:
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", label, p.Error), true)
		renderer.Line(p.renderOptions(), true)
	default:
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", label, Truncate(strings.Join(p.SelectedLabels(), ", "), maxWidth, "")), true)
		renderer.Line(p.renderOptions(), true)
		if p.Hint != "" {
			renderer.Line(fmt.Sprintf("<fg=gray>%s</>", p.Hint), true)
		}
	}

	return renderer.ToString(state)
}

func (p *MultiSelectPrompt) Render() ([]string, error) {
	_, err := p.Prompt.doPrompt(p.View)
	if err != nil {
		return []string{}, err
	}

	return p.selected, nil
}

func (p *MultiSelectPrompt) renderOptions() string {
	visible := p.Visible()
	items := make([]string, 0, len(visible))
	terminalWidth := terminal.Columns()

	for _, label := range visible {
		idx := array.IndexOf(p.labels, label)
		value := p.labelToValues[label]
		label = Truncate(label, terminalWidth-12, "")
		active := p.Highlighted == idx
		selected := p.IsSelected(value)

		if p.State == PromptStateCancel {
			var out string

			switch {
			case active && selected:
				out = fmt.Sprintf("%s %s %s  ", SmallTriangleRight, SquareDefault, Strikethrough(label))
			case active && !selected:
				out = fmt.Sprintf("%s %s %s  ", SmallTriangleRight, SquareOutline, Strikethrough(label))
			case !active && selected:
				out = fmt.Sprintf("  %s %s  ", SquareDefault, Strikethrough(label))
			default:
				out = fmt.Sprintf("  %s %s  ", SquareOutline, Strikethrough(label))
			}

			items = append(items, Dim(out))
		} else {
			var out string

			switch {
			case active && selected:
				out = fmt.Sprintf("<fg=cyan>%s %s</> %s  ", SmallTriangleRight, SquareDefault, label)
			case active && !selected:
				out = fmt.Sprintf("<fg=cyan>%s</> %s %s  ", SmallTriangleRight, SquareOutline, label)
			case !active && selected:
				out = fmt.Sprintf("  <fg=cyan>%s</> %s  ", SquareDefault, Dim(label))
			default:
				out = fmt.Sprintf("  %s %s  ", Dim(SquareOutline), Dim(label))
			}

			items = append(items, out)
		}
	}

	color := ColorCyan
	if p.State == PromptStateCancel {
		color = "dim"
	}

	return strings.Join(ScrollBar(items, p.FirstVisible, p.Scroll, len(p.values), min(Longest(p.values, maxLineLength, 4), terminalWidth-6), color), "\n")
}
