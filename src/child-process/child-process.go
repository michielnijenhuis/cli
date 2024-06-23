package child_process

import (
	"fmt"
	// "os/exec"
	"strings"
)

type ChildProcess struct {
	in       int
	out      int
	err      int
	encoding string
	start    int
	end      int
}

func NewChildProcess() ChildProcess {
	return ChildProcess{}
}

func (p *ChildProcess) Run(cmd string, shell string) (string, error) {
	cmd = prepareCommand(cmd, shell)
	// args := parseArgs(cmd)
	// name := args[0]
	// args = args[1:]

	panic("TODO: ChildProcess.Run()")
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
	quote := ""
	quotes := []string{"'", "\""}

	for _, segment := range segments {
		if quote == "" {
			for _, char := range quotes {
				if strings.HasPrefix(segment, char) && !strings.HasSuffix(segment, char) {
					stack = append(stack, segment)
					quote = char
					break
				}
			}

			if quote == "" {
				out = append(out, segment)
			}
		} else if strings.HasSuffix(segment, quote) {
			stack = append(stack, segment)
			out = append(out, strings.Join(stack, " "))
			stack = make([]string, 0)
			quote = ""
		} else {
			stack = append(stack, segment)
		}
	}

	if len(stack) > 0 {
		out = append(out, stack...)
	}

	return out
}
