package cli

import "fmt"

type TextPrompt struct {
	*Prompt
	Label        string
	Placeholder  string
	DefaultValue string
	Hint         string
}

func NewTextPrompt(i *Input, o *Output, label string, defaultValue string) *TextPrompt {
	tp := &TextPrompt{
		Prompt: NewPrompt("text-prompt", i, o),
		Label:  label,
	}

	tp.Value = func() string {
		return tp.typedValue
	}
	tp.trackTypedValue(defaultValue, true, nil, false)
	tp.allowValueClearance = true
	tp.Required = true

	return tp
}

func (tp *TextPrompt) ValueWithCursor(maxWidth int) string {
	if tp.value() == "" {
		return Dim(tp.addCursor(tp.Placeholder, 0, maxWidth))
	}

	return tp.addCursor(tp.value(), tp.cursorPosition, maxWidth)
}

func (tp *TextPrompt) String() string {
	renderer := NewRenderer()
	terminalWidth, _ := TerminalWidth()
	maxWidth := terminalWidth - 6
	state := tp.State

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", tp.Label, tp.value()), false)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", tp.Label, tp.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", tp.Label, tp.Error), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>></> %s", tp.ValueWithCursor(maxWidth)), true)
	} else {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", tp.Label), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", tp.ValueWithCursor(maxWidth)), true)
	}

	return renderer.ToString(tp.State)
}

func (tp *TextPrompt) Render() (string, error) {
	return tp.Prompt.doPrompt(tp.String)
}
