package input

import "os"

type StreamableInputInterface interface {
	SetStream(stream *os.File)
	GetStream() *os.File
}
