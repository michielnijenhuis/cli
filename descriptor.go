package cli

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

type DescriptorOptions struct {
	namespace  string
	rawText    bool
	short      bool
	totalWidth int
}

type TextDescriptor struct {
	Output *Output
}

func (d *TextDescriptor) DescribeApplication(app *Application, options *DescriptorOptions) {
	checkPtr(app, "describable application")

	var describedNamespace string
	if options != nil {
		describedNamespace = options.namespace
	}

	description := ApplicationDescription{
		Application: app,
		Namespace:   describedNamespace,
	}

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
			d.writeText(help+"\n\n", options)
		}

		d.writeText("<primary>Usage:</primary>\n", options)
		d.writeText("  command [options] [arguments]\n\n", options)

		d.DescribeInputDefinition(&InputDefinition{
			Options: app.Definition().Options,
		}, options)

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
			checkPtr(describedNamespaceInfo, "described namespace info")
			for _, name := range describedNamespaceInfo.commands {
				c, err := description.Command(name)
				if err != nil {
					commands[name] = c
				}
			}
		}

		// calculate max. width based on available commands per namespace
		availableCommands := make(map[string]*Command)
		for _, namespace := range namespaces {
			for _, command := range namespace.commands {
				if c, exists := commands[command]; exists {
					availableCommands[command] = c
				}
			}
		}
		width := columnWidth(availableCommands)

		if describedNamespace != "" {
			d.writeText(fmt.Sprintf("<primary>Available commands for the \"%s\" namespace:</primary>", describedNamespace), options)
		} else {
			d.writeText("<primary>Available commands:</primary>", options)
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
				d.writeText(fmt.Sprintf(" <primary>%s</primary>", namespace.id), options)
			}

			for _, name := range list {
				d.writeText("\n", nil)
				spacingWidth := width - helper.Width(name)
				command := commands[name]

				var commandAliases string
				if name == command.Name {
					commandAliases = d.commandAliasesText(command)
				}

				d.writeText(fmt.Sprintf("  <accent>%s</accent>%s%s%s", name, strings.Repeat(" ", max(spacingWidth, 0)+2), commandAliases, command.Description), options)
			}
		}

		d.writeText("\n", nil)
	}
}

func (d *TextDescriptor) DescribeCommand(command *Command, options *DescriptorOptions) {
	checkPtr(command, "describable command")

	command.MergeApplication(false)

	description := command.Description
	if description != "" {
		d.writeText("<primary>Description:</primary>", options)
		d.writeText("\n", nil)
		d.writeText("  "+description, nil)
		d.writeText("\n\n", nil)
	}

	d.writeText("<primary>Usage:</primary>", options)
	usages := make([]string, 0)
	usages = append(usages, command.Synopsis(true))
	if command.Aliases != nil {
		usages = append(usages, command.Aliases...)
	}
	usages = append(usages, command.Usages()...)
	for _, usage := range usages {
		d.writeText("\n", nil)
		d.writeText("  "+Escape(usage), options)
	}
	d.writeText("\n", nil)

	definition := command.Definition()
	if len(definition.Options) > 0 || len(definition.Arguments) > 0 {
		d.writeText("\n", nil)
		d.DescribeInputDefinition(definition, options)
		d.writeText("\n", nil)
	}

	help := command.ProcessedHelp()
	if help != "" && help != description {
		d.writeText("\n", nil)
		d.writeText("<primary>Help:</primary>", options)
		d.writeText("\n", nil)
		d.writeText("  "+strings.ReplaceAll(help, "\n", "\n  "), options)
		d.writeText("\n", nil)
	}
}

func (d *TextDescriptor) DescribeInputDefinition(definition *InputDefinition, options *DescriptorOptions) {
	checkPtr(definition, "describable input definition")

	totalWidth := calculateTotalWidthForOptions(definition.Options)
	for _, argument := range definition.Arguments {
		totalWidth = max(totalWidth, helper.Width(argument.Name))
	}

	hasArgs := len(definition.Arguments) > 0
	hasOptions := len(definition.Options) > 0

	if hasArgs {
		d.writeText("<primary>Arguments:</primary>", options)
		d.writeText("\n", nil)

		for _, argument := range definition.Arguments {
			d.DescribeInputArgument(argument, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
			d.writeText("\n", nil)
		}
	}

	if hasArgs && hasOptions {
		d.writeText("\n", nil)
	}

	if hasOptions {
		laterOptions := make([]*InputOption, 0)

		d.writeText("<primary>Options:</primary>", options)

		for _, option := range definition.Options {
			if len(option.Shortcut) > 1 {
				laterOptions = append(laterOptions, option)
				continue
			}

			d.writeText("\n", nil)
			d.DescribeInputOption(option, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
		}

		for _, option := range laterOptions {
			d.writeText("\n", nil)
			d.DescribeInputOption(option, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
		}
	}
}

func (d *TextDescriptor) DescribeInputArgument(argument *InputArgument, options *DescriptorOptions) {
	checkPtr(argument, "describable input argument")

	var defaultValue string
	if hasDefaultValue(argument.DefaultValue) {
		defaultValue = fmt.Sprintf("<primary> [default: %s]</primary>", formatDefaultValue(argument.DefaultValue))
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

	d.writeText(fmt.Sprintf("  <accent>%s</accent> %s%s%s", name, width, desc, defaultValue), options)
}

func (d *TextDescriptor) DescribeInputOption(option *InputOption, options *DescriptorOptions) {
	checkPtr(option, "describable input option")

	var defaultValue string
	if hasDefaultValue(option.DefaultValue) {
		defaultValue = fmt.Sprintf("<primary> [default: %s]</primary>", formatDefaultValue(option.DefaultValue))
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
		totalWidth = calculateTotalWidthForOptions([]*InputOption{option})
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
		arr = "<primary> (multiple values allowed)</primary>"
	}

	d.writeText(fmt.Sprintf("  <accent>%s</accent>  %s%s%s%s", synopsis, width, desc, defaultValue, arr), options)
}

func columnWidth(commands map[string]*Command) int {
	width := 0

	for _, command := range commands {
		width = max(width, helper.Width(command.Name))

		for _, alias := range command.Aliases {
			width = max(width, helper.Width(alias))
		}
	}

	return width
}

func calculateTotalWidthForOptions(options []*InputOption) int {
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

func (d *TextDescriptor) commandAliasesText(command *Command) string {
	var text string
	commandAliases := command.Aliases

	if commandAliases == nil {
		return text
	}

	aliases := make([]string, 0, len(commandAliases))

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

	d.Write(content, decorated)
}

func hasDefaultValue(value InputType) bool {
	if value == nil {
		return false
	}

	arr, ok := value.([]any)
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

func formatDefaultValue(value InputType) string {
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

	arr, ok := value.([]string)
	if ok {
		return "[" + strings.Join(arr, ",") + "]"
	}

	return Escape(fmt.Sprintf("%v", value))
}

func (d *TextDescriptor) Write(content string, decorated bool) {
	if decorated {
		d.Output.Write(content, false, OutputNormal)
	} else {
		d.Output.Write(content, false, OutputRaw)
	}
}
