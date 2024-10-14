package cli

type OutputFormatterStyleStack struct {
	Styles     []*OutputFormatterStyle
	EmptyStyle *OutputFormatterStyle
}

func (s *OutputFormatterStyleStack) Reset() {
	s.Styles = make([]*OutputFormatterStyle, 0)
}

func (s *OutputFormatterStyleStack) Clone() *OutputFormatterStyleStack {
	e := s.EmptyStyle.Clone()
	if e == nil {
		e = makeEmptyStyle()
	}

	instance := &OutputFormatterStyleStack{
		EmptyStyle: e,
	}

	for _, style := range s.Styles {
		instance.Push(style)
	}

	return instance
}

func (s *OutputFormatterStyleStack) Push(style *OutputFormatterStyle) {
	if s.Styles == nil {
		s.Styles = make([]*OutputFormatterStyle, 0)
	}

	s.Styles = append(s.Styles, style)
}

func (s *OutputFormatterStyleStack) Pop(style *OutputFormatterStyle) *OutputFormatterStyle {
	if len(s.Styles) == 0 {
		if s.EmptyStyle == nil {
			s.EmptyStyle = makeEmptyStyle()
		}

		return s.EmptyStyle
	}

	if style == nil {
		last := s.Styles[len(s.Styles)-1]
		s.Styles = s.Styles[:len(s.Styles)-1]
		return last
	}

	for i := len(s.Styles) - 1; i >= 0; i-- {
		stackedStyles := s.Styles[i]
		if style.Apply("") == stackedStyles.Apply("") {
			s.Styles = s.Styles[:i]
			return stackedStyles
		}
	}

	return nil
}

func (s *OutputFormatterStyleStack) Current() *OutputFormatterStyle {
	if len(s.Styles) == 0 {
		if s.EmptyStyle == nil {
			s.EmptyStyle = makeEmptyStyle()
		}

		return s.EmptyStyle
	}

	return s.Styles[len(s.Styles)-1]
}

func makeEmptyStyle() *OutputFormatterStyle {
	return NewOutputFormatterStyle("", "", nil)
}
