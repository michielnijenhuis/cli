package output

type ConsoleOutputInterface interface {
	ErrorOutput() OutputInterface
	SetErrorOutput(error OutputInterface)
	Section() ConsoleSectionOutput
}
