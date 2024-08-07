package cli

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/michielnijenhuis/cli/helper/keys"
)

const (
	PromptStateInitial   = 1
	PromptStateActive    = 2
	PromptStateSubmit    = 3
	PromptStateCancel    = 4
	PromptStateError     = 5
	PromptStateSearching = 6
)

type RendererInterface interface {
	Render(prompt any)
	String() string
}

type AnyFunc func() any
type Listener func(key string)

type Prompt struct {
	Name                string
	State               uint
	Error               string
	CancelMessage       string
	Required            bool
	Validator           func(string) string
	Scroll              int
	Highlighted         int
	FirstVisible        int
	DefaultValue        string
	Input               *Input
	Output              *Output
	shouldFallback      bool
	activeTheme         string
	prevFrame           string
	validated           bool
	cursor              Cursor
	listeners           map[string][]Listener
	typedValue          string
	allowValueClearance bool
	CancelUsingFn       AnyFunc
	ValidateUsingFn     func(any) string
	RevertUsingFn       AnyFunc
	// fallback            any // TODO
	// cursorPosition      int // TODO
}

func NewPrompt(name string, i *Input, o *Output) *Prompt {
	cursor := Cursor{
		Input:  i.Stream,
		Output: o,
	}

	return &Prompt{
		Name:          name,
		Input:         i,
		Output:        o,
		State:         PromptStateInitial,
		CancelMessage: "Cancelled.",
		activeTheme:   "default",
		cursor:        cursor,
		listeners:     make(map[string][]Listener),
	}
}

func (p *Prompt) Prompt(renderer func() string) (string, error) {
	if p.ShouldFallback() {
		return p.Fallback(), nil
	}

	if !p.Input.IsInteractive() {
		return p.defaultValue()
	}

	var err error
	var answer string

	_, err = p.setTty("-icanon -isig -echo")
	if err != nil {
		p.Output.Writeln(fmt.Sprintf("<comment>%s</comment>", err.Error()), 0)
		return p.Fallback(), err
	}

	p.cursor.Hide()
	p.Render(renderer())

	scanner := bufio.NewScanner(p.Input.Stream)
	for scanner.Scan() {
		buffer := scanner.Bytes()
		key := string(buffer)

		if key == "" {
			break
		}

		shouldContinue := p.handleKeyPress(key)
		p.Render(renderer())

		if !shouldContinue || key == keys.CtrlC {
			if key == keys.CtrlC {
				if p.CancelUsingFn != nil {
					p.CancelUsingFn()
					break
				} else {
					os.Exit(1)
					break
				}
			}

			if key == keys.CtrlU && p.RevertUsingFn != nil {
				return answer, errors.New("form reverted")
			}

			answer = p.Value()
			break
		}
	}

	if err = scanner.Err(); err != nil {
		return answer, err
	}

	return answer, nil
}

func (p *Prompt) Value() string {
	return ""
}

func (p *Prompt) Render(frame string) {
	if frame == p.prevFrame {
		return
	}

	if p.State == PromptStateInitial {
		p.Output.Write(frame, false, 0)
		p.State = PromptStateActive
		p.prevFrame = frame
		return
	}

	terminalHeight, _ := TerminalHeight()
	previousFrameHeight := len(strings.Split(p.prevFrame, "\n"))
	start := int(math.Abs(float64(min(0, terminalHeight-previousFrameHeight))))
	renderableLines := strings.Split(frame, "\n")[start:]

	p.cursor.MoveToColumn(1)
	up := min(terminalHeight, previousFrameHeight) - 1
	p.cursor.MoveUp(up)
	p.eraseDown()
	p.Output.Write(strings.Join(renderableLines, "\n"), false, 0)

	p.prevFrame = frame
}

func (p *Prompt) ShouldFallback() bool {
	return p.shouldFallback
}

func (p *Prompt) Fallback() string {
	return ""
}

func (p *Prompt) validate(value string) {
	p.validated = true

	if p.Required && strings.TrimSpace(value) == "" {
		p.State = PromptStateError
		p.Error = "Required."
		return
	}

	if p.Validator == nil && p.ValidateUsingFn == nil {
		return
	}

	var err string
	if p.Validator != nil {
		err = p.Validator(value)
	} else if p.ValidateUsingFn != nil {
		err = p.ValidateUsingFn(p)
	}

	p.State = PromptStateError
	if err != "" {
		p.Error = err
	}
}

func (p *Prompt) defaultValue() (string, error) {
	p.validate(p.DefaultValue)

	if p.State == PromptStateError {
		return "", errors.New(p.Error)
	}

	return p.DefaultValue, nil
}

var initialSttyMode string

func (p *Prompt) setTty(mode string) (string, error) {
	if initialSttyMode == "" {
		c := exec.Command("stty", "-g")
		c.Stdin = p.Input.Stream

		out, err := c.Output()
		if err != nil {
			return "", err
		}

		initialSttyMode = string(out)
	}

	c := exec.Command("stty", StringToInputArgs(mode)...)
	c.Stdin = p.Input.Stream

	out, err := c.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (p *Prompt) restoreTty() {
	if initialSttyMode != "" {
		c := exec.Command("stty", StringToInputArgs(initialSttyMode)...)

		err := c.Run()
		if err != nil {
			if p.Output.IsDebug() {
				panic(err)
			}
		}

		initialSttyMode = ""
	}
}

func (p *Prompt) emit(event string, key string) {
	listeners, ok := p.listeners[event]
	if ok {
		for _, listener := range listeners {
			listener(key)
		}
	}
}

func (p *Prompt) Restore() {
	p.cursor.Show()
	p.restoreTty()
}

func (p *Prompt) handleKeyPress(key string) bool {
	if p.State == PromptStateError {
		p.State = PromptStateActive
	}

	p.emit("key", key)

	if p.State == PromptStateSubmit {
		return false
	}

	if key == keys.CtrlU {
		if p.allowValueClearance {
			p.typedValue = ""
			return true
		}

		if p.RevertUsingFn == nil {
			p.State = PromptStateError
			p.Error = "This cannot be reverted."
			return true
		}

		p.State = PromptStateCancel
		p.CancelMessage = "Reverted."

		p.RevertUsingFn()
		return false
	}

	if key == keys.CtrlC {
		p.State = PromptStateCancel
		return false
	}

	if p.validated {
		p.validate(p.Value())
	}

	return true
}

func (p *Prompt) eraseDown() {
	p.writeDirectly("\x1b[J")
}

func (p *Prompt) writeDirectly(s string) {
	p.Output.Write(s, false, 0)
}

// TODO: implement
// func (p *Prompt) trackTypedValue(defaultValue string, submit bool, ignore func(string) bool, allowNewLine bool) {
// 	//
// }
