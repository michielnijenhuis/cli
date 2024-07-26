package formatter

import (
	"fmt"
	"regexp"
	"strings"

	err "github.com/michielnijenhuis/cli/error"
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
		styleStack: NewOutputFormatterStyleStack(),
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

// TODO: implement
func (o *OutputFormatter) FormatAndWrap(message string, width int) string {
	if message == "" {
		return message
	}

	return message
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
