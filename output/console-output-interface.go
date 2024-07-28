package output

type ConsoleOutputInterface interface {
	GetErrorOutput() OutputInterface
	SetErrorOutput(error OutputInterface)
	Section() ConsoleSectionOutput
}
