package internal

import (
	"fmt"
	"os/exec"
	"strconv"
)

type IO struct {
	Input      *Input
	Output     *Output
	definition *InputDefinition
}

func (io *IO) Spawn(cmd string, pipe bool, args ...string) *exec.Cmd {
	cp := exec.Command(cmd, args...)

	if pipe {
		cp.Stdin = io.Input.Stream
		cp.Stdout = io.Output.Stream
		cp.Stderr = io.Output.Stderr.Stream
	}

	return cp
}

func (io *IO) Exec(cmd string, args ...string) error {
	proc := io.Spawn(cmd, true, args...)

	if io.IsDebug() {
		io.Writelnf("$ %s\n", proc.String())
	}

	return proc.Run()
}

func (io *IO) NewLine(count uint) {
	for count > 0 {
		io.Output.Writeln("", 0)
		count--
	}
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
		return false
	}

	return val
}

func (io *IO) String(name string) string {
	str, err := io.Input.String(name)
	if err != nil {
		return ""
	}

	return str
}

func (io *IO) Int(name string) int {
	val, err := io.Input.String(name)
	if err != nil {
		return 0
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return num
}

func (io *IO) Array(name string) []string {
	arr, err := io.Input.Array(name)
	if err != nil {
		return []string{}
	}

	return arr
}

var escape = false

func (io *IO) Critical(v any) {
	io.Output.PrintByTheme("critical", v, escape)
}

func (io *IO) Error(v any) {
	io.Output.PrintByTheme("error", v, escape)
}

func (io *IO) Warning(v any) {
	io.Output.PrintByTheme("warning", v, escape)
}

func (io *IO) Caution(v any) {
	io.Output.PrintByTheme("caution", v, escape)
}

func (io *IO) Success(v any) {
	io.Output.PrintByTheme("success", v, escape)
}

func (io *IO) Debug(v any) {
	io.Output.PrintByTheme("debug", v, escape)
}

func (io *IO) Info(v any) {
	io.Output.PrintByTheme("info", v, escape)
}

func (io *IO) Ok(v any) {
	io.Output.PrintByTheme("ok", v, escape)
}

func (io *IO) Symbol(symbol string, color string, v any) {
	switch v := v.(type) {
	case error:
		io.Writelnf("<fg=%s>%s</> %s", color, symbol, v.Error())
	case string:
		io.Writelnf("<fg=%s>%s</> %s", color, symbol, v)
	case []string:
		if len(v) == 0 {
			return
		}

		v[0] = fmt.Sprintf("<fg=%s>%s</> %s", color, symbol, v[0])
		io.Writelns(v)
	case nil:
		return
	default:
		io.Writelnf("<fg=%s>%s</> %v", color, symbol, v)
	}
}
