package descriptor

import (
	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/input"
)

type DescribeableApplication interface {
	All(namespace string) map[string]*command.Command
	ExtractNamespace(name string, limit int) string
	FindNamespace(namespace string) (string, error)
	Help() string
	Definition() *input.InputDefinition
}
