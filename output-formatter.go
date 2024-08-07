package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

type OutputFormatter struct {
	Decorated  bool
	Styles     map[string]*OutputFormatterStyle
	StyleStack *OutputFormatterStyleStack
}

type OutputTheme map[string]*OutputFormatterStyle

type Theme struct {
	Foreground string
	Background string
	Options    []string
}

var DefaultOutputTheme = map[string]*OutputFormatterStyle{
	"error":     NewOutputFormatterStyle("white", "red", nil),
	"info":      NewOutputFormatterStyle("white", "blue", nil),
	"success":   NewOutputFormatterStyle("white", "green", nil),
	"ok":        NewOutputFormatterStyle("white", "green", nil),
	"warn":      NewOutputFormatterStyle("black", "yellow", nil),
	"warning":   NewOutputFormatterStyle("black", "yellow", nil),
	"caution":   NewOutputFormatterStyle("black", "yellow", nil),
	"comment":   NewOutputFormatterStyle("yellow", "", nil),
	"alert":     NewOutputFormatterStyle("red", "", []string{"bold"}),
	"header":    NewOutputFormatterStyle("yellow", "", nil),
	"highlight": NewOutputFormatterStyle("green", "", nil),
	"prompt":    NewOutputFormatterStyle("cyan", "", nil),
	"question":  NewOutputFormatterStyle("black", "cyan", nil),
}

var CustomOutputTheme = map[string]*OutputFormatterStyle{}

func AddTheme(tag string, theme Theme) {
	CustomOutputTheme[tag] = NewOutputFormatterStyle(theme.Foreground, theme.Background, theme.Options)
}

func (o *OutputFormatter) init() {
	if o.Styles == nil {
		o.Styles = make(map[string]*OutputFormatterStyle)
	}

	if o.StyleStack == nil {
		o.StyleStack = &OutputFormatterStyleStack{
			EmptyStyle: makeEmptyStyle(),
			Styles:     make([]*OutputFormatterStyle, 0),
		}
	}
}

func (o *OutputFormatter) SetStyle(name string, style *OutputFormatterStyle) {
	o.init()
	o.Styles[strings.ToLower(name)] = style
}

func (o *OutputFormatter) HasStyle(name string) bool {
	o.init()
	style, _ := o.Style(name)
	return style != nil
}

func (o *OutputFormatter) Style(name string) (*OutputFormatterStyle, error) {
	name = strings.ToLower(name)

	ownStyle, ok := o.Styles[name]
	if ok {
		return ownStyle, nil
	}

	customStyle, ok := CustomOutputTheme[name]
	if ok {
		return customStyle, nil
	}

	defaultStyle, ok := DefaultOutputTheme[name]
	if ok {
		return defaultStyle, nil
	}

	return nil, fmt.Errorf("undefined style: \"%s\"", name)
}

func (o *OutputFormatter) Format(message string) string {
	return o.FormatAndWrap(message, 0)
}

func (o *OutputFormatter) FormatAndWrap(message string, width int) string {
	if message == "" {
		return message
	}

	o.init()

	var offset int
	var output string
	var currentLineLength int

	re := regexp.MustCompile(`<\/?([a-z]+)(?:=([a-zA-Z-]+(?:;[a-zA-Z-]+=[a-zA-Z-]+)*))?>|<\/>`)
	matches := re.FindAllStringSubmatchIndex(message, -1)

	for _, match := range matches {
		pos := match[0]
		text := message[pos:match[1]]

		if pos != 0 && message[pos-1] == '\\' {
			continue
		}

		output += o.applyCurrentStyle(message[offset:pos], output, width, currentLineLength)
		offset = pos + len(text)

		open := text[1] != '/'
		tag := ""
		if open {
			if strings.Contains(text, "=") {
				tag = fmt.Sprintf("%s=%s", message[match[2]:match[3]], message[match[4]:match[5]])
			} else {
				tag = message[match[2]:match[3]]
			}
		} else {
			if match[4] != -1 {
				tag = message[match[4]:match[5]]
			}
		}

		if !open && tag == "" {
			// </>
			o.StyleStack.Pop(nil)
		} else {
			style := o.createStyleFromString(tag)

			if style == nil {
				output += o.applyCurrentStyle(text, output, width, currentLineLength)
			} else if open {
				o.StyleStack.Push(style)
			} else {
				o.StyleStack.Pop(style)
			}
		}
	}

	str := message[offset:]
	output += o.applyCurrentStyle(str, output, width, currentLineLength)

	output = strings.ReplaceAll(output, "\x00", "\\")
	output = strings.ReplaceAll(output, "\\<", "<")
	output = strings.ReplaceAll(output, "\\>", ">")

	return output
}

func (o *OutputFormatter) RemoveDecoration(str string) string {
	isDecorated := o.Decorated
	o.Decorated = false

	str = o.Format(str)

	re1 := regexp.MustCompile(`\033\[[^m]*m`)
	str = re1.ReplaceAllString(str, "")

	re2 := regexp.MustCompile(`\\033]8;[^;]*;[^\\033]*\\033\\\\`)
	str = re2.ReplaceAllString(str, "")

	o.Decorated = isDecorated

	return str
}

