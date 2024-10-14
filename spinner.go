package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Spinner struct {
	*Prompt
	Interval          int
	Count             int
	Message           string
	Frames            []string
	Color             string
	KeepRenderedLines bool
}

type SpinnerFrame struct {
	Frames []string
	i      int
}

func (s *SpinnerFrame) Next() string {
	s.i++
	return s.Frames[s.i%len(s.Frames)]
}

type spinnerRenderer struct {
	*Renderer
	frames   []string
	interval int
	state    uint
}

func (r *spinnerRenderer) Render(p any) {
	s := p.(*Spinner)

	if s.State == PromptStateCancel {
		r.Line(fmt.Sprintf("<fg=yellow>%s</>", s.CancelMessage), true)
		return
	}

	r.state = s.State
	s.Interval = r.interval

	frame := r.frames[s.Count%len(r.frames)]
	line := fmt.Sprintf("<fg=%s>%s</> %s", s.Color, frame, s.Message)

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
	p := NewPrompt(i, o)

	if frames == nil {
		frames = DotSpinner
	}

	if color == "" {
		color = ColorCyan
	}

	s := &Spinner{
		Interval: 100,
		Prompt:   p,
		Message:  message,
		Frames:   frames,
		Color:    color,
	}

	return s
}

func RenderSpinner(s *Spinner) string {
	r := NewSpinnerRenderer(s)
	r.Render(s)
	return r.String()
}

var ErrCancelledSpinner = errors.New("cancelled")

func (s *Spinner) Spin(fn func()) {
	s.cursor.Hide()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)

	go func(c context.Context) {
		for {
			select {
			case <-c.Done():
				return
			default:
				s.Count++
				time.Sleep(time.Duration(s.Interval) * time.Millisecond)
			}
		}
	}(ctx)

	go func(cb func(), d chan<- bool) {
		cb()
		done <- true
	}(fn, done)

	go func(s *Spinner) {
		<-sigs
		s.State = PromptStateCancel
		cancel()
	}(s)

	<-done

	if !s.KeepRenderedLines {
		s.eraseRenderedLines()
	}

	s.cursor.Show()
	cancel()

	if s.State == PromptStateCancel {
		s.Render(RenderSpinner(s))
		s.output.NewLine(1)
		os.Exit(1)
	}
}

func (s *Spinner) eraseRenderedLines() {
	lines := strings.Split(s.prevFrame, Eol)
	s.cursor.Move(-999, (-1*len(lines))+1)
	s.eraseDown()
}
