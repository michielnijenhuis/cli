package exec

import (
	"fmt"
	"os"
	os_exec "os/exec"
	"strings"
)

type ChildProcessOptions struct {
	Args    []string
	Shell   string
	Pipe    bool
	Inherit bool
	Stdin   *os.File
	Stdout  *os.File
	Stderr  *os.File
}

type ChildProcess struct {
	*ChildProcessOptions
	Cmd string
	err error
	c   *os_exec.Cmd
}

func Exec(cmd string, options *ChildProcessOptions) (string, error) {
	if options == nil {
		options = &ChildProcessOptions{}
	}

	cp := &ChildProcess{
		Cmd:                 cmd,
		ChildProcessOptions: options,
	}

	return cp.Run()
}

func (cp *ChildProcess) Run() (string, error) {
	c := cp.createCommand()

	if cp.Pipe {
		output, err := c.Output()
		return string(output), err
	}

	err := c.Run()
	return "", err
}

func (cp *ChildProcess) Start() error {
	if cp.c != nil {
		return cp.err
	}

	cp.c = cp.createCommand()
	cp.err = cp.c.Start()
	return cp.err
}

func (cp *ChildProcess) Wait() error {
	if cp.c == nil {
		return nil
	}

	err := cp.c.Wait()
	cp.err = err
	return err
}

func (cp *ChildProcess) createCommand() *os_exec.Cmd {
	cp.init()

	cmd := prepareCommand(cp.Cmd, cp.Shell)

	args := cp.Args
	if cmd != "" && args == nil {
		args = parseArgs(cmd)
	} else if args == nil {
		args = []string{}
	}

	name := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	c := os_exec.Command(name, args...)

	if !cp.Pipe {
		c.Stdin = cp.Stdin
		c.Stdout = cp.Stdout
		c.Stderr = cp.Stderr
	}

	return c
}

func (cp *ChildProcess) init() {
	if !cp.Inherit {
		return
	}

	if cp.Stdin == nil {
		cp.Stdin = os.Stdin
	}

	if cp.Stdout == nil {
		cp.Stdout = os.Stdout
	}

	if cp.Stderr == nil {
		cp.Stderr = os.Stderr
	}
}

func prepareCommand(cmd string, shell string) string {
	switch shell {
	case "zsh":
		return fmt.Sprintf("zsh -i -c 'source ~/.zshrc; %s'", cmd)
	default:
		return cmd
	}
}

func parseArgs(cmd string) []string {
	segments := strings.Split(cmd, " ")
	out := make([]string, 0)
	stack := make([]string, 0)
	var i int
	var collectingBy string

	for i < len(segments) {
		current := segments[i]
		i++

		if collectingBy == "" {
			for _, char := range []string{"'", `"`} {
				if strings.HasPrefix(current, char) && !strings.HasSuffix(current, char) {
					stack = append(stack, current[1:])
					collectingBy = char
					break
				}
			}

			if collectingBy == "" {
				out = append(out, current)
			}

		} else if strings.HasSuffix(current, collectingBy) {
			stack = append(stack, current[:len(current)-1])
			out = append(out, strings.Join(stack, " "))
			stack = make([]string, 0)
			collectingBy = ""
		} else {
			stack = append(stack, current)
		}
	}

	if len(stack) > 0 {
		out = append(out, stack...)
	}

	return out
}
