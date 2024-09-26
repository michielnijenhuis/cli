package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/helper/keys"
	"github.com/michielnijenhuis/cli/terminal"
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
	*View
	State               uint
	Error               string
	CancelMessage       string
	Validator           func(string) string
	GetValue            func() string
	input               *Input
	output              *Output
	activeTheme         string
	validated           bool
	cursor              Cursor
	listeners           map[string][]Listener
	allowValueClearance bool
	CancelUsingFn       AnyFunc
	ValidateUsingFn     func(any) string
	RevertUsingFn       AnyFunc
	isChild             bool

	// Value tracking
	AllowNewLine   bool
	Submit         bool
	Ignore         func(key string) bool
	submitted      bool
	typedValue     string
	cursorPosition int

	// Scrolling
	Highlighted  int
	Scroll       int
	FirstVisible int
	Required     bool
}

func NewPrompt(i *Input, o *Output) *Prompt {
	cursor := Cursor{
		Input:  i.Stream,
		Output: o,
	}

	return &Prompt{
		View:          NewView(o),
		input:         i,
		output:        o,
		State:         PromptStateInitial,
		CancelMessage: "Cancelled.",
		activeTheme:   "default",
		cursor:        cursor,
		listeners:     make(map[string][]Listener),
	}
}

func (p *Prompt) doPrompt(renderer func() string) (string, error) {
	if !p.input.IsInteractive() {
		return p.defaultValue()
	}

	var err error
	var answer string

	if !p.isChild {
		_, err = p.input.SetTty("-icanon -isig -echo")
		if err != nil {
			p.output.Writeln(fmt.Sprintf("<comment>%s</comment>", err.Error()), 0)
			return "", err
		}

		p.cursor.Hide()

		defer p.Restore(false)
	}

	p.render(renderer)

	stream := p.input.Stream
	buffer := make([]byte, 3)

	for {
		read, err := stream.Read(buffer)
		if err != nil {
			return "", err
		}

		key := string(buffer[:read])
		if key == "" {
			break
		}

		shouldContinue := p.handleKeyPress(key)
		p.render(renderer)

		if !shouldContinue || keys.Is(key, keys.CtrlC) {
			if keys.Is(key, keys.CtrlC) {
				if p.CancelUsingFn != nil {
					p.CancelUsingFn()
					break
				} else {
					p.Restore(true)
					os.Exit(1)
					break
				}
			}

			if keys.Is(key, keys.CtrlU) && p.RevertUsingFn != nil {
				return answer, errors.New("form reverted")
			}

			answer = p.value()
			break
		}
	}

	return answer, nil
}

func (p *Prompt) Restore(force bool) {
	if !p.isChild || force {
		p.input.RestoreTty()
		p.cursor.Show()
	}
}

func (p *Prompt) render(renderer func() string) {
	p.Render(renderer())

	if p.State == PromptStateInitial {
		p.State = PromptStateActive
	}
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
	p.validate(p.value())

	if p.State == PromptStateError {
		return "", errors.New(p.Error)
	}

	return p.value(), nil
}

