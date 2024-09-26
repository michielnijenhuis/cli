package cli

import (
	"fmt"

	"github.com/michielnijenhuis/cli/terminal"
)

type TextPrompt struct {
	*Prompt
	Label        string
	Placeholder  string
	DefaultValue string
	Hint         string
}

func NewTextPrompt(i *Input, o *Output, label string, defaultValue string) *TextPrompt {
	p := &TextPrompt{
		Prompt: NewPrompt(i, o),
		Label:  label,
	}

	p.GetValue = func() string {
		return p.TypedValue()
	}
	p.trackTypedValue(defaultValue, true, nil, false)
	p.allowValueClearance = true
	p.Required = true

	return p
}

func (p *TextPrompt) ValueWithCursor(maxWidth int) string {
	if p.value() == "" {
		return Dim(p.AddCursor(p.Placeholder, 0, maxWidth))
	}

	return p.AddCursor(p.value(), p.cursorPosition, maxWidth)
}

func (p *TextPrompt) View() string {
	renderer := NewRenderer()
	maxWidth := terminal.Columns() - 6
	state := p.State

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", p.Label, p.value()), false)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", p.Label, p.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", p.Label, p.Error), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>></> %s", p.ValueWithCursor(maxWidth)), true)
	} else {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", p.Label), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", p.ValueWithCursor(maxWidth)), true)
	}

	return renderer.ToString(p.State)
}

func (p *TextPrompt) Render() (string, error) {
	return p.Prompt.doPrompt(p.View)
}
