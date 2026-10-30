package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/michielnijenhuis/cli/internal/helper/array"
)

type OutputTheme map[string]*OutputFormatterStyle

type Theme struct {
	Foreground   string
	Background   string
	Options      []string
	Label        string
	FullyColored bool
	ColorOutput  bool
	Padding      bool
	style        *OutputFormatterStyle
}

var currentTheme string = "default"

var themes = map[string]map[string]*Theme{
	"default": {
		"critical": {
			Background:   "red",
			Foreground:   "bright-white",
			Label:        " CRITICAL ",
			FullyColored: true,
		},
		"error": {
			Background:   "red",
			Foreground:   "bright-white",
			Label:        " ERROR ",
			FullyColored: true,
		},
		"info": {
			Background:   "blue",
			Foreground:   "bright-white",
			Label:        " INFO ",
			FullyColored: true,
		},
		"success": {
			Background:   "green",
			Foreground:   "black",
			Label:        " SUCCESS ",
			FullyColored: true,
		},
		"ok": {
			Background:   "green",
			Foreground:   "black",
			Label:        " OK ",
			FullyColored: true,
		},
		"warn": {
			Background:   "bright-yellow",
			Foreground:   "black",
			Label:        " WARNING ",
			FullyColored: true,
		},
		"warning": {
			Background:   "bright-yellow",
			Foreground:   "black",
			Label:        " WARNING ",
			FullyColored: true,
		},
		"caution": {
			Background:   "bright-yellow",
			Foreground:   "black",
			Label:        " CAUTION ",
			FullyColored: true,
		},
		"comment": {
			Foreground:  "gray",
			Background:  "default",
			Label:       "// ",
			ColorOutput: true,
		},
		"note": {
			Foreground:   "yellow",
			Label:        "# note:",
			FullyColored: false,
		},
		"debug": {
			Foreground: "black",
			Background: "bright-white",
			Label:      " DEBUG ",
		},
		"primary": {
			Foreground: "bright-magenta",
		},
		"secondary": {
			Foreground: "bright-cyan",
		},
		"accent": {
			Foreground: "bright-white",
		},
	},
}

func init() {
	envPrimary := os.Getenv("CLI_THEME_PRIMARY")
	if envPrimary != "" {
		themes["default"]["primary"].Foreground = envPrimary
	}

	envSecondary := os.Getenv("CLI_THEME_SECONDARY")
	if envSecondary != "" {
		themes["default"]["secondary"].Foreground = envSecondary
	}

	envAccent := os.Getenv("CLI_THEME_ACCENT")
	if envAccent != "" {
		themes["default"]["accent"].Foreground = envAccent
	}
}

var styleTags []string

func GetStyleTags() []string {
	if styleTags == nil {
		themeSet, ok := themes[currentTheme]
		if ok {
			styleTags = array.Keys(themeSet)
		} else {
			styleTags = array.Keys(themes["default"])
		}
	}

	return styleTags
}

func AddThemeSet(name string, themeSet map[string]*Theme) {
	name = strings.ToLower(name)
	themes[name] = themeSet
}

func SetCurrentThemeSet(name string) {
	currentTheme = name
}

func AddTheme(set string, tag string, theme *Theme) {
	set = strings.ToLower(set)
	if set == "" {
		if currentTheme == "" {
			currentTheme = "default"
		}

		set = currentTheme
	}

	tag = strings.ToLower(tag)

	themeSet := themes[set]
	if themeSet == nil {
		themeSet = make(map[string]*Theme)
		themes[set] = themeSet
	}

	if _, ok := themeSet[tag]; !ok && styleTags != nil {
		styleTags = append(styleTags, tag)
	}

	themeSet[tag] = theme
}

func SetBaseTheme(primary string, secondary string, accent string) {
	if primary != "" {
		AddTheme(currentTheme, "primary", &Theme{
			Foreground: primary,
		})
	}

	if secondary != "" {
		AddTheme(currentTheme, "secondary", &Theme{
			Foreground: secondary,
		})
	}

	if accent != "" {
		AddTheme(currentTheme, "accent", &Theme{
			Foreground: accent,
		})
	}
}

func GetTheme(tag string) (*Theme, error) {
	tag = strings.ToLower(tag)

	if currentTheme == "" {
		currentTheme = "default"
	}

	themeSet, ok := themes[currentTheme]

	var errUnknownCurrentTheme error
	if !ok {
		themeSet = themes["default"]
		errUnknownCurrentTheme = fmt.Errorf("unknown theme: \"%s\"", currentTheme)
	}

	theme, ok := themeSet[tag]
	if !ok {
		return &Theme{}, fmt.Errorf("unknown tag for theme \"%s\": \"%s\"", currentTheme, tag)
	}

	return theme, errUnknownCurrentTheme
}

func (t *Theme) GetStyle() *OutputFormatterStyle {
	if t.style == nil {
		t.style = NewOutputFormatterStyle(t.Foreground, t.Background, t.Options)
	}

	return t.style
}

func GetGumColorCode(color string) string {
	if code, ok := Colors[color]; ok {
		return fmt.Sprintf("%d", code)
	}

	if code, ok := BrightColors[color]; ok {
		return fmt.Sprintf("%d", code)
	}

	return color
}

func GetGumColorForTheme(tag string) (string, error) {
	theme, err := GetTheme(tag)
	if err != nil {
		return "", err
	}

	return GetGumColorCode(theme.Foreground), nil
}

func ConfigureGumTheme() error {
	primary, err := GetGumColorForTheme("primary")
	if err != nil {
		return err
	}

	secondary, err := GetGumColorForTheme("secondary")
	if err != nil {
		return err
	}

	_, err = GetGumColorForTheme("accent")
	if err != nil {
		return err
	}

	if primary != "" {
		os.Setenv("GUM_CHOOSE_CURSOR_FOREGROUND", primary)
		os.Setenv("GUM_INPUT_CURSOR_FOREGROUND", primary)
		os.Setenv("GUM_CONFIRM_SELECTED_BACKGROUND", primary)
		os.Setenv("GUM_SPIN_SPINNER_FOREGROUND", primary)
	}

	if secondary != "" {
		os.Setenv("GUM_CHOOSE_HEADER_FOREGROUND", secondary)
		os.Setenv("GUM_CONFIRM_PROMPT_FOREGROUND", secondary)
		os.Setenv("GUM_INPUT_HEADER_FOREGROUND", secondary)
	}

	return nil
}
