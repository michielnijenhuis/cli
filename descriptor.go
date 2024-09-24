package cli

import (
	"fmt"
	"regexp"
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

func (d *TextDescriptor) DescribeCommand(command *Command, options *DescriptorOptions) {
	intro := command.GetHelp()
	if intro != "" {
		d.writeText(intro+Eol+Eol, options)
	}

	definition := command.Definition()

	d.writeText("<primary>Usage:</primary>", options)
	d.writeText(Eol, options)

	if command.Run != nil || command.RunE != nil {
		d.writeText("  "+command.Synopsis(true), options)

		if command.HasSubcommands() {
			d.writeText(Eol, options)
		}
	}

	if command.HasSubcommands() {
		if command.parent == nil {
			d.writeText(fmt.Sprintf("  %s [command] [flags] [--] [arguments]", command.FullName()), options)
		} else {
			d.writeText(fmt.Sprintf("  %s [command]", command.FullName()), options)
		}
	}

	ns := command.Namespace()
	if ns != "" {
		ns += " "
	}

	for _, alias := range command.Aliases {
		d.writeText(Eol, options)
		d.writeText(fmt.Sprintf("  %s%s", ns, alias), options)
	}

	d.writeText(Eol+Eol, options)

	d.DescribeInputDefinition(definition, options)

	commands := command.All()

	if len(commands) > 0 {
		d.writeText(Eol, nil)
		d.writeText(Eol, nil)
		d.writeText("<primary>Available commands:</primary>", options)

		commandNames := array.SortedKeys(commands)
		width := 0
		for _, name := range commandNames {
			width = max(width, len(name))
		}

		for _, name := range commandNames {
			cmd := commands[name]
			d.writeText(Eol, nil)
			spacingWidth := width - helper.Width(name)
			command := commands[name]

			var commandAliases string
			if name == command.Name {
				commandAliases = d.commandAliasesText(cmd)
			}

			nameParts := strings.Split(name, ":")
			name = strings.Join(nameParts, " ")

			d.writeText(fmt.Sprintf("  <accent>%s</accent>%s%s%s", name, strings.Repeat(" ", max(spacingWidth, 0)+2), commandAliases, command.Description), options)
		}
	}

	d.writeText(Eol, nil)

	help := command.ProcessedHelp()
	if help != "" && help != intro && help != command.Description {
		d.writeText(Eol, nil)
		d.writeText("<primary>Help:</primary>", options)
		d.writeText(Eol, nil)
		d.writeText("  "+strings.ReplaceAll(help, Eol, "\n  "), options)
		d.writeText(Eol, nil)
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
		d.writeText(Eol, nil)

		for _, argument := range definition.arguments {
			d.DescribeArgument(argument, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
			d.writeText(Eol, nil)
		}
	}

	if hasArgs && hasFlags {
		d.writeText(Eol, nil)
	}

	if hasFlags {
		laterFlags := make([]Flag, 0)

		d.writeText("<primary>Flags:</primary>", options)

		for _, flag := range definition.flags {
			if len(flag.GetShortcutString()) > 1 {
				laterFlags = append(laterFlags, flag)
				continue
			}

			d.writeText(Eol, nil)
			d.DescribeFlag(flag, &DescriptorOptions{
				namespace:  options.namespace,
				rawText:    options.rawText,
				short:      options.short,
				totalWidth: totalWidth,
			})
		}

		for _, flag := range laterFlags {
			d.writeText(Eol, nil)
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