func (p *Prompt) emit(event string, key string) {
	if p.listeners == nil {
		return
	}

	for _, listener := range p.listeners[event] {
		listener(key)
	}
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
	if p.GetValue != nil {
		return p.GetValue()
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
	p.SetValue(defaultValue)
	p.Submit = submit
	p.Ignore = ignore
	p.AllowNewLine = allowNewLine

	p.on("key", func(key string) {
		p.Track(key)
	})
}

func (p *Prompt) TypedValue() string {
	return p.typedValue
}

func (p *Prompt) SetValue(value string) {
	p.typedValue = value
	p.cursorPosition = len(value)
}

func (p *Prompt) Track(key string) {
	if keys.Is(key, "\x1b") || keys.Is(key, keys.CtrlB, keys.CtrlF, keys.CtrlA, keys.CtrlE) {
		if p.ignoreKey(key) {
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

	if p.ignoreKey(key) {
		return
	}

	if keys.Is(key, keys.Enter) {
		if p.Submit {
			p.submit()
			return
		}

		if p.AllowNewLine {
			p.typedValue = p.typedValue[:p.cursorPosition] + Eol + p.typedValue[p.cursorPosition:]
			p.cursorPosition++
		}
	} else if keys.Is(key, keys.Backspace) || keys.Is(key, keys.CtrlH) {
		if p.cursorPosition == 0 {
			return
		}

		p.typedValue = p.typedValue[:p.cursorPosition-1] + p.typedValue[p.cursorPosition:]
		p.cursorPosition--
	} else if k := string(key[0]); key[0] >= 32 {
		p.typedValue = p.typedValue[:p.cursorPosition] + k + p.typedValue[p.cursorPosition:]
		p.cursorPosition++
	}
}

func (p *Prompt) CursorPosition() int {
	return p.cursorPosition
}

func (p *Prompt) Submitted() bool {
	return p.submitted
}

func (p *Prompt) ignoreKey(key string) bool {
	if p.Ignore != nil {
		return p.Ignore(key)
	}

	return false
}

func (p *Prompt) Reset() {
	p.submitted = false
	p.typedValue = ""
	p.cursorPosition = 0
}

func (p *Prompt) Resetvalue() {
	p.typedValue = ""
	p.cursorPosition = 0
}

func (p *Prompt) AddCursor(value string, cursorPosition int, maxWidth int) string {
	if maxWidth <= 0 {
		tw := terminal.Columns()
		maxWidth = tw
	}

	before := ""
	current := ""
	after := ""

	if helper.Len(value) >= cursorPosition {
		before = value[0:cursorPosition]

		if helper.Len(value) >= cursorPosition+1 {
			current = value[cursorPosition : cursorPosition+1]
		}

		if helper.Len(value) >= cursorPosition+2 {
			after = value[cursorPosition+1:]
		}
	}

	cursor := " "
	if helper.Len(current) > 0 && current != Eol {
		cursor = current
	}

	var spaceBefore int
	if maxWidth <= 0 {
		spaceBefore = helper.Len(before)
	} else {
		spaceBefore = maxWidth - helper.Len(cursor)

		if helper.Len(after) > 0 {
			spaceBefore--
		}
	}

	truncatedBefore := before
	wasTruncatedBefore := false
	if helper.Len(before) > spaceBefore {
		truncatedBefore = TrimWidthBackwards(before, 0, spaceBefore-1)
		wasTruncatedBefore = true
	}

	var spaceAfter int
	if maxWidth <= 0 {
		spaceAfter = helper.Len(after)
	} else {
		spaceAfter = maxWidth - helper.Len(truncatedBefore) - helper.Len(cursor)

		if wasTruncatedBefore {
			spaceAfter--
		}
	}

	truncatedAfter := after
	wasTruncatedAfter := false
	if helper.Len(after) > spaceAfter {
		truncatedAfter = StrimWidth(after, 0, spaceAfter-1, "")
		wasTruncatedBefore = true
	}

	var out strings.Builder
	if wasTruncatedBefore {
		out.WriteString(Dim("…"))
	}
	out.WriteString(truncatedBefore)
	out.WriteString(Inverse(cursor))
	if current == Eol {
		out.WriteString(Eol)
	}
	out.WriteString(truncatedAfter)
	if wasTruncatedAfter {
		out.WriteString(Dim("…"))
	}

	return out.String()
}

func (p *Prompt) InitializeScrolling(highlighted int, reservedLines int) {
	p.Highlighted = highlighted

	p.ReduceScrollingToFitTerminal(reservedLines)
}

func (p *Prompt) ReduceScrollingToFitTerminal(reservedLines int) {
	terminalHeight := terminal.Lines()
	p.Scroll = max(1, min(p.Scroll, terminalHeight-reservedLines))
}

func (p *Prompt) Highlight(index int) {
	p.Highlighted = index

	if p.Highlighted < 0 {
		return
	}

	if p.Highlighted < p.FirstVisible {
		p.FirstVisible = p.Highlighted
	} else if p.Highlighted >= p.FirstVisible+p.Scroll {
		p.FirstVisible = p.Highlighted - p.Scroll + 1
	}
}

func (p *Prompt) HighlightPrevious(total int) {
	if total <= 0 {
		return
	}

	if p.Highlighted < 0 {
		p.Highlight(total - 1)
	} else if p.Highlighted == 0 {
		if !p.Required {
			p.Highlight(-1)
		} else {
			p.Highlight(total - 1)
		}
	} else {
		p.Highlight(p.Highlighted - 1)
	}
}

func (p *Prompt) HighlightNext(total int) {
	if total <= 0 {
		return
	}

	if p.Highlighted == total-1 {
		if !p.Required {
			p.Highlight(-1)
		} else {
			p.Highlight(0)
		}
	} else {
		if p.Highlighted < 0 {
			p.Highlighted = -1
		}

		p.Highlight(p.Highlighted + 1)
	}
}

func (p *Prompt) ScrollToHighlighted(total int) {
	if p.Highlighted < p.Scroll {
		return
	}

	remaining := total - p.Highlighted - 1
	halfScroll := p.Scroll / 2
	endOffset := max(0, halfScroll-remaining)

	if p.Scroll%2 == 0 {
		endOffset--
	}

	p.FirstVisible = max(0, p.Highlighted-halfScroll-endOffset)
}
