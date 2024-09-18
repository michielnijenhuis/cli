package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Ctx struct {
	Input      *Input
	Output     *Output
	Args       []string
	Code       int
	Debug      bool
	definition *InputDefinition
	Logger
}

func (c *Ctx) ChildProcess(cmd string) *ChildProcess {
	return c.Spawn(cmd, "", true)
}

func (c *Ctx) Spawn(cmd string, shell string, inherit bool) *ChildProcess {
	cp := &ChildProcess{
		Cmd:   cmd,
		Shell: shell,
		Pipe:  !inherit,
	}

	if inherit {
		i := c.Input
		o := c.Output

		cp.Stdin = i.Stream
		cp.Stdout = o.Stream
		cp.Stderr = o.Stderr.Stream
	}

	return cp
}

func (c *Ctx) Exec(cmd string, shell string, inherit bool) (string, error) {
	return c.Spawn(cmd, shell, inherit).Run()
}

func (c *Ctx) Zsh(cmd string) error {
	_, err := c.Exec(cmd, "zsh", true)
	return err
}

func (c *Ctx) ZshPipe(cmd string) (string, error) {
	return c.Exec(cmd, "zsh", false)
}

func (c *Ctx) Sh(cmd string) error {
	_, err := c.Exec(cmd, "", true)
	return err
}

func (c *Ctx) ShPipe(cmd string) (string, error) {
	return c.Exec(cmd, "", false)
}

func (c *Ctx) NewLine(count uint) {
	for count > 0 {
		c.Output.Writeln("", 0)
		count--
	}
}

func (c *Ctx) Error(messages ...string) {
	c.Output.Error(messages...)
}

func (c *Ctx) Err(err error) {
	c.Output.Err(err)
}

func (c *Ctx) Info(messages ...string) {
	c.Output.Info(messages...)
}

func (c *Ctx) Warn(messages ...string) {
	c.Output.Warning(messages...)
}

func (c *Ctx) Note(messages ...string) {
	c.Output.Note(messages...)
}

func (c *Ctx) Ok(messages ...string) {
	c.Output.Ok(messages...)
}

func (c *Ctx) Success(messages ...string) {
	c.Output.Success(messages...)
}

func (c *Ctx) Comment(messages ...string) {
	c.Output.Comment(messages...)
}

func (c *Ctx) Alert(messages ...string) {
	length := 0
	for _, message := range messages {
		length = max(length, len(message))
	}
	length += 12

	c.writeLine([]string{fmt.Sprintf("<fg=yellow>%s </>", strings.Repeat("*", length))}, "alert", "", false)
	for i := range messages {
		messages[i] = fmt.Sprintf("%s<fg=yellow>*</>     %s     <fg=yellow>*</>", strings.Repeat(" ", 8), messages[i])
	}
	c.Writelns(messages)
	c.Writeln(fmt.Sprintf("<fg=yellow>%s%s</>", strings.Repeat(" ", 8), strings.Repeat("*", length)))
	c.NewLine(1)
}

func (c *Ctx) Write(message string) {
	c.Output.Write(message, false, 0)
}

func (c *Ctx) Writef(format string, a ...any) {
	c.Write(fmt.Sprintf(format, a...))
}

func (c *Ctx) Writeln(message string) {
	c.Output.Writeln(message, 0)
}

func (c *Ctx) Writelnf(format string, a ...any) {
	c.Writeln(fmt.Sprintf(format, a...))
}

func (c *Ctx) Writelns(messages []string) {
	c.Output.Writelns(messages, 0)
}

func (c *Ctx) writeLine(messages []string, tag string, label string, fullyColored bool) {
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

	c.Output.Writelns(messages, 0)
}

func (c *Ctx) Spinner(fn func(), message string) {
	style, _ := c.Output.Formatter().Style("prompt")
	s := NewSpinner(c.Input, c.Output, message, nil, style.foreground)
	s.Spin(fn)
}

func (c *Ctx) IsQuiet() bool {
	return c.Output.IsQuiet()
}

func (c *Ctx) IsVerbose() bool {
	return c.Output.IsVerbose()
}

func (c *Ctx) IsVeryVerbose() bool {
	return c.Output.IsVeryVerbose()
}

func (c *Ctx) IsDebug() bool {
	return c.Output.IsDebug() || c.Debug
}

func (c *Ctx) IsDecorated() bool {
	return c.Output.IsDecorated()
}

func (c *Ctx) Bool(name string) bool {
	val, err := c.Input.Bool(name)
	if err != nil {
		if c.Debug {
			panic(err)
		}
		return false
	}
	return val
}

func (c *Ctx) String(name string) string {
	str, err := c.Input.String(name)
	if err != nil {
		if c.Debug {
			panic(err)
		}
		return ""
	}
	return str
}

func (c *Ctx) Array(name string) []string {
	arr, err := c.Input.Array(name)
	if err != nil {
		if c.Debug {
			panic(err)
		}
		return []string{}
	}
	return arr
}

func (c *Ctx) Ask(question string, defaultValue string) (string, error) {
	return c.Output.Ask(question, defaultValue)
}

func (c *Ctx) Confirm(question string, defaultValue bool) (bool, error) {
	return c.Output.Confirm(question, defaultValue)
}

func (c *Ctx) Choice(question string, choices map[string]string, defaultValue string) (string, error) {
	return c.Output.Choice(question, choices, defaultValue)
}

func (c *Ctx) View() *View {
	return NewView(c.Output)
}

func (c *Ctx) WithGracefulExit(fn func(done <-chan bool)) bool {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go func(d chan bool) {
		fn(d)
		d <- true
	}(done)

	go func(s chan os.Signal, d chan bool) {
		<-s
		d <- false
	}(sigs, done)

	success := <-done
	return success
}
