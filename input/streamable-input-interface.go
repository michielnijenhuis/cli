package input

import "os"

type StreamableInputInterface interface {
	SetStream(stream *os.File)
	Stream() *os.File
}
