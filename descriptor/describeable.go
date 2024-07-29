package descriptor

import (
	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/input"
)

type DescribeableApplication interface {
	All(namespace string) map[string]*command.Command
	ExtractNamespace(name string, limit int) string
	FindNamespaces(namespace string) string
	Help() string
	Definition() *input.InputDefinition
}
