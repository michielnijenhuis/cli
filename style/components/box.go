package components

type Box struct {
	output   string
	minWidth int
}

func NewBox() *Box {
	return &Box{
		output:   "",
		minWidth: 60,
	}
}

// TODO: implement
func (b *Box) Draw(title string, body string, footer string, color string, info string) string {
	return ""
}
