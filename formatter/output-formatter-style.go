package formatter

import (
	"fmt"

	"github.com/michielnijenhuis/cli/color"
)

type OutputFormatterStyle struct {
	foreground            string
	background            string
	options               []string
	color                 *color.Color
	href                  string
	handlesHrefGracefully bool
}

func NewOutputFormatterStyle(foreground string, background string, options []string) *OutputFormatterStyle {
	if options == nil {
		options = []string{}
	}

	return &OutputFormatterStyle{
		foreground:            foreground,
		background:            background,
		options:               options,
		color:                 color.NewColor(foreground, background, options),
		href:                  "",
		handlesHrefGracefully: false,
	}
}

func (s *OutputFormatterStyle) Clone() OutputFormatterStyleInterface {
	options := make([]string, 0, len(s.options))
	copy(options, s.options)

	return NewOutputFormatterStyle(s.foreground, s.background, options)
}

func (s *OutputFormatterStyle) SetForeground(c string) {
	s.foreground = c
	s.color = color.NewColor(c, s.background, s.options)
}

func (s *OutputFormatterStyle) SetBackground(c string) {
	s.background = c
	s.color = color.NewColor(s.foreground, c, s.options)
}

func (s *OutputFormatterStyle) SetHref(href string) {
	s.href = href
}

func (s *OutputFormatterStyle) SetOption(option string) {
	s.options = append(s.options, option)
	s.color = color.NewColor(s.foreground, s.background, s.options)
}

func (s *OutputFormatterStyle) UnsetOption(option string) {
	i := -1
	for j := 0; j < len(s.options); j++ {
		if s.options[j] == option {
			i = j
			break
		}
	}

	if i >= 0 {
		if len(s.options) > 1 {
			s.options[i] = s.options[len(s.options)-1]
		} else {
			s.options = make([]string, 0)
		}
	}

	s.color = color.NewColor(s.foreground, s.background, s.options)
}

func (s *OutputFormatterStyle) SetOptions(options []string) {
	s.options = options
	s.color = color.NewColor(s.foreground, s.background, options)
}

func (s *OutputFormatterStyle) Apply(text string) string {
	if s.href != "" && s.handlesHrefGracefully {
		text = fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", s.href, text)
	}

	return s.color.Apply(text)
}
