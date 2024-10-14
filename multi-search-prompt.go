package cli

// import (
// 	"fmt"
// 	"sort"
// 	"strings"

// 	"github.com/michielnijenhuis/cli/helper/array"
// 	"github.com/michielnijenhuis/cli/helper/keys"
// 	"github.com/michielnijenhuis/cli/terminal"
// )

// // May be a map[string]string or []string
// type MultiSearchResult interface{}

// type MultiSearchPrompt struct {
// 	*Prompt
// 	Label         string
// 	Placeholder   string
// 	Hint          string
// 	Options       func(string) MultiSearchResult
// 	matchedValues []string
// 	matchedLabels []string
// 	selected      []string
// }

// func NewMultiSearchPrompt(i *Input, o *Output, label string, options func(string) MultiSearchResult, placeholder string, defaultValues []string) *MultiSearchPrompt {
// 	p := &MultiSearchPrompt{
// 		Prompt:      NewPrompt(i, o),
// 		Label:       label,
// 		Placeholder: placeholder,
// 		Options:     options,
// 		selected:    make([]string, 0),
// 	}

// 	p.Scroll = 5

// 	if defaultValues != nil {
// 		p.selected = defaultValues
// 	}

// 	p.trackTypedValue("", false, func(key string) bool {
// 		if p.Highlighted < 0 {
// 			return false
// 		}

// 		if keys.Is(key, keys.Home...) || keys.Is(key, keys.End...) || keys.Is(key, keys.Space, keys.CtrlA, keys.CtrlE) {
// 			return true
// 		}

// 		return false
// 	}, false)

// 	p.InitializeScrolling(-1, 0)

// 	p.on("key", func(key string) {
// 		switch {
// 		case keys.Is(key, keys.Up, keys.UpArrow, keys.ShiftTab):
// 			p.HighlightPrevious(len(p.matchedValues))
// 		case keys.Is(key, keys.Down, keys.DownArrow, keys.Tab):
// 			p.HighlightNext(len(p.matchedValues))
// 		case keys.Is(key, keys.Home...):
// 			if p.Highlighted >= 0 {
// 				p.Highlight(0)
// 			}
// 		case keys.Is(key, keys.End...):
// 			if p.Highlighted >= 0 {
// 				p.Highlight(len(p.matchedValues) - 1)
// 			}
// 		case keys.Is(key, keys.Space):
// 			if p.Highlighted >= 0 {
// 				p.toggleHighlighted()
// 			}
// 		case keys.Is(key, keys.CtrlA):
// 			if p.Highlighted >= 0 {
// 				p.toggleAll()
// 			}
// 		case keys.Is(key, keys.CtrlE):
// 			// noop
// 		case keys.Is(key, keys.Enter):
// 			p.submit()
// 		case keys.Is(key, keys.Left, keys.LeftArrow, keys.Right, keys.RightArrow):
// 			p.Highlighted = -1
// 		default:
// 			p.search()
// 		}
// 	})

// 	p.GetValue = func() string {
// 		return "IDK"
// 	}

// 	return p
// }

// func (p *MultiSearchPrompt) SetRequired() {
// 	p.Prompt.Required = true
// }

// func (p *MultiSearchPrompt) Matches() []string {
// 	if p.matchedLabels == nil {
// 		result := p.Options(p.typedValue)
// 		if result == nil {
// 			p.matchedLabels = []string{}
// 			p.matchedValues = []string{}
// 			return p.matchedLabels
// 		}

// 		switch r := result.(type) {
// 		case []string:
// 			sort.Strings(r)
// 			p.matchedLabels = r
// 			p.matchedValues = r
// 		case map[string]string:
// 			labelToValue := make(map[string]string)
// 			p.matchedLabels = make([]string, 0, len(r))
// 			for value, label := range r {
// 				labelToValue[label] = value
// 				p.matchedLabels = append(p.matchedLabels, label)
// 			}
// 			sort.Strings(p.matchedLabels)
// 			p.matchedValues = make([]string, 0, len(r))
// 			for _, label := range p.matchedLabels {
// 				p.matchedValues = append(p.matchedValues, labelToValue[label])
// 			}
// 		default:
// 			p.matchedLabels = []string{}
// 			p.matchedValues = []string{}
// 		}

// 	}

// 	return p.matchedLabels
// }

// func (p *MultiSearchPrompt) ValueWithCursor(maxWidth int) string {
// 	if p.Highlighted >= 0 {
// 		if p.typedValue == "" {
// 			return Dim(Truncate(p.Placeholder, maxWidth, ""))
// 		} else {
// 			return Truncate(p.typedValue, maxWidth, "")
// 		}
// 	}

// 	if p.typedValue == "" {
// 		return Dim(p.AddCursor(p.Placeholder, 0, maxWidth))
// 	}

// 	return p.AddCursor(p.typedValue, p.CursorPosition(), maxWidth)
// }

// func (p *MultiSearchPrompt) toggleHighlighted() {
// 	if p.Highlighted < 0 {
// 		return
// 	}

// 	value := p.values[p.Highlighted]
// 	if p.IsSelected(value) {
// 		p.selected = array.Remove(p.selected, value)
// 	} else {
// 		p.selected = append(p.selected, value)
// 	}
// }

