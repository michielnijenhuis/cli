package formatter

type OutputFormatterInferface interface {
	SetDecorated(decorated bool)
	IsDecorated() bool
	SetStyle(name string, style OutputFormatterStyleInterface)
	HasStyle(name string) bool
	GetStyle(name string) OutputFormatterStyleInterface
	Format(message string) string
	Clone() OutputFormatterInferface
}
