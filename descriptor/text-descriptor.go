package descriptor

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
)

type TextDescriptor struct {
	output output.OutputInterface
}

func NewTextDescriptor(output output.OutputInterface) *TextDescriptor {
	return &TextDescriptor{
		output: output,
	}
}

func (d *TextDescriptor) DescribeApplication(app DescribeableApplication, options *DescriptorOptions) {
	var describedNamespace string
	if options != nil {
		describedNamespace = options.namespace
	}

	description := NewApplicationDescription(app, describedNamespace, false)

	if options != nil && options.rawText {
		commands := description.Commands()
		width := columnWidth(commands)

		for _, command := range commands {
			name := command.Name
			d.writeText(fmt.Sprintf("%s %s", name[0:width], command.Description), options)
			d.writeText("\n", nil)
		}
	} else {
		help := app.Help()
		if help != "" {
			d.writeText("Help\n\n", options)
		}

		d.writeText("<header>Usage:</header>\n", options)
		d.writeText("  command [options] [arguments]\n\n", options)

		d.DescribeInputDefinition(input.NewInputDefinition(nil, app.Definition().OptionsArray()), options)

		d.writeText("\n", nil)
		d.writeText("\n", nil)

		commands := description.Commands()
		namespaces := description.Namespaces()

		var firstNamespace *NamespaceCommands
		for _, ns := range namespaces {
			firstNamespace = ns
			break
		}

		if describedNamespace != "" && len(namespaces) > 0 {
			// make sure all alias commands are included when describing a specific namespace
			describedNamespaceInfo := firstNamespace
			for _, name := range describedNamespaceInfo.commands {
				c, err := description.Command(name)
				if err != nil {
					commands[name] = c
				}
			}
		}

		// calculate max. width based on available commands per namespace
		availableCommands := make(map[string]*command.Command)
		for _, namespace := range namespaces {
			for _, command := range namespace.commands {
				if c, exists := commands[command]; exists {
					availableCommands[command] = c
				}
			}
		}
		width := columnWidth(availableCommands)

		if describedNamespace != "" {
			d.writeText(fmt.Sprintf("<header>Available commands for the \"%s\" namespace:</header>", describedNamespace), options)
		} else {
			d.writeText("<header>Available commands:</header>", options)
		}

		for _, namespace := range namespaces {
			list := make([]string, 0)
			for _, command := range namespace.commands {
				if _, exists := commands[command]; exists {
					list = append(list, command)
				}
			}

			if len(list) == 0 {
				continue
			}

			if describedNamespace == "" && namespace.id != "_global" {
				d.writeText("\n", nil)
				d.writeText(fmt.Sprintf(" <header>%s</header>", namespace.id), options)
			}

			for _, name := range list {
				d.writeText("\n", nil)
				spacingWidth := width - helper.Width(name)
				command := commands[name]

				var commandAliases string
				if name == command.Name {
					commandAliases = d.commandAliasesText(command)
				}

				d.writeText(fmt.Sprintf("  <highlight>%s</highlight>%s%s%s", name, strings.Repeat(" ", max(spacingWidth, 2)), commandAliases, command.Description), options)
			}
		}

		d.writeText("\n", nil)
	}
}

func (d *TextDescriptor) DescribeCommand(command *command.Command, options *DescriptorOptions) {
	command.MergeApplication(false)

	description := command.Description
	if description != "" {
		d.writeText("<header>Description:</header>", options)
		d.writeText("\n", nil)
		d.writeText("  "+description, nil)
		d.writeText("\n\n", nil)
	}

	d.writeText("<header>Usage:</header>", options)
	usages := make([]string, 0)
	usages = append(usages, command.Synopsis(true))
	if command.Aliases != nil {
		usages = append(usages, command.Aliases...)
	}
	usages = append(usages, command.Usages()...)
	for _, usage := range usages {
		d.writeText("\n", nil)
		d.writeText("  "+formatter.Escape(usage), options)
	}
	d.writeText("\n", nil)

	definition := command.Definition()
	if len(definition.Options()) > 0 || len(definition.Arguments()) > 0 {
		d.writeText("\n", nil)
		d.DescribeInputDefinition(definition, options)
		d.writeText("\n", nil)
	}

	help := command.ProcessedHelp()
	if help != "" && help != description {
		d.writeText("\n", nil)
		d.writeText("<header>Help:</header>", options)
		d.writeText("\n", nil)
		d.writeText("  "+strings.ReplaceAll(help, "\n", "\n  "), options)
		d.writeText("\n", nil)
	}
}

func (d *TextDescriptor) DescribeInputDefinition(definition *input.InputDefinition, options *DescriptorOptions) {
	totalWidth := calculateTotalWidthForOptions(definition.OptionsArray())
	for _, argument := range definition.Arguments() {
		totalWidth = max(totalWidth, helper.Width(argument.Name))
	}

	hasArgs := len(definition.Arguments()) > 0
	hasOptions := len(definition.Options()) > 0

	if hasArgs {
		d.writeText("<header>Arguments:</header>", options)
		d.writeText("\n", nil)

		for _, argument := range definition.Arguments() {
			d.DescribeInputArgument(argument, NewDescriptorOptions(options.namespace, options.rawText, options.short, totalWidth))
			d.writeText("\n", nil)
		}
	}

	if hasArgs && hasOptions {
		d.writeText("\n", nil)
	}

	if hasOptions {
		laterOptions := make([]*input.InputOption, 0)

		d.writeText("<header>Options:</header>", options)

		for _, option := range definition.Options() {
			if len(option.Shortcut) > 1 {
				laterOptions = append(laterOptions, option)
				continue
			}

			d.writeText("\n", nil)
			d.DescribeInputOption(option, NewDescriptorOptions(options.namespace, options.rawText, options.short, totalWidth))
		}

		for _, option := range laterOptions {
			d.writeText("\n", nil)
			d.DescribeInputOption(option, NewDescriptorOptions(options.namespace, options.rawText, options.short, totalWidth))
		}
	}
}

