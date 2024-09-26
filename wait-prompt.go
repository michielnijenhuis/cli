package cli

import (
	"fmt"
)

type WaitPrompt struct {
	*Spinner
	WaitFunc  func() bool
	cancelled bool
	done      bool
}

func NewWaitPrompt(i *Input, o *Output, waitFunc func() bool, message string) *WaitPrompt {
	p := &WaitPrompt{
		Spinner:  NewSpinner(i, o, message, nil, ""),
		WaitFunc: waitFunc,
	}

	p.CancelUsingFn = func() any {
		return nil
	}

	return p
}

func (p *WaitPrompt) String() string {
	renderer := NewRenderer()
	state := p.State

	if state == PromptStateCancel {
		renderer.Line(fmt.Sprintf("<fg=yellow>%s</>", p.CancelMessage), true)
	}

	return renderer.ToString(state)
}

func (p *WaitPrompt) Render() error {
	go func(p *WaitPrompt) {
		p.Spin(func() {
			for !p.WaitFunc() && !p.cancelled {
				//
			}
		})

		p.done = true
	}(p)

	_, err := p.doPrompt(p.String)
	return err
}
