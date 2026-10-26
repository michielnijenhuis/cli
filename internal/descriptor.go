package internal

import (
	"fmt"
	"slices"
	"strings"

	"github.com/michielnijenhuis/cli/internal/helper"
	"github.com/michielnijenhuis/cli/internal/helper/array"
)

type DescriptorOptions struct {
	totalWidth int
	indent     int
}

type TextDescriptor struct {
	Output *Output
}

func (d *TextDescriptor) DescribeCommand(command *Command, options *DescriptorOptions, namespacedCommands []*Command, namespace string) {
	if command.parent == nil && namespace == "" && command.Intro != "" {
		d.writeText(command.Intro + Eol + Eol)
	}

	introHelp := command.GetHelp()
	if introHelp != "" {
		d.writeText(introHelp + Eol + Eol)
	}

	definition, _ := command.Definition()

	d.writeText("<primary>Usage:</primary>")
	d.writeText(Eol)

	if command.Run != nil || command.RunE != nil {
		d.writeText("  " + command.Synopsis(true))

		if command.HasSubcommands() {
			d.writeText(Eol)
		}
	}

	if command.HasSubcommands() {
		if command.parent == nil {
			d.writeText(fmt.Sprintf("  <fg=gray>$</> %s <command> [flags] [--] [arguments]", command.FullName()))
		} else {
			d.writeText(fmt.Sprintf("  <fg=gray>$</> %s <command>", command.FullName()))
		}
	}

	ns := command.Namespace()
	if ns != "" {
		ns += " "
	}

	for _, alias := range command.Aliases {
		d.writeText(Eol)
		d.writeText(fmt.Sprintf("  <fg=gray>$</> %s%s", ns, alias))
	}

	d.writeText(Eol + Eol)

	d.DescribeInputDefinition(definition, options)

	if command.HasSubcommands() {
		if len(namespacedCommands) > 0 {
			d.writeText(Eol)
			d.writeText(Eol)
			d.writeText(fmt.Sprintf("<primary>Available commands for the \"%s\" namespace:</primary>", namespace))

			slices.SortFunc(namespacedCommands, func(a *Command, b *Command) int {
				return strings.Compare(a.Name, b.Name)
			})

			for _, cmd := range namespacedCommands {
				d.writeText(Eol)
				d.DescribeSubcommand(cmd, options)
			}
		} else {
			commands := command.All()

			commandsWithoutNamespace := make(map[string]*Command)
			namespacedCommands := make(map[string][]*Command)

			for _, cmd := range commands {
				if cmd.Hidden {
					continue
				}

				if !strings.Contains(cmd.Name, ":") {
					commandsWithoutNamespace[cmd.Name] = cmd
				} else {
					nsParts := strings.Split(cmd.Name, ":")
					namespace := strings.Join(nsParts[:len(nsParts)-1], ":")

					cmdsForNamespace, ok := namespacedCommands[namespace]
					if !ok {
						cmdsForNamespace = make([]*Command, 0)
					}

					namespacedCommands[namespace] = append(cmdsForNamespace, cmd)
				}
			}

			d.writeText(Eol)
			d.writeText(Eol)
			d.writeText("<primary>Available commands:</primary>")

			for _, name := range array.SortedKeys(commandsWithoutNamespace) {
				cmd := commandsWithoutNamespace[name]
				d.writeText(Eol)
				d.DescribeSubcommand(cmd, options)
			}

			for _, ns := range array.SortedKeys(namespacedCommands) {
				cmds := namespacedCommands[ns]
				d.writeText(Eol)
				d.writeText(fmt.Sprintf("<primary> %s</primary>", ns))

				for _, cmd := range cmds {
					d.writeText(Eol)
					d.DescribeSubcommand(cmd, options)
				}
			}
		}
	}

	d.writeText(Eol)

	help := command.ProcessedHelp()
	if help != "" && help != introHelp && help != command.Description {
		d.writeText(Eol)
		d.writeText("<primary>Help:</primary>")
		d.writeText(Eol)
		d.writeText("  " + strings.ReplaceAll(help, Eol, "\n  "))
		d.writeText(Eol)
	}
}

func (d *TextDescriptor) DescribeInputDefinition(definition *InputDefinition, options *DescriptorOptions) {
	hasArgs := len(definition.arguments) > 0
	hasFlags := len(definition.flags) > 0

	if hasArgs {
		d.writeText("<primary>Arguments:</primary>")
		d.writeText(Eol)

		for i, argument := range definition.arguments {
			d.DescribeArgument(argument, options)

			if i+1 < len(definition.arguments) || hasFlags {
				d.writeText(Eol)
			}
		}
	}

	if hasArgs && hasFlags {
		d.writeText(Eol)
	}

	if hasFlags {
		laterFlags := make([]Flag, 0)

		d.writeText("<primary>Flags:</primary>")

		for _, flag := range definition.flags {
			if len(flag.GetShortcutString()) > 1 {
				laterFlags = append(laterFlags, flag)
				continue
			}

			d.writeText(Eol)
			d.DescribeFlag(flag, options)
		}

		for _, flag := range laterFlags {
			d.writeText(Eol)
			d.DescribeFlag(flag, options)
		}
	}
}

