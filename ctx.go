package cli

import (
	"fmt"
	"strings"
)

type Ctx struct {
	Input      *Input
	Output     *Output
	Args       []string
	Code       int
	Debug      bool
	definition *InputDefinition
}

func (c *Ctx) ChildProcess(cmd string) *ChildProcess {
	return c.Spawn(cmd, "", true)
}

func (c *Ctx) Spawn(cmd string, shell string, inherit bool) *ChildProcess {
	cp := &ChildProcess{
		Cmd:     cmd,
		Shell:   shell,
		Inherit: inherit,
		Pipe:    !inherit,
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

func (c *Ctx) NewLine(count uint) {
	for count > 0 {
		c.Output.Writeln("", 0)
		count--
	}
}

func (c *Ctx) Error(messages ...string) {
	c.writeLine(messages, "error")
}

func (c *Ctx) Err(err error) {
	c.writeLine([]string{err.Error()}, "error")
}

func (c *Ctx) Info(messages ...string) {
	c.writeLine(messages, "info")
}

func (c *Ctx) Warn(messages ...string) {
	c.writeLine(messages, "warning")
}

func (c *Ctx) Ok(messages ...string) {
	c.writeLine(messages, "ok")
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

	c.writeLine([]string{fmt.Sprintf("<fg=yellow>%s </>", strings.Repeat("*", length))}, "alert")
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

func (c *Ctx) writeLine(messages []string, tag string) {
	if len(messages) == 0 {
		return
	}

	if tag != "" {
		messages[0] = fmt.Sprintf("<%s> %s </%s> %s", tag, strings.ToUpper(tag), tag, messages[0])

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
	return c.Output.Ask(question, defaultValue, nil)
}

func (c *Ctx) Confirm(question string, defaultValue bool) (bool, error) {
	return c.Output.Confirm(question, defaultValue)
}

func (c *Ctx) Choice(question string, choices map[string]string, defaultValue string) (string, error) {
	return c.Output.Choice(question, choices, defaultValue)
}