// func (p *MultiSearchPrompt) toggleAll() {
// 	if len(p.selected) == len(p.values) {
// 		p.selected = make([]string, 0)
// 	} else {
// 		p.selected = make([]string, len(p.values))
// 		copy(p.selected, p.values)
// 	}
// }

// func (p *MultiSearchPrompt) SelectedLabels() []string {
// 	labels := make([]string, 0, len(p.selected))
// 	for i, v := range p.values {
// 		for _, s := range p.selected {
// 			if v == s {
// 				labels = append(labels, p.labels[i])
// 			}
// 		}
// 	}
// 	return labels
// }

// func (p *MultiSearchPrompt) Visible() []string {
// 	length := len(p.labels)
// 	if length == 0 {
// 		return p.labels
// 	}

// 	start := max(0, p.FirstVisible)
// 	end := min(length, start+p.Scroll)
// 	return p.labels[start:end]
// }

// func (p *MultiSearchPrompt) IsHighlighted(value string) bool {
// 	i := array.IndexOf(p.values, value)
// 	if i < 0 {
// 		return false
// 	}

// 	return p.Highlighted == i
// }

// func (p *MultiSearchPrompt) IsSelected(value string) bool {
// 	for _, v := range p.selected {
// 		if v == value {
// 			return true
// 		}
// 	}

// 	return false
// }

// func (p *MultiSearchPrompt) View() string {
// 	renderer := NewRenderer()
// 	terminalWidth := terminal.Columns()
// 	maxWidth := terminalWidth - 6
// 	state := p.State
// 	label := Truncate(p.Label, maxWidth, "")

// 	switch state {
// 	case PromptStateSubmit:
// 		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", Dim(label), Truncate(strings.Join(p.SelectedLabels(), ", "), maxWidth, "")), false)
// 	case PromptStateCancel:
// 		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", Dim(label), p.CancelMessage), true)
// 		renderer.Line(p.renderOptions(), true)
// 	case PromptStateError:
// 		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", label, p.Error), true)
// 		renderer.Line(p.renderOptions(), true)
// 	default:
// 		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", label, Truncate(strings.Join(p.SelectedLabels(), ", "), maxWidth, "")), true)
// 		renderer.Line(p.renderOptions(), true)
// 		if p.Hint != "" {
// 			renderer.Line(fmt.Sprintf("<fg=gray>%s</>", p.Hint), true)
// 		}
// 	}

// 	return renderer.ToString(state)
// }

// func (p *MultiSearchPrompt) Render() ([]string, error) {
// 	_, err := p.Prompt.doPrompt(p.View)
// 	if err != nil {
// 		return []string{}, err
// 	}

// 	return p.selected, nil
// }

// func (p *MultiSearchPrompt) renderOptions() string {
// 	visible := p.Visible()
// 	items := make([]string, 0, len(visible))
// 	terminalWidth := terminal.Columns()

// 	for _, label := range visible {
// 		idx := array.IndexOf(p.labels, label)
// 		value := p.labelToValues[label]
// 		label = Truncate(label, terminalWidth-12, "")
// 		active := p.Highlighted == idx
// 		selected := p.IsSelected(value)

// 		if p.State == PromptStateCancel {
// 			var out string

// 			switch {
// 			case active && selected:
// 				out = fmt.Sprintf("%s %s %s  ", SmallTriangleRight, SquareDefault, Strikethrough(label))
// 			case active && !selected:
// 				out = fmt.Sprintf("%s %s %s  ", SmallTriangleRight, SquareOutline, Strikethrough(label))
// 			case !active && selected:
// 				out = fmt.Sprintf("  %s %s  ", SquareDefault, Strikethrough(label))
// 			default:
// 				out = fmt.Sprintf("  %s %s  ", SquareOutline, Strikethrough(label))
// 			}

// 			items = append(items, Dim(out))
// 		} else {
// 			var out string

// 			switch {
// 			case active && selected:
// 				out = fmt.Sprintf("<fg=cyan>%s %s</> %s  ", SmallTriangleRight, SquareDefault, label)
// 			case active && !selected:
// 				out = fmt.Sprintf("<fg=cyan>%s</> %s %s  ", SmallTriangleRight, SquareOutline, label)
// 			case !active && selected:
// 				out = fmt.Sprintf("  <fg=cyan>%s</> %s  ", SquareDefault, Dim(label))
// 			default:
// 				out = fmt.Sprintf("  %s %s  ", Dim(SquareOutline), Dim(label))
// 			}

// 			items = append(items, out)
// 		}
// 	}

// 	color := ColorCyan
// 	if p.State == PromptStateCancel {
// 		color = "dim"
// 	}

// 	return strings.Join(ScrollBar(items, p.FirstVisible, p.Scroll, len(p.values), min(Longest(p.values, maxLineLength, 4), terminalWidth-6), color), "\n")
// }

// func (p *MultiSearchPrompt) search() {
// 	p.Prompt.State = PromptStateSearching
// 	p.Prompt.Highlighted = -1
// 	p.render(p.View)
// 	p.matchedLabels = nil
// 	p.matchedValues = nil
// 	p.Prompt.FirstVisible = 0
// 	p.Prompt.State = PromptStateActive

// 	if len(p.Matches()) == 1 {
// 		p.Highlight(0)
// 	}
// }
