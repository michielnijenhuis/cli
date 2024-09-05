package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

type Renderer struct {
	output   string
	minWidth int
}

func NewRenderer() *Renderer {
	return &Renderer{
		minWidth: 60,
	}
}

func (r *Renderer) Line(message string, newLine bool) {
	r.output += message
	if newLine {
		r.output += "\n"
	}
}

func (r *Renderer) NewLine(count int) {
	for count > 0 {
		r.output += "\n"
		count--
	}
}

func (r *Renderer) Warning(message string) {
	r.Line(fmt.Sprintf("<fg=yellow>  ⚠ %s</>", message), true)
}

func (r *Renderer) Err(message string) {
	r.Line(fmt.Sprintf("<fg=red>  ⚠ %s</>", message), true)
}

func (r *Renderer) Hint(message string) {
	if strings.TrimSpace(message) == "" {
		return
	}

	terminalWidth, _ := TerminalWidth()
	message = helper.TruncateStart(message, terminalWidth-6)

	r.Line(fmt.Sprintf("  <fg=gray>%s</>", message), true)
}

func (r *Renderer) When(condition bool, cb func(*Renderer), defaultCb func(*Renderer)) {
	if condition {
		cb(r)
	} else {
		defaultCb(r)
	}
}

func (r *Renderer) ToString(state uint) string {
	s := r.output

	if state == PromptStateSubmit || state == PromptStateCancel {
		s += "\n"
	}

	return s
}
