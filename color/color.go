package color

import (
	"fmt"
	"strconv"
	"strings"
)

type AvailableOption struct {
	set   int
	unset int
}

type Color struct {
	foreground string
	background string
	options    map[string]AvailableOption
}

var COLORS = map[string]int{
	"black":   0,
	"red":     1,
	"green":   2,
	"yellow":  3,
	"blue":    4,
	"magenta": 5,
	"cyan":    6,
	"white":   7,
	"default": 9,
}

var BRIGHT_COLORS = map[string]int{
	"gray":           0,
	"bright-red":     1,
	"bright-green":   2,
	"bright-yellow":  3,
	"bright-blue":    4,
	"bright-magenta": 5,
	"bright-cyan":    6,
	"bright-white":   7,
}

var AVAILABLE_OPTIONS = map[string]AvailableOption{
	"bold":       {set: 1, unset: 22},
	"underscore": {set: 4, unset: 24},
	"blink":      {set: 5, unset: 25},
	"reverse":    {set: 7, unset: 27},
	"conceal":    {set: 8, unset: 28},
}

func NewColor(foreground string, background string, options []string) *Color {
	opts := make(map[string]AvailableOption)

	for _, opt := range options {
		_, exists := AVAILABLE_OPTIONS[opt]
		if exists {
			opts[opt] = AVAILABLE_OPTIONS[opt]
		}
	}

	fg, _ := parseColor(foreground, false)
	bg, _ := parseColor(background, true)

	return &Color{
		foreground: fg,
		background: bg,
		options:    opts,
	}
}

func (c *Color) Apply(text string) string {
	return c.Set() + text + c.Unset()
}

func (c *Color) Set() string {
	setCodes := make([]string, 0)

	if c.foreground != "" {
		setCodes = append(setCodes, c.foreground)
	}

	if c.background != "" {
		setCodes = append(setCodes, c.background)
	}

	for _, opt := range c.options {
		setCodes = append(setCodes, strconv.Itoa(opt.set))
	}

	if len(setCodes) == 0 {
		return ""
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(setCodes, ";"))
}

func (c *Color) Unset() string {
	unsetCodes := make([]string, 0)

	if c.foreground != "" {
		unsetCodes = append(unsetCodes, "39")
	}

	if c.background != "" {
		unsetCodes = append(unsetCodes, "49")
	}

	for _, opt := range c.options {
		unsetCodes = append(unsetCodes, strconv.Itoa(opt.unset))
	}

	if len(unsetCodes) == 0 {
		return ""
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(unsetCodes, ";"))
}

func parseColor(color string, background bool) (string, error) {
	if color == "" {
		return "", nil
	}

	if color[0] == '#' {
		var out string
		if background {
			out += "4"
		} else {
			out += "3"
		}

		converted, err := ConvertFromHexToAnsiColorCode(ColorMode(), color)
		if err != nil {
			return "", err
		}

		out += converted

		return out, nil
	}

	if code, contains := COLORS[color]; contains {
		if background {
			return "4" + strconv.Itoa(code), nil
		}

		return "3" + strconv.Itoa(code), nil
	}

	if code, contains := BRIGHT_COLORS[color]; contains {
		if background {
			return "10" + strconv.Itoa(code), nil
		}

		return "9" + strconv.Itoa(code), nil
	}

	opts := make([]string, 0)
	for key := range COLORS {
		opts = append(opts, key)
	}
	for key := range BRIGHT_COLORS {
		opts = append(opts, key)
	}

	optsString := strings.Join(opts, ", ")
	return "", fmt.Errorf("invalid \"%s\" color; expected one of (%s)", color, optsString)
}
