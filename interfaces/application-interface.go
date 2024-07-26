package interfaces

import (
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
)

type ApplicationInterface interface {
	Run(input input.InputInterface, output output.OutputInterface) (int, error)
	SetDefinition(definition *input.InputDefinition)
	GetHelp() string
	AreErrorsCaught() bool
	SetCatchErrors(boolean bool)
	IsAutoExitEnabled() bool
	SetAutoExit(boolean bool)
	GetName() string
	SetName(name string)
	GetVersion() string
	SetVersion(version string)
}
