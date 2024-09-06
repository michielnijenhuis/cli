package cli

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
)

type OutputTheme map[string]*OutputFormatterStyle

type Theme struct {
	Foreground   string
	Background   string
	Options      []string
	Icon         string
	Label        string
	FullyColored bool
	Padding      bool
	style        *OutputFormatterStyle
}

const (
	ThemeBlock = "block"
	ThemeIcon  = "icon"
)

var currentTheme string = ThemeIcon

var themes = map[string]map[string]*Theme{
	"_default": {
		"error": {
			Foreground:   "red",
			FullyColored: true,
		},
		"info": {
			Foreground:   "bright-blue",
			FullyColored: true,
		},
		"success": {
			Foreground:   "bright-green",
			FullyColored: true,
		},
		"ok": {
			Foreground:   "bright-green",
			FullyColored: true,
		},
		"warn": {
			Foreground:   "yellow",
			FullyColored: true,
		},
		"warning": {
			Foreground:   "yellow",
			FullyColored: true,
		},
		"caution": {
			Foreground:   "yellow",
			FullyColored: true,
		},
		"comment": {
			Foreground:   "default",
			Background:   "default",
			Label:        " // ",
			FullyColored: true,
		},
		"note": {
			Foreground:   "yellow",
			FullyColored: true,
		},
		"primary": {
			Foreground:   "bright-magenta",
			FullyColored: true,
		},
		"accent": {
			Foreground:   "bright-cyan",
			FullyColored: true,
		},
		"prompt": {
			Foreground:   "cyan",
			FullyColored: true,
		},
		"question": {
			Foreground:   "cyan",
			FullyColored: true,
		},
	},
	ThemeBlock: {
		"error": {
			Foreground:   "white",
			Background:   "red",
			Label:        " [ERROR] ",
			Padding:      true,
			FullyColored: true,
		},
		"info": {
			Foreground:   "white",
			Background:   "blue",
			Label:        " [INFO] ",
			Padding:      true,
			FullyColored: true,
		},
		"success": {
			Foreground:   "black",
			Background:   "green",
			Label:        " [OK] ",
			Padding:      true,
			FullyColored: true,
		},
		"ok": {
			Foreground:   "black",
			Background:   "green",
			Label:        " [OK] ",
			Padding:      true,
			FullyColored: true,
		},
		"warn": {
			Foreground:   "black",
			Background:   "yellow",
			Label:        " [WARNING] ",
			Padding:      true,
			FullyColored: true,
		},
		"warning": {
			Foreground:   "black",
			Background:   "yellow",
			Label:        " [WARNING] ",
			Padding:      true,
			FullyColored: true,
		},
		"caution": {
			Foreground:   "black",
			Background:   "yellow",
			Label:        " [CAUTION] ",
			Padding:      true,
			FullyColored: true,
		},
		"comment": {
			Foreground: "default",
			Background: "default",
			Label:      " // ",
		},
		"note": {
			Foreground:   "yellow",
			Label:        " [NOTE] ",
			Padding:      true,
			FullyColored: true,
		},
		"primary": {
			Foreground: "bright-magenta",
		},
		"accent": {
			Foreground: "bright-cyan",
		},
		"prompt": {
			Foreground: "cyan",
		},
		"question": {
			Foreground: "cyan",
		},
	},
	ThemeIcon: {
		"error": {
			Foreground:   "bright-red",
			Label:        "Error: ",
			Icon:         IconWarning,
			FullyColored: true,
		},
		"info": {
			Foreground:   "bright-blue",
			Icon:         IconInfo,
			FullyColored: true,
		},
		"success": {
			Foreground:   "bright-green",
			Icon:         IconTick,
			FullyColored: true,
		},
		"ok": {
			Foreground:   "bright-green",
			Icon:         IconTick,
			FullyColored: true,
		},
		"warn": {
			Foreground:   "bright-yellow",
			Icon:         IconWarning,
			Label:        "Warning: ",
			FullyColored: true,
		},
		"warning": {
			Foreground:   "bright-yellow",
			Icon:         IconWarning,
			Label:        "Warning: ",
			FullyColored: true,
		},
		"caution": {
			Foreground:   "bright-yellow",
			Icon:         IconWarning,
			Label:        "Caution: ",
			FullyColored: true,
		},
		"comment": {
			Foreground: "default",
			Background: "default",
			Label:      " // ",
		},
		"note": {
			Foreground:   "yellow",
			Icon:         IconWarning,
			Label:        "Note: ",
			FullyColored: false,
		},
		"primary": {
			Foreground: "bright-magenta",
		},
		"accent": {
			Foreground: "bright-cyan",
		},
		"prompt": {
			Foreground: "bright-cyan",
		},
		"question": {
			Foreground: "bright-cyan",
		},
	},
}

var styleTags []string

func GetStyleTags() []string {
	if styleTags == nil {
		themeSet, ok := themes[currentTheme]
		if ok {
			styleTags = array.Keys(themeSet)
		} else {
			styleTags = array.Keys(themes["_default"])
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
			currentTheme = "_default"
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

func SetBaseTheme(primary string, accent string) {
	AddTheme(currentTheme, "primary", &Theme{
		Foreground: primary,
	})

	AddTheme(currentTheme, "accent", &Theme{
		Foreground: accent,
	})
}

func GetTheme(tag string) (*Theme, error) {
	tag = strings.ToLower(tag)

	if currentTheme == "" {
		currentTheme = "_default"
	}

	themeSet, ok := themes[currentTheme]

	var errUnknownCurrentTheme error
	if !ok {
		themeSet = themes["_default"]
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
