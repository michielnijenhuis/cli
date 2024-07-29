package color

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	ANSI_4  uint8 = 4
	ANSI_8  uint8 = 8
	ANSI_24 uint8 = 24
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
	case ANSI_4:
		return convertFromRGB(mode, r, g, b)
	case ANSI_8:
		str, e := convertFromRGB(mode, r, g, b)
		if e != nil {
			return "", e
		}
		return ("8;5;" + str), nil
	case ANSI_24:
		return ("8;2;" + strings.Join([]string{strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))}, ";")), nil
	default:
		return "", errors.New("invalid Ansi color mode. Options: 4, 8, 24")
	}
}

func convertFromRGB(mode uint8, r int64, g int64, b int64) (string, error) {
	switch mode {
	case ANSI_4:
		return strconv.Itoa(degradeHexColorToAnsi4(r, g, b)), nil
	case ANSI_8:
		return strconv.Itoa(degradeHexColorToAnsi8(r, g, b)), nil
	case ANSI_24:
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
			colorMode = ANSI_24
			return colorMode
		}

		if strings.Contains(envColorTerm, "256color") {
			colorMode = ANSI_8
			return colorMode
		}
	}

	colorMode = ANSI_4
	return colorMode
}
