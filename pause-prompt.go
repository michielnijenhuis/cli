package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper/keys"
)

type PausePrompt struct {
	*Prompt
	Message string
}

func NewPausePrompt(i *Input, o *Output, message string) *PausePrompt {
	if message == "" {
		message = "Press enter to continue..."
	}

	p := &PausePrompt{
		Prompt:  NewPrompt(i, o),
		Message: message,
	}

	p.on("key", func(key string) {
		if keys.Is(key, keys.Enter) {
			p.submit()
		}
	})

	return p
}

func (p *PausePrompt) String() string {
	renderer := NewRenderer()

	if p.State == PromptStateCancel || p.State == PromptStateError {
		renderer.Line("<fg=red>Aborted.</>", true)
	} else {
		color := "green"
		eolOnLastLine := true

		if p.State == PromptStateSubmit {
			color = "gray"
			eolOnLastLine = false
		}

		lines := strings.Split(p.Message, Eol)
		for i, line := range lines {
			eol := true
			if i == len(lines)-1 && !eolOnLastLine {
				eol = false
			}

			renderer.Line(fmt.Sprintf("<fg=%s>%s</>", color, line), eol)
		}
	}

	return renderer.ToString(p.State)
}

func (p *PausePrompt) Render() (bool, error) {
	_, err := p.Prompt.doPrompt(p.String)
	if err != nil {
		return false, err
	}
	return true, nil
}
