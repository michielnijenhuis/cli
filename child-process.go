package cli

import (
	"fmt"
	"os"
	os_exec "os/exec"
)

type ChildProcess struct {
	Cmd     string
	Args    []string
	Shell   string
	Pipe    bool
	Inherit bool
	Stdin   *os.File
	Stdout  *os.File
	Stderr  *os.File
	err     error
	c       *os_exec.Cmd
}

func (cp *ChildProcess) Run() (string, error) {
	c := cp.createCommand()

	if cp.Pipe {
		output, err := c.CombinedOutput()
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
	case "bash":
		return fmt.Sprintf("bash -i -c '%s'", cmd)
	default:
		return cmd
	}
}
