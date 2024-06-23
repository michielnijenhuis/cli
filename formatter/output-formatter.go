package formatter

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
	panic("TODO: OutputFormatter.Escape()")
}
