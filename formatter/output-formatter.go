package formatter

import (
	"fmt"
	"regexp"
	"strings"

	err "github.com/michielnijenhuis/cli/error"
	"github.com/michielnijenhuis/cli/helper"
)

type OutputFormatter struct {
	decorated  bool
	styles     map[string]OutputFormatterStyleInterface
	styleStack *OutputFormatterStyleStack
}

func NewOutputFormatter(decorated bool, styles map[string]OutputFormatterStyleInterface) *OutputFormatter {
	if styles == nil {
		styles = make(map[string]OutputFormatterStyleInterface)
	}

	of := &OutputFormatter{
		decorated:  decorated,
		styles:     styles,
		styleStack: NewOutputFormatterStyleStack(nil),
	}

	of.SetStyle("error", NewOutputFormatterStyle("white", "red", nil))
	of.SetStyle("info", NewOutputFormatterStyle("white", "blue", nil))
	of.SetStyle("success", NewOutputFormatterStyle("white", "green", nil))
	of.SetStyle("ok", NewOutputFormatterStyle("white", "green", nil))
	of.SetStyle("warn", NewOutputFormatterStyle("black", "yellow", nil))
	of.SetStyle("warning", NewOutputFormatterStyle("black", "yellow", nil))
	of.SetStyle("caution", NewOutputFormatterStyle("black", "yellow", nil))
	of.SetStyle("comment", NewOutputFormatterStyle("yellow", "", nil))
	of.SetStyle("alert", NewOutputFormatterStyle("red", "", []string{"bold"}))
	of.SetStyle("header", NewOutputFormatterStyle("yellow", "", nil))
	of.SetStyle("highlight", NewOutputFormatterStyle("green", "", nil))
	of.SetStyle("question", NewOutputFormatterStyle("black", "cyan", nil))

	for name, style := range styles {
		of.SetStyle(name, style)
	}

	return of
}

func (o *OutputFormatter) SetDecorated(decorated bool) {
	o.decorated = decorated
}

func (o *OutputFormatter) IsDecorated() bool {
	return o.decorated
}

func (o *OutputFormatter) SetStyle(name string, style OutputFormatterStyleInterface) {
	o.styles[strings.ToLower(name)] = style
}

func (o *OutputFormatter) HasStyle(name string) bool {
	return o.styles[strings.ToLower(name)] != nil
}

func (o *OutputFormatter) GetStyle(name string) (OutputFormatterStyleInterface, error) {
	if !o.HasStyle(name) {
		return nil, err.NewInvalidArgumentError(fmt.Sprintf("Undefined style: \"%s\".", name))
	}

	return o.styles[strings.ToLower(name)], nil
}

func (o *OutputFormatter) Format(message string) string {
	return o.FormatAndWrap(message, 0)
}

func (o *OutputFormatter) FormatAndWrap(message string, width int) string {
	if message == "" {
		return message
	}

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
			o.styleStack.Pop(nil)
		} else {
			style := o.createStyleFromString(tag)

			if style == nil {
				output += o.applyCurrentStyle(text, output, width, currentLineLength)
			} else if open {
				o.styleStack.Push(style)
			} else {
				o.styleStack.Pop(style)
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

func (o *OutputFormatter) StyleStack() *OutputFormatterStyleStack {
	return o.styleStack
}

func (o *OutputFormatter) Clone() OutputFormatterInferface {
	clone := NewOutputFormatter(false, nil)

	clone.styleStack = o.styleStack.Clone()
	clone.decorated = o.decorated

	for key, value := range o.styles {
		clone.styles[key] = value.Clone()
	}

	return clone
}

func Escape(s string) string {
	re := regexp.MustCompile(`([^\\]|^)([<>])`)
	s = re.ReplaceAllString(s, "\\$1")

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

func (o *OutputFormatter) createStyleFromString(s string) OutputFormatterStyleInterface {
	if style, ok := o.styles[s]; ok {
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
		if o.decorated {
			return o.styleStack.Current().Apply(text)
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

	if o.decorated {
		for i, line := range lines {
			lines[i] = o.styleStack.Current().Apply(line)
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
