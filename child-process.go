package cli

import (
	"fmt"
	"os"
	os_exec "os/exec"
	"strings"

	"github.com/michielnijenhuis/cli/terminal"
)

type ChildProcess struct {
	Cmd    string
	Args   []string
	Shell  string
	Pipe   bool
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
	err    error
	Env    []string
	c      *os_exec.Cmd
}

func (cp *ChildProcess) Run() (string, error) {
	if cp.c == nil {
		cp.c = cp.createCommand()
	}

	if cp.Pipe {
		output, err := cp.c.Output()
		return string(output), err
	}

	err := cp.c.Run()
	return "", err
}

func (cp *ChildProcess) Start() error {
	if cp.c == nil {
		cp.c = cp.createCommand()
	}

	cp.err = cp.c.Start()
	return cp.err
}

func (cp *ChildProcess) AddEnv(name string, value string) {
	if cp.Env == nil {
		cp.Env = make([]string, 0, 1)
	}

	cp.Env = append(cp.Env, fmt.Sprintf("%s=%s", name, value))
}

func (cp *ChildProcess) Wait() error {
	if cp.c == nil {
		cp.c = cp.createCommand()
		cp.err = cp.c.Start()
	}

	err := cp.c.Wait()
	cp.err = err
	return err
}

func (cp *ChildProcess) createCommand() *os_exec.Cmd {
	cmd := cp.prepareCommand(cp.Cmd, cp.Shell)

	args := cp.Args
	if cmd != "" && args == nil {
		args = StringToInputArgs(cmd)
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
		cp.inherit(c)
	}

	cp.configureEnv(c)

	return c
}

func (cp *ChildProcess) inherit(c *os_exec.Cmd) {
	if cp.Stdin == nil {
		cp.Stdin = os.Stdin
	}

	if cp.Stdout == nil {
		cp.Stdout = os.Stdout
	}

	if cp.Stderr == nil {
		cp.Stderr = os.Stderr
	}

	c.Stdin = cp.Stdin
	c.Stdout = cp.Stdout
	c.Stderr = cp.Stderr
}

func (cp *ChildProcess) configureEnv(c *os_exec.Cmd) {
	if cp.Env != nil {
		env := os.Environ()
		env = append(env, cp.Env...)
		c.Env = env
	}
}

func (cp *ChildProcess) String() string {
	if cp.c == nil {
		cp.c = cp.createCommand()
	}

	return cp.c.String()
}

func (cp *ChildProcess) prepareCommand(cmd string, shell string) string {
	if shell == "" {
		return cmd
	}

	flags := make([]string, 0, 2)
	if !cp.Pipe && terminal.IsInteractive() {
		flags = append(flags, "-i")
	}
	flags = append(flags, "-c")

	parts := []string{shell, strings.Join(flags, " ")}

	switch shell {
	case "zsh":
		parts = append(parts, fmt.Sprintf("'source $HOME/dotfiles/.zshrc.base; %s'", cmd))
	default:
		parts = append(parts, fmt.Sprintf("'%s'", cmd))
	}

	cmd = strings.Join(parts, " ")

	return cmd
}
