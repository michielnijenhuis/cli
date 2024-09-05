package cli

import (
	"fmt"

	"github.com/michielnijenhuis/cli/helper/keys"
)

type ConfirmPrompt struct {
	*Prompt
	Label        string
	DefaultValue bool
	Yes          string
	No           string
	Hint         string
	Confirmed    bool
}

func NewConfirmPrompt(i *Input, o *Output, label string) *ConfirmPrompt {
	cp := &ConfirmPrompt{
		Prompt:       NewPrompt("confirm-prompt", i, o),
		Label:        label,
		DefaultValue: true,
		Confirmed:    true,
		Yes:          "yes",
		No:           "no",
	}

	cp.Value = func() string {
		if cp.Confirmed {
			return cp.Yes
		}

		return cp.No
	}

	cp.Validator = func(s string) string {
		if !cp.Required || s == cp.Yes {
			return ""
		}

		return "Required."
	}

	cp.on("key", func(key string) {
		if keys.Is(key, "y") {
			cp.Confirmed = true
			return
		}

		if keys.Is(key, "n") {
			cp.Confirmed = false
			return
		}

		if keys.Is(key, keys.Tab, keys.Up, keys.UpArrow, keys.Down, keys.DownArrow, keys.Left, keys.LeftArrow, keys.Right, keys.RightArrow, keys.CtrlP, keys.CtrlF, keys.CtrlN, keys.CtrlB, "h", "j", "k", "l") {
			cp.Confirmed = !cp.Confirmed
			return
		}

		if keys.Is(key, keys.Enter) {
			cp.submit()
		}
	})

	return cp
}

func (cp *ConfirmPrompt) String() string {
	renderer := NewRenderer()
	state := cp.State

	if state == PromptStateSubmit {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=cyan>%s</>", cp.Label, cp.value()), false)
	} else if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=yellow>%s</>", cp.Label, cp.CancelMessage), true)
	} else if state == PromptStateError {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</> <fg=red>%s</>", cp.Label, cp.Error), true)
		renderer.Line(cp.renderOptions(), true)
	} else {
		renderer.Line(fmt.Sprintf("<fg=green>?</> <options=bold>%s</>", cp.Label), true)
		renderer.Line(cp.renderOptions(), true)
	}

	return renderer.ToString(state)
}

func (cp *ConfirmPrompt) renderOptions() string {
	terminalWidth, _ := TerminalWidth()
	length := (terminalWidth - 14) / 2
	yes := Truncate(cp.Yes, length, "")
	no := Truncate(cp.No, length, "")

	var text string

	if cp.State == PromptStateCancel {
		if cp.Confirmed {
			text = fmt.Sprintf("● %s / ○ %s", Strikethrough(yes), Strikethrough(no))
		} else {
			text = fmt.Sprintf("○ %s / ● %s", Strikethrough(yes), Strikethrough(no))
		}

		return Dim(text)
	}

	if cp.Confirmed {
		return fmt.Sprintf("<fg=cyan>●</> %s %s", cp.Yes, Dim("/ ○ "+no))
	}

	return fmt.Sprintf("%s <fg=cyan>●</> %s", Dim("○ "+yes+" /"), no)
}

func (cp *ConfirmPrompt) Render() (bool, error) {
	value, err := cp.Prompt.doPrompt(cp.String)
	boolean := value == cp.Yes
	return boolean, err
}
