package formatter

type OutputFormatterStyleStack struct {
	styles     []OutputFormatterStyleInterface
	emptyStyle OutputFormatterStyleInterface
}

// TODO: implement
func NewOutputFormatterStyleStack() *OutputFormatterStyleStack {
	return &OutputFormatterStyleStack{}
}

// TODO: implement
func (s *OutputFormatterStyleStack) Clone() *OutputFormatterStyleStack {
	return &OutputFormatterStyleStack{}
}

func (s *OutputFormatterStyleStack) Push(style OutputFormatterStyleInterface) {
	s.styles = append(s.styles, style)
}

// TODO: implement
func (s *OutputFormatterStyleStack) Pop(style OutputFormatterStyleInterface) OutputFormatterStyleInterface {
	return nil
}

func (s *OutputFormatterStyleStack) Current() OutputFormatterStyleInterface {
	if len(s.styles) == 0 {
		return s.emptyStyle
	}

	return s.styles[len(s.styles)-1]
}

func (s *OutputFormatterStyleStack) SetEmptyStyle(style OutputFormatterStyleInterface) {
	s.emptyStyle = style
}

func (s *OutputFormatterStyleStack) EmptyStyle() OutputFormatterStyleInterface {
	return s.emptyStyle
}
