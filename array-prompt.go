package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
	"github.com/michielnijenhuis/cli/helper/keys"
	"github.com/michielnijenhuis/cli/terminal"
)

type ArrayPrompt struct {
	*TextPrompt
	Label        string
	Placeholder  string
	DefaultValue string
	Hint         string
	Values       []string
	deletePrompt *SelectPrompt // todo: make multiselect, with search
}

func NewArrayPrompt(i *Input, o *Output, label string, defaultValue []string) *ArrayPrompt {
	p := &ArrayPrompt{
		TextPrompt: &TextPrompt{
			Prompt:      NewPrompt(i, o),
			Placeholder: "Add new value",
		},
		Label:  label,
		Values: make([]string, 0),
	}

	p.GetValue = func() string {
		return p.TypedValue()
	}

	// TODO: fix required handling
	p.Validator = func(value string) string {
		if !p.Required {
			return ""
		}

		if len(p.Values) == 0 && strings.TrimSpace(value) == "" {
			p.validated = false
			return "At least one value is required."
		}

		return ""
	}

	p.trackTypedValue(strings.Join(defaultValue, ", "), false, nil, false)
	p.allowValueClearance = true

	p.on("key", func(key string) {
		if p.State == PromptStateDeleting {
			return
		}

		if keys.Is(key, keys.Enter) {
			if p.TypedValue() == "" {
				p.submit()
			} else {
				p.Values = append(p.Values, p.TypedValue())
				p.SetValue("")
			}

			return
		}

		if keys.Is(key, keys.ShiftTab) && len(p.Values) > 0 {
			p.State = PromptStateDeleting
			p.Prompt.render(p.String)

			if p.deletePrompt == nil {
				p.deletePrompt = NewSelectPrompt(p.input, p.output, "Delete values?", p.Values, nil, "")
				p.deletePrompt.isChild = true

				toDelete, err := p.deletePrompt.Render()

				if err == nil && toDelete != "" {
					i := array.IndexOf(p.Values, toDelete)
					if i >= 0 {
						p.Values = append(p.Values[:i], p.Values[i+1:]...)
					}
				}

				p.State = PromptStateActive
				p.deletePrompt = nil
				lines := strings.Split(p.prevFrame, Eol)
				p.cursor.Move(-999, (-1*len(lines))+2)
				p.eraseDown()
			}

			return
		}
	})

	return p
}

func (p *ArrayPrompt) String() string {
	renderer := NewRenderer()
	terminalWidth := terminal.Columns()
	maxWidth := terminalWidth - 6
	state := p.State

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", p.Label, strings.Join(p.Values, ", ")), false)
	} else if state == PromptStateDeleting {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", p.Label), true)
		renderer.Line("<fg=red>Deleting:</>", true)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", p.Label, p.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", p.Label, p.Error), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>></> %s", p.ValueWithCursor(maxWidth)), true)
	} else {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", p.Label, strings.Join(p.Values, ", ")), true)
		renderer.Line(fmt.Sprintf("<fg=cyan>›</> %s", p.ValueWithCursor(maxWidth)), true)
	}

	return renderer.ToString(p.State)
}

func (p *ArrayPrompt) Render() ([]string, error) {
	_, err := p.Prompt.doPrompt(p.String)
	if err != nil {
		return nil, err
	}

	return p.Values, nil
}
