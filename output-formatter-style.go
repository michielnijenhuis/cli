package cli

import (
	"fmt"
)

type OutputFormatterStyle struct {
	foreground            string
	background            string
	options               []string
	color                 *Color
	href                  string
	handlesHrefGracefully bool
}

func NewOutputFormatterStyle(foreground string, background string, options []string) *OutputFormatterStyle {
	if options == nil {
		options = []string{}
	}

	return &OutputFormatterStyle{
		foreground: foreground,
		background: background,
		options:    options,
		color: &Color{
			Foreground: foreground,
			Background: background,
			Options:    options,
		},
		href:                  "",
		handlesHrefGracefully: false,
	}
}

func (s *OutputFormatterStyle) Clone() *OutputFormatterStyle {
	options := make([]string, 0, len(s.options))
	options = append(options, s.options...)

	return NewOutputFormatterStyle(s.foreground, s.background, options)
}

func (s *OutputFormatterStyle) SetForeground(c string) {
	s.foreground = c
	s.color = &Color{
		Foreground: c,
		Background: s.background,
		Options:    s.options,
	}
}

func (s *OutputFormatterStyle) SetBackground(c string) {
	s.background = c
	s.color = &Color{
		Foreground: s.foreground,
		Background: c,
		Options:    s.options,
	}
}

func (s *OutputFormatterStyle) SetHref(href string) {
	s.href = href
}

func (s *OutputFormatterStyle) SetOption(option string) {
	if s.options == nil {
		s.options = []string{option}
	} else {
		s.options = append(s.options, option)
	}

	s.color = &Color{
		Foreground: s.foreground,
		Background: s.background,
		Options:    s.options,
	}
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

	s.color = &Color{
		Foreground: s.foreground,
		Background: s.background,
		Options:    s.options,
	}
}

func (s *OutputFormatterStyle) SetOptions(options []string) {
	s.options = options
	s.color = &Color{
		Foreground: s.foreground,
		Background: s.background,
		Options:    options,
	}
}

func (s *OutputFormatterStyle) Apply(text string) string {
	if s.href != "" && s.handlesHrefGracefully {
		text = fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", s.href, text)
	}

	return s.color.Apply(text)
}
