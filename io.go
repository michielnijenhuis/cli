package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type IO struct {
	Input      *Input
	Output     *Output
	Args       []string
	definition *InputDefinition
}

func (io *IO) ChildProcess(cmd string) *ChildProcess {
	return io.Spawn(cmd, "", true)
}

func (io *IO) Spawn(cmd string, shell string, inherit bool) *ChildProcess {
	cp := &ChildProcess{
		Cmd:   cmd,
		Shell: shell,
		Pipe:  !inherit,
	}

	if inherit {
		i := io.Input
		o := io.Output

		cp.Stdin = i.Stream
		cp.Stdout = o.Stream
		cp.Stderr = o.Stderr.Stream
	}

	return cp
}

func (io *IO) Exec(cmd string, shell string, inherit bool) (string, error) {
	return io.Spawn(cmd, shell, inherit).Run()
}

func (io *IO) Zsh(cmd string) error {
	_, err := io.Exec(cmd, "zsh", true)
	return err
}

func (io *IO) ZshPipe(cmd string) (string, error) {
	return io.Exec(cmd, "zsh", false)
}

func (io *IO) Sh(cmd string) error {
	_, err := io.Exec(cmd, "", true)
	return err
}

func (io *IO) ShPipe(cmd string) (string, error) {
	return io.Exec(cmd, "", false)
}

func (io *IO) NewLine(count uint) {
	for count > 0 {
		io.Output.Writeln("", 0)
		count--
	}
}

func (io *IO) Error(messages ...string) {
	io.Output.Error(messages...)
}

func (io *IO) Err(err error) {
	io.Output.Err(err)
}

func (io *IO) Info(messages ...string) {
	io.Output.Info(messages...)
}

func (io *IO) Warn(messages ...string) {
	io.Output.Warning(messages...)
}

func (io *IO) Note(messages ...string) {
	io.Output.Note(messages...)
}

func (io *IO) Ok(messages ...string) {
	io.Output.Ok(messages...)
}

func (io *IO) Success(messages ...string) {
	io.Output.Success(messages...)
}

func (io *IO) Comment(messages ...string) {
	io.Output.Comment(messages...)
}

func (io *IO) Alert(messages ...string) {
	length := 0
	for _, message := range messages {
		length = max(length, len(message))
	}
	length += 12

	io.writeLine([]string{fmt.Sprintf("<fg=yellow>%s </>", strings.Repeat("*", length))}, "alert", "", false)
	for i := range messages {
		messages[i] = fmt.Sprintf("%s<fg=yellow>*</>     %s     <fg=yellow>*</>", strings.Repeat(" ", 8), messages[i])
	}
	io.Writelns(messages)
	io.Writeln(fmt.Sprintf("<fg=yellow>%s%s</>", strings.Repeat(" ", 8), strings.Repeat("*", length)))
	io.NewLine(1)
}

func (io *IO) Write(message string) {
	io.Output.Write(message, false, 0)
}

func (io *IO) Writef(format string, a ...any) {
	io.Write(fmt.Sprintf(format, a...))
}

func (io *IO) Writeln(message string) {
	io.Output.Writeln(message, 0)
}

func (io *IO) Writelnf(format string, a ...any) {
	io.Writeln(fmt.Sprintf(format, a...))
}

func (io *IO) Writelns(messages []string) {
	io.Output.Writelns(messages, 0)
}

func (io *IO) writeLine(messages []string, tag string, label string, fullyColored bool) {
	if len(messages) == 0 {
		return
	}

	if tag != "" {
		if label == "" {
			label = fmt.Sprintf(" %s ", strings.ToUpper(tag))
		}

		closingTagLabelFmt := "</%s>"
		closingTagLabel := tag
		closingTagTextFmt := "%s"
		closingTagText := ""
		if fullyColored {
			closingTagTextFmt = closingTagLabelFmt
			closingTagText = closingTagLabel
			closingTagLabelFmt = "%s"
			closingTagLabel = ""
		}

		format := `<%s>%s` + closingTagLabelFmt + ` %s` + closingTagTextFmt
		messages[0] = fmt.Sprintf(format, tag, label, closingTagLabel, messages[0], closingTagText)

		for i := 1; i < len(messages); i++ {
			messages[i] = strings.Repeat(" ", len(tag)+3) + messages[i]
		}
	}

	io.Output.Writelns(messages, 0)
}

func (io *IO) Spinner(fn func(), message string) {
	style, _ := io.Output.Formatter().Style("prompt")
	s := NewSpinner(io.Input, io.Output, message, nil, style.foreground)
	s.Spin(fn)
}

func (io *IO) IsQuiet() bool {
	return io.Output.IsQuiet()
}

func (io *IO) IsVerbose() bool {
	return io.Output.IsVerbose()
}

func (io *IO) IsVeryVerbose() bool {
	return io.Output.IsVeryVerbose()
}

func (io *IO) IsDebug() bool {
	return io.Output.IsDebug()
}

func (io *IO) IsDecorated() bool {
	return io.Output.IsDecorated()
}

func (io *IO) Bool(name string) bool {
	val, err := io.Input.Bool(name)
	if err != nil {
		if io.Output.IsDebug() {
			panic(err)
		}
		return false
	}
	return val
}

func (io *IO) String(name string) string {
	str, err := io.Input.String(name)
	if err != nil {
		if io.Output.IsDebug() {
			panic(err)
		}
		return ""
	}
	return str
}

func (io *IO) Array(name string) []string {
	arr, err := io.Input.Array(name)
	if err != nil {
		if io.Output.IsDebug() {
			panic(err)
		}
		return []string{}
	}
	return arr
}

func (io *IO) Ask(question string, defaultValue string) (string, error) {
	return io.Output.Ask(question, defaultValue)
}

func (io *IO) Confirm(question string, defaultValue bool) (bool, error) {
	return io.Output.Confirm(question, defaultValue)
}

func (io *IO) Choice(question string, choices map[string]string, defaultValue string) (string, error) {
	return io.Output.Choice(question, choices, defaultValue)
}

func (io *IO) View() *View {
	return NewView(io.Output)
}

func (io *IO) WithGracefulExit(fn func(done <-chan bool)) bool {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go func(d chan bool) {
		fn(d)
		d <- true
	}(done)

	go func(s <-chan os.Signal, d chan<- bool) {
		<-s
		d <- false
	}(sigs, done)

	success := <-done
	return success
}
