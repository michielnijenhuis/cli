package clidebug

import (
	"fmt"
	"os"
	"time"
)

func LogToFile(format string, args ...any) error {
	path := "./cli.log"
	var file *os.File
	var err error

	// create the file if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			return err
		}
	} else {
		// get the file handle
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}

	// format the string with timestamp included
	format = fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), format)

	// write to the file
	_, err = file.WriteString(fmt.Sprintf(format, args...) + "\n")
	if err != nil {
		return err
	}

	return nil
}
