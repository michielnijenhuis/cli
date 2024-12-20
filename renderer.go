package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/terminal"
)

type Renderer struct {
	output   strings.Builder
	minWidth int
}

func NewRenderer() *Renderer {
	return &Renderer{
		minWidth: 60,
	}
}

func (r *Renderer) Line(message string, newLine bool) {
	r.output.WriteString(message)
	if newLine {
		r.output.WriteString(Eol)
	}
}

func (r *Renderer) NewLine(count int) {
	for count > 0 {
		r.output.WriteString(Eol)
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

	terminalWidth := terminal.Columns()
	message = TruncateStart(message, terminalWidth-6)

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
	if state == PromptStateSubmit || state == PromptStateCancel {
		r.output.WriteString(Eol)
	}

	return r.output.String()
}
