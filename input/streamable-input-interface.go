package input

type StreamableInputInterface interface {
	SetStream(stream interface{})
	GetStream() interface{}
}