func (o *OutputFormatter) Clone() *OutputFormatter {
	clone := &OutputFormatter{
		Decorated:  o.Decorated,
		Styles:     make(map[string]*OutputFormatterStyle),
		StyleStack: o.StyleStack.Clone(),
	}

	for key, value := range o.Styles {
		clone.Styles[key] = value.Clone()
	}

	return clone
}

func Escape(s string) string {
	re := regexp.MustCompile(`([^\\]|^)([<>])`)
	s = re.ReplaceAllString(s, `$1`)

	return EscapeTrailingBackslash(s)
}

func EscapeTrailingBackslash(s string) string {
	if strings.HasSuffix(s, "\\") {
		length := len(s)
		currentLength := length

		for strings.HasSuffix(s, "\\") {
			s = s[:currentLength-1]
			currentLength--
		}

		s = strings.Replace(s, `\0`, "", 1)
		s += strings.Repeat(`\0`, length-len(s))
	}

	return s
}

func (o *OutputFormatter) createStyleFromString(s string) *OutputFormatterStyle {
	if style, err := o.Style(s); err == nil {
		return style
	}

	re := regexp.MustCompile(`([^;=]+)=([a-zA-Z-]+)(;|$)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return nil
	}

	style := NewOutputFormatterStyle("", "", nil)
	for _, match := range matches {
		key := strings.ToLower(match[1])
		value := strings.ToLower(match[2])

		switch key {
		case "fg":
			style.SetForeground(value)
		case "bg":
			style.SetBackground(value)
		case "href":
			url := strings.ReplaceAll(value, `\`, "")
			style.SetHref(url)
		case "options":
			options := strings.Split(value, ",")
			for _, option := range options {
				style.SetOption(option)
			}
		default:
			return nil
		}
	}

	return style
}

func (o *OutputFormatter) applyCurrentStyle(text string, current string, width int, currentLineLength int) string {
	if text == "" {
		return ""
	}

	if width == 0 {
		if o.Decorated {
			return o.StyleStack.Current().Apply(text)
		}
		return text
	}

	if currentLineLength == 0 && current != "" {
		text = strings.TrimLeft(text, " ")
	}

	prefix := ""
	if currentLineLength != 0 {
		prefixLength := width - currentLineLength
		if len(text) > prefixLength {
			prefix = text[:prefixLength] + "\n"
			text = text[prefixLength:]
		}
	}

	text = prefix + o.addLineBreaks(text, width)
	text = strings.TrimRight(text, " ")
	if currentLineLength == 0 && current != "" && !strings.HasSuffix(current, "\n") {
		text = "\n" + text
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		currentLineLength += len(line)
		if width <= currentLineLength {
			currentLineLength = 0
		}
	}

	if o.Decorated {
		for i, line := range lines {
			lines[i] = o.StyleStack.Current().Apply(line)
		}
	}

	return strings.Join(lines, "\n")
}

func (o *OutputFormatter) addLineBreaks(text string, width int) string {
	words := []string{}
	wordStartIndex := 0

	for i := 0; i < len(text); i++ {
		charWidth := helper.Width(string(text[i]))
		if charWidth == 0 || charWidth > 1 {
			if i > wordStartIndex {
				words = append(words, text[wordStartIndex:i])
			}
			words = append(words, string(text[i]))
			wordStartIndex = i + 1
		} else if i == len(text)-1 {
			words = append(words, text[wordStartIndex:])
		}
	}

	result := ""
	lineLength := 0

	for _, word := range words {
		wordLength := helper.Len(word)
		if lineLength+wordLength > width {
			result += "\n" + word
			lineLength = wordLength
		} else {
			if result != "" {
				result += " "
				lineLength++
			}
			result += word
			lineLength += wordLength
		}
	}

	return result
}

func FormatSection(section string, message string, style string) string {
	return fmt.Sprintf("<%s>[%s]</%s> %s", style, section, message, style)
}

func FormatBlock(messages []string, style string, large bool) string {
	var length int
	lines := make([]string, 0, len(messages))

	for _, message := range messages {
		message = Escape(message)
		if large {
			lines = append(lines, fmt.Sprintf("  %s  ", message))
		} else {
			lines = append(lines, fmt.Sprintf(" %s ", message))
		}

		if large {
			length = max(helper.Width(message)+4, length)
		} else {
			length = max(helper.Width(message)+2, length)
		}
	}

	messages = make([]string, 0)
	if large {
		messages = append(messages, strings.Repeat(" ", length))
	}

	for _, line := range lines {
		messages = append(messages, line+strings.Repeat(" ", length-helper.Width(line)))
	}

	if large {
		messages = append(messages, strings.Repeat(" ", length))
	}

	for i := range messages {
		messages[i] = fmt.Sprintf("<%s>%s</%s>", style, messages[i], style)
	}

	return strings.Join(messages, "\n")
}

func Truncate(message string, length int, suffix string) string {
	if suffix == "" {
		suffix = "..."
	}

	computedLength := length - helper.Width(suffix)
	if computedLength > helper.Width(message) {
		return message + suffix
	}

	return message[:length-1] + suffix
}
