package cli

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/helper/array"
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
		d.writeText("  command [flags] [--] [arguments]\n\n", options)

		d.DescribeInputDefinition(&InputDefinition{
			flags: app.Definition().flags,
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
			d.writeText(fmt.Sprintf("<primary>Commands for the \"%s\" namespace:</primary>", describedNamespace), options)
		} else {
			d.writeText("<primary>Commands:</primary>", options)
		}

		for _, namespace := range array.SortedKeys(namespaces) {
			ns := namespaces[namespace]
			list := make([]string, 0)
			sort.Strings(ns.commands)
			for _, command := range ns.commands {
				if _, exists := commands[command]; exists {
					list = append(list, command)
				}
			}

			if len(list) == 0 {
				continue
			}

			if describedNamespace == "" && ns.id != "_global" {
				d.writeText("\n", nil)
				d.writeText(fmt.Sprintf(" <primary>%s</primary>", ns.id), options)
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
	if len(definition.flags) > 0 || len(definition.arguments) > 0 {
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
	totalWidth := calculateTotalWidthForFlags(definition.flags)
	for _, argument := range definition.arguments {
		totalWidth = max(totalWidth, helper.Width(argument.GetName()))
	}

	hasArgs := len(definition.arguments) > 0
	hasFlags := len(definition.flags) > 0

	if hasArgs {
		d.writeText("<primary>Arguments:</primary>", options)
		d.writeText("\n", nil)

		for _, argument := range definition.arguments {
			d.DescribeArgument(argument, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
			d.writeText("\n", nil)
		}
	}

	if hasArgs && hasFlags {
		d.writeText("\n", nil)
	}

	if hasFlags {
		laterFlags := make([]Flag, 0)

		d.writeText("<primary>Flags:</primary>", options)

		for _, flag := range definition.flags {
			if len(flag.GetShortcutString()) > 1 {
				laterFlags = append(laterFlags, flag)
				continue
			}

			d.writeText("\n", nil)
			d.DescribeFlag(flag, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
		}

		for _, flag := range laterFlags {
			d.writeText("\n", nil)
			d.DescribeFlag(flag, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
		}
	}
}

func (d *TextDescriptor) DescribeArgument(argument Arg, options *DescriptorOptions) {
	var defaultValue string
	if argHasDefaultValue(argument) {
		defaultValue = fmt.Sprintf("<primary> [default: %s]</primary>", formatArgValue(argument))
	}

	name := argument.GetName()

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = helper.Width(name)
	}

	spacingWidth := totalWidth - len(name) + 1
	width := strings.Repeat(" ", spacingWidth)
	re := regexp.MustCompile(`\s*[\r\n]\s*`)
	desc := re.ReplaceAllString(argument.GetDescription(), strings.Repeat(" ", totalWidth+4))

	d.writeText(fmt.Sprintf("  <accent>%s</accent> %s%s%s", name, width, desc, defaultValue), options)
}

func argHasDefaultValue(arg Arg) bool {
	if a, ok := arg.(*StringArg); ok {
		return a.Value != ""
	}
	if a, ok := arg.(*ArrayArg); ok {
		return len(a.Value) > 0
	}
	return false
}

func formatArgValue(arg Arg) string {
	switch a := arg.(type) {
	case *StringArg:
		return formatDefaultValue(a.Value)
	case *ArrayArg:
		return formatDefaultValue(a.Value)
	default:
		panic("invalid argument type")
	}
}

func formatFlagValue(flag Flag) string {
	switch f := flag.(type) {
	case *StringFlag:
		return formatDefaultValue(f.Value)
	case *ArrayFlag:
		return formatDefaultValue(f.Value)
	case *BoolFlag:
		return formatDefaultValue(f.Value)
	case *OptionalStringFlag:
		v := formatDefaultValue(f.Value)
		if v != "" {
			return v
		}
		return formatDefaultValue(f.Boolean)
	case *OptionalArrayFlag:
		v := formatDefaultValue(f.Value)
		if v != "" {
			return v
		}
		return formatDefaultValue(f.Boolean)
	default:
		panic("invalid argument type")
	}
}

func (d *TextDescriptor) DescribeFlag(flag Flag, options *DescriptorOptions) {
	var defaultValue string
	if FlagHasDefaultValue(flag) {
		defaultValue = fmt.Sprintf("<primary> [default: %s]</primary>", formatFlagValue(flag))
	}

	name := flag.GetName()

	var value string
	if FlagAcceptsValue(flag) {
		value = "=" + strings.ToUpper(name)

		if FlagValueIsOptional(flag) {
			value = "[" + value + "]"
		}
	}

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = calculateTotalWidthForFlags([]Flag{flag})
	}

	var synopsis strings.Builder
	if s := flag.GetShortcutString(); s != "" {
		synopsis.WriteString(fmt.Sprintf("-%s, ", s))
	} else {
		synopsis.WriteString("    ")
	}

	if FlagIsNegatable(flag) {
		synopsis.WriteString(fmt.Sprintf("--%s|--no-%s", name, name))
	} else {
		synopsis.WriteString("--" + name)
	}

	synopsis.WriteString(value)
	synopsisString := synopsis.String()
	spacingWidth := max(0, totalWidth-helper.Width(synopsisString))
	width := strings.Repeat(" ", spacingWidth)
	re := regexp.MustCompile(`\s*[\r\n]\s*`)
	desc := re.ReplaceAllString(flag.GetDescription(), strings.Repeat(" ", totalWidth+4))

	var arr string
	if FlagIsArray(flag) {
		arr = "<primary> (multiple values allowed)</primary>"
	}

	d.writeText(fmt.Sprintf("  <accent>%s</accent>  %s%s%s%s", synopsisString, width, desc, defaultValue, arr), options)
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

func calculateTotalWidthForFlags(flags []Flag) int {
	var totalWidth int

	for _, flag := range flags {
		// "-" + shortcut + ", --" + name
		nameLength := 1 + max(helper.Width(flag.GetShortcutString()), 1) + 4 + helper.Width(flag.GetName())

		if FlagIsNegatable(flag) {
			nameLength += 6 + helper.Width(flag.GetName()) // |--no- + name
		} else if FlagAcceptsValue(flag) {
			valueLength := 1 + helper.Width(flag.GetName()) // = + value
			if FlagValueIsOptional(flag) {
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

func formatDefaultValue(value InputType) string {
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
