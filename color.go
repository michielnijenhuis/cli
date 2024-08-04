package cli

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type AvailableOption struct {
	set   int
	unset int
}

type Color struct {
	Foreground string
	Background string
	Options    []string
	options    map[string]AvailableOption
	parsed     bool
}

var colors = map[string]int{
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

var brightColors = map[string]int{
	"gray":           0,
	"bright-red":     1,
	"bright-green":   2,
	"bright-yellow":  3,
	"bright-blue":    4,
	"bright-magenta": 5,
	"bright-cyan":    6,
	"bright-white":   7,
}

var availableOptions = map[string]AvailableOption{
	"bold":       {set: 1, unset: 22},
	"underscore": {set: 4, unset: 24},
	"blink":      {set: 5, unset: 25},
	"reverse":    {set: 7, unset: 27},
	"conceal":    {set: 8, unset: 28},
}

func (c *Color) parse() {
	if c.parsed {
		return
	}

	c.parsed = true

	opts := make(map[string]AvailableOption)
	if c.Options != nil {
		for _, opt := range c.Options {
			_, exists := availableOptions[opt]
			if exists {
				opts[opt] = availableOptions[opt]
			}
		}
	}
	c.options = opts

	fg, _ := parseColor(c.Foreground, false)
	bg, _ := parseColor(c.Background, true)
	c.Foreground = fg
	c.Background = bg
}

func (c *Color) Apply(text string) string {
	c.parse()

	return c.Set() + text + c.Unset()
}

func (c *Color) Set() string {
	c.parse()

	setCodes := make([]string, 0)

	if c.Foreground != "" {
		setCodes = append(setCodes, c.Foreground)
	}

	if c.Background != "" {
		setCodes = append(setCodes, c.Background)
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
	c.parse()

	unsetCodes := make([]string, 0)

	if c.Foreground != "" {
		unsetCodes = append(unsetCodes, "39")
	}

	if c.Background != "" {
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

	if code, contains := colors[color]; contains {
		if background {
			return "4" + strconv.Itoa(code), nil
		}

		return "3" + strconv.Itoa(code), nil
	}

	if code, contains := brightColors[color]; contains {
		if background {
			return "10" + strconv.Itoa(code), nil
		}

		return "9" + strconv.Itoa(code), nil
	}

	opts := make([]string, 0)
	for key := range colors {
		opts = append(opts, key)
	}
	for key := range brightColors {
		opts = append(opts, key)
	}

	optsString := strings.Join(opts, ", ")
	return color, fmt.Errorf("invalid \"%s\" color; expected one of (%s)", color, optsString)
}

const (
	ansi4  uint8 = 4
	ansi8  uint8 = 8
	ansi24 uint8 = 24
)

func ConvertFromHexToAnsiColorCode(mode uint8, hexColor string) (string, error) {
	hexColor = strings.Replace(hexColor, "#", "", 1)

	if len(hexColor) == 3 {
		hexColor = string(hexColor[0] + hexColor[0] + hexColor[1] + hexColor[1] + hexColor[2] + hexColor[2])
	}

	if len(hexColor) != 6 {
		return "", fmt.Errorf("invalid \"#%s\" color", hexColor)
	}

	color, e := strconv.ParseInt(hexColor, 16, 64)
	if e != nil {
		return "", e
	}

	r := (color >> 16) & 255
	g := (color >> 8) & 255
	b := color & 255

	switch mode {
	case ansi4:
		return convertFromRGB(mode, r, g, b)
	case ansi8:
		str, e := convertFromRGB(mode, r, g, b)
		if e != nil {
			return "", e
		}
		return ("8;5;" + str), nil
	case ansi24:
		return ("8;2;" + strings.Join([]string{strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))}, ";")), nil
	default:
		return "", errors.New("invalid Ansi color mode. Options: 4, 8, 24")
	}
}

func convertFromRGB(mode uint8, r int64, g int64, b int64) (string, error) {
	switch mode {
	case ansi4:
		return strconv.Itoa(degradeHexColorToAnsi4(r, g, b)), nil
	case ansi8:
		return strconv.Itoa(degradeHexColorToAnsi8(r, g, b)), nil
	case ansi24:
		return "", errors.New("rgb cannot be converted to Ansi24")
	default:
		return "", errors.New("invalid Ansi color mode. Options: 4, 8, 24")
	}
}

func degradeHexColorToAnsi4(r int64, g int64, b int64) int {
	return (int(math.Round(float64(b/255))) << 2) | (int(math.Round(float64(g/255))) << 1) | int(math.Round(float64(r/255)))
}

func degradeHexColorToAnsi8(r int64, g int64, b int64) int {
	if r == g && g == b {
		if r < 8 {
			return 16
		}

		if r > 248 {
			return 231
		}

		return int(math.Round(float64(((r-8)/247)*24))) + 232
	} else {
		return 16 + 36 + int(math.Round(float64((r/255)*5))) + 6*int(math.Round(float64((g/255)*5))) + int(math.Round(float64((b/255)*5)))
	}
}

var colorMode uint8 = 0

func ColorMode() uint8 {
	if colorMode > 0 {
		return colorMode
	}

	envColorTerm := os.Getenv("COLOR_TERM")
	if envColorTerm != "" {
		envColorTerm = strings.ToLower(envColorTerm)

		if strings.Contains(envColorTerm, "truecolor") {
			colorMode = ansi24
			return colorMode
		}

		if strings.Contains(envColorTerm, "256color") {
			colorMode = ansi8
			return colorMode
		}
	}

	colorMode = ansi4
	return colorMode
}

const falseString = "false"

func HasColorSupport() bool {
	_, envSet := os.LookupEnv("NO_COLOR")
	if envSet {
		return false
	}

	if os.Getenv("TERM_PROGRAM") == "Hyper" ||
		os.Getenv("COLORTERM") != falseString ||
		os.Getenv("ANSICON") != falseString ||
		os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}

	re := regexp.MustCompile("/^((screen|xterm|vt100|vt220|putty|rxvt|ansi|cygwin|linux).*)|(.*-256(color)?(-bce)?)/")
	return re.MatchString(term)
}
