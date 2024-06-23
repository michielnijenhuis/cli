package formatter

import (
	"regexp"
	"strings"
)

type OutputFormatter struct{}

func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{}
}

func (o *OutputFormatter) SetDecorated(decorated bool) {}

func (o *OutputFormatter) IsDecorated() bool {
	return false
}

func (o *OutputFormatter) SetStyle(name string, style OutputFormatterStyleInterface) {}

func (o *OutputFormatter) HasStyle(name string) bool {
	return false
}

func (o *OutputFormatter) GetStyle(name string) OutputFormatterStyleInterface {
	return nil
}

func (o *OutputFormatter) Format(message string) string {
	return ""
}

func (o *OutputFormatter) Clone() OutputFormatterInferface {
	return nil
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
			s = s[:currentLength - 1]
			currentLength--
		}

		s = strings.Replace(s, `\0`, "", 1)
		s += strings.Repeat(`\0`, length - len(s))
	}

	return s
}
