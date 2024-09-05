package cli

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/helper/keys"
)

const (
	PromptStateInitial = iota
	PromptStateActive
	PromptStateSubmit
	PromptStateCancel
	PromptStateError
	PromptStateSearching
	PromptStateDeleting
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
	Value               func() string
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
	// fallback            any // TODO: fix type
	cursorPosition int
	isChild        bool
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

func (p *Prompt) doPrompt(renderer func() string) (string, error) {
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

	if !p.isChild {
		p.cursor.Hide()
	}

	p.render(renderer())

	stream := p.Input.Stream
	buffer := make([]byte, 3)

	for {
		// reset buffer
		for i := 0; i < len(buffer); i++ {
			buffer[i] = 0
		}

		_, err := stream.Read(buffer)
		if err != nil {
			p.cursor.Show()
			return "", err
		}

		key := string(buffer)
		if key == "" {
			break
		}

		shouldContinue := p.handleKeyPress(key)
		p.render(renderer())

		if !shouldContinue || keys.Is(key, keys.CtrlC) {
			if keys.Is(key, keys.CtrlC) {
				if p.CancelUsingFn != nil {
					p.CancelUsingFn()
					break
				} else {
					p.cursor.Show()
					os.Exit(1)
					break
				}
			}

			if keys.Is(key, keys.CtrlU) && p.RevertUsingFn != nil {
				p.cursor.Show()
				return answer, errors.New("form reverted")
			}

			answer = p.value()
			break
		}
	}

	p.cursor.Show()
	return answer, nil
}

func (p *Prompt) writeFrame(frame string) {
	p.Output.Write(frame, false, 0)
	p.prevFrame = frame
}

func (p *Prompt) render(frame string) {
	if frame == p.prevFrame {
		return
	}

	if p.State == PromptStateInitial {
		p.writeFrame(frame)
		p.State = PromptStateActive
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

	if err != "" {
		p.State = PromptStateError
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
	if p.listeners == nil {
		return
	}

	for _, listener := range p.listeners[event] {
		listener(key)
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

	if keys.Is(key, keys.CtrlU) {
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

	if keys.Is(key, keys.CtrlC) {
		p.State = PromptStateCancel
		return false
	}

	if p.validated {
		p.validate(p.value())
	}

	return true
}

func (p *Prompt) eraseDown() {
	p.writeDirectly("\x1b[J")
}

func (p *Prompt) writeDirectly(s string) {
	p.Output.Write(s, false, 0)
}

func (p *Prompt) on(event string, fn func(key string)) {
	if p.listeners == nil {
		p.listeners = make(map[string][]Listener)
	}

	if p.listeners[event] == nil {
		p.listeners[event] = make([]Listener, 0)
	}

	p.listeners[event] = append(p.listeners[event], fn)
}

func (p *Prompt) ClearListeners() {
	p.listeners = nil
}

func (p *Prompt) value() string {
	if p.Value != nil {
		return p.Value()
	}

	return ""
}

func (p *Prompt) submit() {
	p.validate(p.value())

	if p.State != PromptStateError {
		p.State = PromptStateSubmit
	}
}

func (p *Prompt) trackTypedValue(defaultValue string, submit bool, ignore func(key string) bool, allowNewLine bool) {
	p.typedValue = defaultValue

	if p.typedValue != "" {
		p.cursorPosition = len(p.typedValue)
	}

	p.on("key", func(key string) {
		if string(key[0]) == "\x1b" || keys.Is(key, keys.CtrlB, keys.CtrlF, keys.CtrlA, keys.CtrlE) {
			if ignore != nil && ignore(key) {
				return
			}

			switch {
			case keys.Is(key, keys.Left, keys.LeftArrow, keys.CtrlB):
				p.cursorPosition = max(0, p.cursorPosition-1)
			case keys.Is(key, keys.Right, keys.RightArrow, keys.CtrlF):
				p.cursorPosition = min(len(p.typedValue), p.cursorPosition+1)
			case keys.Is(key, keys.CtrlA, keys.Home...):
				p.cursorPosition = 0
			case keys.Is(key, keys.CtrlE, keys.End...):
				p.cursorPosition = len(p.typedValue)
			case keys.Is(key, keys.Delete):
				p.typedValue = p.typedValue[:p.cursorPosition] + p.typedValue[p.cursorPosition+1:]
			default:
			}

			return
		}

		for _, k := range strings.Split(key, "") {
			if ignore != nil && ignore(k) {
				return
			}

			if keys.Is(k, keys.Enter) {
				if submit {
					p.submit()
					return
				}

				if allowNewLine {
					p.typedValue = p.typedValue[:p.cursorPosition] + "\n" + p.typedValue[p.cursorPosition:]
					p.cursorPosition++
				}
			} else if keys.Is(k, keys.Backspace) || keys.Is(k, keys.CtrlH) {
				if p.cursorPosition == 0 {
					return
				}

				p.typedValue = p.typedValue[:p.cursorPosition-1] + p.typedValue[p.cursorPosition:]
				p.cursorPosition--
			} else if k[0] >= 32 {
				p.typedValue = p.typedValue[:p.cursorPosition] + k + p.typedValue[p.cursorPosition:]
				p.cursorPosition++
			}
		}
	})
	//
}

func (p *Prompt) addCursor(value string, cursorPosition int, maxWidth int) string {
	before := ""
	current := ""
	after := ""

	if len(value) >= cursorPosition {
		before = value[0:cursorPosition]

		if len(value) >= cursorPosition+1 {
			current = value[cursorPosition : cursorPosition+1]
		}

		if len(value) >= cursorPosition+2 {
			after = value[cursorPosition+1:]
		}
	}

	cursor := " "
	if len(current) > 0 && current != "\n" {
		cursor = current
	}

	var spaceBefore int
	if maxWidth <= 0 {
		spaceBefore = len(before)
	} else {
		spaceBefore = maxWidth - len(cursor)

		if len(after) > 0 {
			spaceBefore--
		}
	}

	truncatedBefore := before
	wasTruncatedBefore := false
	if len(before) > spaceBefore {
		truncatedBefore = helper.TrimWidthBackwards(before, 0, spaceBefore-1)
		wasTruncatedBefore = true
	}

	var spaceAfter int
	if maxWidth <= 0 {
		spaceAfter = len(after)
	} else {
		spaceAfter = maxWidth - len(truncatedBefore) - len(cursor)

		if wasTruncatedBefore {
			spaceAfter--
		}
	}

	truncatedAfter := after
	wasTruncatedAfter := false
	if len(after) > spaceAfter {
		truncatedAfter = helper.StrimWidth(after, 0, spaceAfter-1, "")
		wasTruncatedBefore = true
	}

	var out strings.Builder
	if wasTruncatedBefore {
		out.WriteString(Dim("…"))
	}
	out.WriteString(truncatedBefore)
	out.WriteString(Inverse(cursor))
	if current == "\n" {
		out.WriteString("\n")
	}
	out.WriteString(truncatedAfter)
	if wasTruncatedAfter {
		out.WriteString(Dim("…"))
	}

	return out.String()
}

func (p *Prompt) initializeScrolling(highlighted int) {
	p.Highlighted = highlighted

	p.reduceScrollingToFitTerminal()
}

func (p *Prompt) reduceScrollingToFitTerminal() {
	terminalHeight, _ := TerminalHeight()
	p.Scroll = max(1, min(p.Scroll, terminalHeight))
}

func (p *Prompt) highlight(index int) {
	p.Highlighted = index

	if index < 0 {
		return
	}

	if index < p.FirstVisible {
		p.FirstVisible = index
	} else if index < p.FirstVisible+p.Scroll-1 {
		p.FirstVisible = index - p.Scroll + 1
	}
}

func (p *Prompt) highlightPrevious(total int) {
	if total <= 0 {
		return
	}

	if p.Highlighted < 0 {
		p.highlight(total - 1)
	} else if p.Highlighted == 0 {
		if !p.Required {
			p.highlight(-1)
		} else {
			p.highlight(total - 1)
		}
	} else {
		p.highlight(p.Highlighted - 1)
	}
}

func (p *Prompt) highlightNext(total int) {
	if total <= 0 {
		return
	}

	if p.Highlighted == total-1 {
		if !p.Required {
			p.highlight(-1)
		} else {
			p.highlight(0)
		}
	} else {
		if p.Highlighted < -1 {
			p.Highlighted = -1
		}

		p.highlight(p.Highlighted + 1)
	}
}

// TODO: fix
// func (p *Prompt) scrollToHighlighted(total int) {
// 	if p.Highlighted < 0 || p.Highlighted < p.Scroll {
// 		return
// 	}

// 	remaining := total - p.Highlighted - 1
// 	halfScroll := p.Scroll / 2
// 	endOffset := max(0, halfScroll-remaining)

// 	if p.Scroll%2 == 0 {
// 		endOffset--
// 	}

// 	p.FirstVisible = max(0, p.Highlighted-halfScroll-endOffset)
// }
