package formatter

type OutputFormatterStyleStack struct {
	styles     []OutputFormatterStyleInterface
	emptyStyle OutputFormatterStyleInterface
}

func NewOutputFormatterStyleStack(emptyStyle OutputFormatterStyleInterface) *OutputFormatterStyleStack {
	if emptyStyle == nil {
		emptyStyle = NewOutputFormatterStyle("", "", nil)
	}

	s := &OutputFormatterStyleStack{
		styles:     make([]OutputFormatterStyleInterface, 0),
		emptyStyle: emptyStyle,
	}

	return s
}

func (s *OutputFormatterStyleStack) Reset() {
	s.styles = make([]OutputFormatterStyleInterface, 0)
}

func (s *OutputFormatterStyleStack) Clone() *OutputFormatterStyleStack {
	instance := NewOutputFormatterStyleStack(s.emptyStyle.Clone())
	for _, style := range s.styles {
		instance.Push(style)
	}
	return instance
}

func (s *OutputFormatterStyleStack) Push(style OutputFormatterStyleInterface) {
	s.styles = append(s.styles, style)
}

func (s *OutputFormatterStyleStack) Pop(style OutputFormatterStyleInterface) OutputFormatterStyleInterface {
	if s.styles == nil || len(s.styles) == 0 {
		return s.emptyStyle
	}

	if style == nil {
		last := s.styles[len(s.styles)-1]
		s.styles = s.styles[:len(s.styles)-1]
		return last
	}

	for i := len(s.styles) - 1; i >= 0; i-- {
		stackedStyles := s.styles[i]
		if style.Apply("") == stackedStyles.Apply("") {
			s.styles = s.styles[:i]
			return stackedStyles
		}
	}

	panic("Incorrectly nested style tag found.")
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
