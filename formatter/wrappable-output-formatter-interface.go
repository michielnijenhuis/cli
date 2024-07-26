package formatter

type WrappableOutputFormatterInterface interface {
	FormatAndWrap(message string, width int) string
}
