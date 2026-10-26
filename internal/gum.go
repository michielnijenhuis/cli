package internal

import (
	"fmt"
	"strings"
)

type ArgList []string

func NewArgList(initial ...string) *ArgList {
	a := make(ArgList, 0, len(initial))
	for _, item := range initial {
		a = append(a, item)
	}

	return &a
}

func (a *ArgList) AddArg(arg string) {
	*a = append(*a, arg)
}

func (a *ArgList) AddArgs(args []string) {
	*a = append(*a, args...)
}

func (a *ArgList) AddStringFlag(key string, value string) {
	if value == "" {
		return
	}

	*a = append(*a, fmt.Sprintf("--%s=%s", key, value))
}

func (a *ArgList) AddBoolFlag(key string, value bool) {
	if !value {
		return
	}

	*a = append(*a, "--"+key)
}

func (a *ArgList) AddNegatableBoolFlag(key string, value bool) {
	if !value {
		*a = append(*a, fmt.Sprintf("--no-%s", key))
	} else {
		*a = append(*a, "--"+key)
	}
}

func (a *ArgList) AddIntFlag(key string, value int) {
	*a = append(*a, fmt.Sprintf("--%s=%d", key, value))
}

func (a *ArgList) String() string {
	return strings.Join(*a, " ")
}
