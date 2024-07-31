package output

import "os"

type StreamOutputInterface interface {
	Stream() *os.File
}
