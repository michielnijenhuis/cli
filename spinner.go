package cli

import (
	"fmt"
	"strings"
	"time"
)

type Spinner struct {
	*Prompt
	Interval int
	Count    int
	Message  string
	Frames   []string
	Color    string
	spinning bool
}

type spinnerRenderer struct {
	*Renderer
	frames   []string
	interval int
	state    uint
}

func (r *spinnerRenderer) Render(p any) {
	s, ok := p.(*Spinner)
	if !ok {
		panic("expected a spinner prompt")
	}

	r.state = s.State
	s.Interval = r.interval

	frame := r.frames[s.Count%len(r.frames)]
	line := fmt.Sprintf(" <fg=%s>%s</> %s", s.Color, frame, s.Message)

	r.Line(line, true)
}

func (r *spinnerRenderer) String() string {
	return r.ToString(r.state)
}

func NewSpinnerRenderer(s *Spinner) RendererInterface {
	return &spinnerRenderer{
		Renderer: NewRenderer(),
		frames:   s.Frames, // https://www.fileformat.info/info/unicode/block/braille_patterns/images.htm
		interval: 75,
	}
}

func NewSpinner(i *Input, o *Output, message string, frames []string, color string) *Spinner {
	p := NewPrompt("spinner", i, o)

	if frames == nil {
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}

	if color == "" {
		color = "cyan"
	}

	return &Spinner{
		Interval: 100,
		Prompt:   p,
		Message:  message,
		Frames:   frames,
		Color:    color,
	}
}

func RenderSpinner(s *Spinner) string {
	r := NewSpinnerRenderer(s)
	r.Render(s)
	return r.String()
}

func (s *Spinner) Spin(fn func()) {
	s.cursor.Hide()

	defer s.cursor.Show()

	s.Render(RenderSpinner(s))

	s.spinning = true
	go func(cb func()) {
		fn()
		s.spinning = false
	}(fn)

	for s.spinning {
		time.Sleep(time.Duration(s.Interval) * time.Millisecond)
		s.Render(RenderSpinner(s))
		s.Count++
	}

	s.eraseRenderedLines()
}

func (s *Spinner) eraseRenderedLines() {
	lines := strings.Split(s.prevFrame, "\n")
	s.cursor.Move(-999, (-1*len(lines))+1)
	s.eraseDown()
}
