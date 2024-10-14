package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/michielnijenhuis/cli/terminal"
)

type View struct {
	cursor    Cursor
	output    *Output
	prevFrame string
	init      bool
}

func NewView(o *Output) *View {
	return &View{
		cursor: Cursor{
			Output: o,
		},
		output: o,
	}
}

func (v *View) HideCursor() {
	v.cursor.Hide()
}

func (v *View) ShowCursor() {
	v.cursor.Show()
}

func (v *View) Clear() {
	terminalHeight := terminal.Lines()
	previousFrameHeight := v.prevHeight()
	v.cursor.MoveToColumn(1)
	up := min(terminalHeight, previousFrameHeight) - 1
	v.cursor.MoveUp(up)
	v.eraseDown()
}

func (v *View) prevHeight() int {
	return len(strings.Split(v.prevFrame, Eol))
}

func (v *View) Render(frame string) {
	if frame == v.prevFrame {
		return
	}

	if !v.init {
		v.Write(frame)
		v.init = true
	}

	v.Clear()

	terminalHeight := terminal.Lines()
	previousFrameHeight := v.prevHeight()

	start := int(math.Abs(float64(min(0, terminalHeight-previousFrameHeight))))
	renderableLines := strings.Split(frame, Eol)[start:]
	v.Write(strings.Join(renderableLines, Eol))
}

func (v *View) RenderLine(frame string) {
	if !strings.HasSuffix(frame, Eol) {
		frame += Eol
	}

	v.Render(frame)
}

func (v *View) RenderLinef(format string, args ...any) {
	v.RenderLine(fmt.Sprintf(format, args...))
}

func (v *View) eraseDown() {
	v.writeDirectly("\x1b[J")
}

func (v *View) writeDirectly(s string) {
	v.output.Write(s, false, 0)
}

func (v *View) Write(frame string) {
	v.output.Write(frame, false, 0)
	v.prevFrame = frame
}