func (d *TextDescriptor) DescribeInputArgument(argument *input.InputArgument, options *DescriptorOptions) {
	var defaultValue string
	if hasDefaultValue(argument.DefaultValue) {
		defaultValue = fmt.Sprintf("<header> [default: %s]</header>", formatDefaultValue(argument.DefaultValue))
	}

	name := argument.Name

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = helper.Width(name)
	}

	spacingWidth := totalWidth - len(name) + 1
	width := strings.Repeat(" ", spacingWidth)
	re := regexp.MustCompile(`\s*[\r\n]\s*`)
	desc := re.ReplaceAllString(argument.Description, strings.Repeat(" ", totalWidth+4))

	d.writeText(fmt.Sprintf("  <highlight>%s</highlight> %s%s%s", name, width, desc, defaultValue), options)
}

func (d *TextDescriptor) DescribeInputOption(option *input.InputOption, options *DescriptorOptions) {
	var defaultValue string
	if hasDefaultValue(option.DefaultValue) {
		defaultValue = fmt.Sprintf("<header> [default: %s]</header>", formatDefaultValue(option.DefaultValue))
	}

	name := option.Name

	var value string
	if option.AcceptValue() {
		value = "=" + strings.ToUpper(name)

		if option.IsValueOptional() {
			value = "[" + value + "]"
		}
	}

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = calculateTotalWidthForOptions([]*input.InputOption{option})
	}

	var synopsis string
	if option.Shortcut != "" {
		synopsis = fmt.Sprintf("-%s, ", option.Shortcut)
	} else {
		synopsis = "    "
	}

	if option.IsNegatable() {
		synopsis += fmt.Sprintf("--%s|--no-%s", name, name)
	} else {
		synopsis += "--" + name
	}

	synopsis += value

	spacingWidth := max(0, totalWidth-helper.Width(synopsis))
	width := strings.Repeat(" ", spacingWidth)
	re := regexp.MustCompile(`\s*[\r\n]\s*`)
	desc := re.ReplaceAllString(option.Description, strings.Repeat(" ", totalWidth+4))

	var arr string
	if option.IsArray() {
		arr = "<header> (multiple values allowed)</header>"
	}

	d.writeText(fmt.Sprintf("  <highlight>%s</highlight>  %s%s%s%s", synopsis, width, desc, defaultValue, arr), options)
}

func columnWidth(commands map[string]*command.Command) int {
	width := 0

	for _, command := range commands {
		width = max(width, helper.Width(command.Name))

		if command.Aliases != nil {
			for _, alias := range command.Aliases {
				width = max(width, helper.Width(alias))
			}
		}
	}

	return width
}

func calculateTotalWidthForOptions(options []*input.InputOption) int {
	var totalWidth int

	for _, option := range options {
		// "-" + shortcut + ", --" + name
		nameLength := 1 + max(helper.Width(option.Shortcut), 1) + 4 + helper.Width(option.Name)

		if option.IsNegatable() {
			nameLength += 6 + helper.Width(option.Name) // |--no- + name
		} else if option.AcceptValue() {
			valueLength := 1 + helper.Width(option.Name) // = + value
			if option.IsValueOptional() {
				valueLength += 2 // [ + ]
			}

			nameLength += valueLength
		}

		totalWidth = max(totalWidth, nameLength)
	}

	return totalWidth
}

func (d *TextDescriptor) commandAliasesText(command *command.Command) string {
	var text string
	commandAliases := command.Aliases

	if commandAliases == nil {
		return text
	}

	aliases := make([]string, len(commandAliases))

	for _, alias := range commandAliases {
		segments := strings.Split(alias, ":")
		if len(segments) > 1 {
			segments = segments[1:]
		}
		aliases = append(aliases, strings.Join(segments, ":"))
	}

	if len(aliases) > 0 {
		text = fmt.Sprintf("[%s] ", strings.Join(aliases, "|"))
	}

	return text
}

func (d *TextDescriptor) writeText(content string, options *DescriptorOptions) {
	decorated := true

	if options != nil {
		if options.rawText {
			re := regexp.MustCompile(`<\/?[^>]+(>|$)`)
			content = re.ReplaceAllString(content, "")
		}
	}

	Write(d.output, content, decorated)
}

func hasDefaultValue(value input.InputType) bool {
	if value == nil {
		return false
	}

	arr, ok := value.([]interface{})
	if ok {
		return len(arr) > 0
	}

	_, isBool := value.(bool)
	if isBool {
		return false
	}

	str, ok := value.(string)
	if ok {
		return str != ""
	}

	return true
}

func formatDefaultValue(value input.InputType) string {
	if value == math.Inf(0) || value == math.Inf(-1) {
		return "INF"
	}

	number, ok := value.(int)
	if ok {
		return strconv.Itoa(number)
	}

	boolean, ok := value.(bool)
	if ok {
		if boolean {
			return "true"
		}

		return "false"
	}

	arr, ok := value.([]interface{})
	if ok {
		elements := make([]string, 0, len(arr))
		for _, el := range arr {
			str, ok := el.(string)
			if ok {
				elements = append(elements, formatter.Escape(str))
			}
		}
		return "[" + strings.Join(elements, ",") + "]"
	}

	return formatter.Escape(value.(string))
}