func (d *TextDescriptor) DescribeArgument(argument Arg, options *DescriptorOptions) {
	var accent string
	if argHasDefaultValue(argument) {
		accent = fmt.Sprintf("<primary> [default: %s]</primary>", formatArgValue(argument))
	} else if argument.IsRequired() && len(argument.Opts()) > 0 {
		accent = fmt.Sprintf("<primary> [options: %s]</primary>", strings.Join(argument.Opts(), ", "))
	}

	name := argument.GetName()

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = helper.Width(name)
	}

	finalDescription := d.Output.Format(argument.GetDescription() + accent)
	wrappedDescriptionLines := helper.WrapToSlice(finalDescription, d.Output.lineLength-totalWidth, Eol, false)
	indent := strings.Repeat(" ", totalWidth+4)

	for i, line := range wrappedDescriptionLines {
		if i == 0 {
			wrappedDescriptionLines[i] = strings.Repeat(" ", totalWidth-helper.Width(name)) + line
		} else {
			wrappedDescriptionLines[i] = indent + line
		}
	}

	d.writeText(fmt.Sprintf("  <secondary>%s</secondary>  %s", name, strings.Join(wrappedDescriptionLines, "")))
}

func (d *TextDescriptor) DescribeSubcommand(command *Command, options *DescriptorOptions) {
	name := command.Name

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = helper.Width(name)
	}

	var commandAliases string
	if name == command.Name {
		commandAliases = d.commandAliasesText(command)
	}

	finalDescription := d.Output.Format(commandAliases + command.Description)
	wrappedDescriptionLines := helper.WrapToSlice(finalDescription, d.Output.lineLength-totalWidth, Eol, false)

	baseIndentSize := 2
	spacingSize := 2

	indentSize := baseIndentSize
	if options != nil && options.indent > 0 {
		indentSize += options.indent
		spacingSize -= options.indent
	}

	indent := strings.Repeat(" ", totalWidth+indentSize+spacingSize)
	spacing := strings.Repeat(" ", spacingSize)

	for i, line := range wrappedDescriptionLines {
		if i == 0 {
			wrappedDescriptionLines[i] = strings.Repeat(" ", totalWidth-helper.Width(name)) + line
		} else {
			wrappedDescriptionLines[i] = indent + line
		}
	}

	// indent := 2
	// if options != nil && options.indent > 0 {
	// 	indent += options.indent
	// }

	// indentSpaces := strings.Repeat(" ", indent)

	d.writeText(fmt.Sprintf("%s<secondary>%s</secondary>%s%s", strings.Repeat(" ", indentSize), name, spacing, strings.Join(wrappedDescriptionLines, "")))
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
		return ""
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
		return ""
	}
}

func (d *TextDescriptor) DescribeFlag(flag Flag, options *DescriptorOptions) {
	var defaultValue string
	if FlagHasDefaultValue(flag) {
		defaultValue = fmt.Sprintf("<accent> [default: %s]</accent>", formatFlagValue(flag))
	}

	name := flag.GetName()

	var value string
	if FlagAcceptsValue(flag) {
		value = "=" + strings.ToUpper(name)

		if FlagValueIsOptional(flag) {
			value = "[" + value + "]"
		}
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

	var totalWidth int
	if options != nil && options.totalWidth > 0 {
		totalWidth = options.totalWidth
	} else {
		totalWidth = helper.Width(synopsisString)
	}

	var arr string
	if FlagIsArray(flag) {
		arr = "<accent> (multiple values allowed)</accent>"
	}

	finalDescription := d.Output.Format(flag.GetDescription() + defaultValue + arr)
	wrappedDescriptionLines := helper.WrapToSlice(finalDescription, d.Output.lineLength-totalWidth, Eol, false)
	indent := strings.Repeat(" ", totalWidth+4)

	for i, line := range wrappedDescriptionLines {
		if i == 0 {
			wrappedDescriptionLines[i] = strings.Repeat(" ", totalWidth-helper.Width(synopsisString)) + line
		} else {
			wrappedDescriptionLines[i] = indent + line
		}
	}

	d.writeText(fmt.Sprintf("  <secondary>%s</secondary>  %s", synopsisString, strings.Join(wrappedDescriptionLines, "")))
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
	if len(command.Aliases) > 0 {
		text = fmt.Sprintf("[%s] ", strings.Join(command.Aliases, "|"))
	}

	return text
}

func (d *TextDescriptor) writeText(content string) {
	d.Write(content, true)
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
