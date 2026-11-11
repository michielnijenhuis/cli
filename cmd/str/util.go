package main

import (
	"regexp"
	"strings"
)

var (
	wordBoundary1 = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	wordBoundary2 = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	nonAlphaNum   = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	s = wordBoundary2.ReplaceAllString(s, `${1} ${2}`)
	s = wordBoundary1.ReplaceAllString(s, `${1} ${2}`)
	s = nonAlphaNum.ReplaceAllString(s, " ")

	return strings.Fields(strings.ToLower(s))
}

func ToCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(words[0])

	for _, word := range words[1:] {
		result.WriteString(strings.ToUpper(word[:1]))
		result.WriteString(word[1:])
	}

	return result.String()
}

func ToKebabCase(s string) string {
	return strings.Join(splitWords(s), "-")
}

func ToPascalCase(s string) string {
	words := splitWords(s)

	var result strings.Builder
	for _, word := range words {
		result.WriteString(strings.ToUpper(word[:1]))
		result.WriteString(word[1:])
	}

	return result.String()
}

func ToSnakeCase(s string) string {
	return strings.Join(splitWords(s), "_")
}
