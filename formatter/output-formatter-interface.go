package formatter

type OutputFormatterInferface interface {
	SetDecorated(decorated bool)
	IsDecorated() bool
	SetStyle(name string, style OutputFormatterStyleInterface)
	HasStyle(name string) bool
	Style(name string) (OutputFormatterStyleInterface, error)
	Format(message string) string
	Clone() OutputFormatterInferface
}
