// modified version of completion.go from Cobra

// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	ShellCompRequestCmd = "__complete"
)

type ShellCompDirective int

const (
	// ShellCompDirectiveError indicates an error occurred and completions should be ignored.
	ShellCompDirectiveError ShellCompDirective = 1 << iota

	// ShellCompDirectiveNoSpace indicates that the shell should not add a space
	// after the completion even if there is a single completion provided.
	ShellCompDirectiveNoSpace

	// ShellCompDirectiveNoFileComp indicates that the shell should not provide
	// file completion even when no completion is provided.
	ShellCompDirectiveNoFileComp

	// ShellCompDirectiveFilterFileExt indicates that the provided completions
	// should be used as file extension filters.
	// For flags, using Command.MarkFlagFilename() and Command.MarkPersistentFlagFilename()
	// is a shortcut to using this directive explicitly.  The BashCompFilenameExt
	// annotation can also be used to obtain the same behavior for flags.
	ShellCompDirectiveFilterFileExt

	// ShellCompDirectiveFilterDirs indicates that only directory names should
	// be provided in file completion.  To request directory names within another
	// directory, the returned completions should specify the directory within
	// which to search.  The BashCompSubdirsInDir annotation can be used to
	// obtain the same behavior but only for flags.
	ShellCompDirectiveFilterDirs

	// ShellCompDirectiveKeepOrder indicates that the shell should preserve the order
	// in which the completions are provided
	ShellCompDirectiveKeepOrder

	// ===========================================================================

	// All directives using iota should be above this one.
	// For internal use.
	shellCompDirectiveMaxValue

	// ShellCompDirectiveDefault indicates to let the shell perform its default
	// behavior after completions have been provided.
	// This one must be last to avoid messing up the iota count.
	ShellCompDirectiveDefault ShellCompDirective = 0
)

func (d ShellCompDirective) string() string {
	var directives []string
	if d&ShellCompDirectiveError != 0 {
		directives = append(directives, "ShellCompDirectiveError")
	}
	if d&ShellCompDirectiveNoSpace != 0 {
		directives = append(directives, "ShellCompDirectiveNoSpace")
	}
	if d&ShellCompDirectiveNoFileComp != 0 {
		directives = append(directives, "ShellCompDirectiveNoFileComp")
	}
	if d&ShellCompDirectiveFilterFileExt != 0 {
		directives = append(directives, "ShellCompDirectiveFilterFileExt")
	}
	if d&ShellCompDirectiveFilterDirs != 0 {
		directives = append(directives, "ShellCompDirectiveFilterDirs")
	}
	if d&ShellCompDirectiveKeepOrder != 0 {
		directives = append(directives, "ShellCompDirectiveKeepOrder")
	}
	if len(directives) == 0 {
		directives = append(directives, "ShellCompDirectiveDefault")
	}

	if d >= shellCompDirectiveMaxValue {
		return fmt.Sprintf("ERROR: unexpected ShellCompDirective value: %d", d)
	}
	return strings.Join(directives, ", ")
}

func (c *Command) initCompleteCmd(args []string) {
	completeCmd := &Command{
		Name:   ShellCompRequestCmd,
		Hidden: true,
		Arguments: []Arg{
			&ArrayArg{
				Name:        "command-line",
				Description: "The command line to request completions for.",
				Min:         1,
			},
		},
		Description: "Request shell completion choices for the specified command-line",
		Help: fmt.Sprintf("%[2]s is a special command that is used by the shell completion logic\n%[1]s",
			"to request completion choices for the specified command-line.", ShellCompRequestCmd),
		Run: func(io *IO) {
			cmd := io.Command
			_, completions, directive, err := cmd.getCompletions(io.Input, args)
			if err != nil {
				CompErrorln(io.Output.Formatter().RemoveDecoration(StripEscapeSequences(err.Error())))
			}

			out := io.Output.Stream
			for _, comp := range completions {
				comp = strings.SplitN(comp, "\n", 2)[0]
				comp = strings.TrimSpace(comp)
				fmt.Fprintln(out, comp)
			}

			fmt.Fprintf(out, ":%d\n", directive)
			fmt.Fprintf(io.Output.Stderr.Stream, "Completion ended with directive: %s\n", directive.string())
		},
	}

	c.AddCommand(completeCmd)
}

func (c *Command) getCompletions(i *Input, args []string) (finalCmd *Command, completions []string, directive ShellCompDirective, err error) {
	root := c.Root()
	root.init()

	toComplete := args[len(args)-1] // TODO: use (show flags, or apply search)
	trimmedArgs := args[:len(args)-1]

	flagCompletion := false
	if strings.HasPrefix(toComplete, "-") && (len(trimmedArgs) == 0 || trimmedArgs[len(trimmedArgs)-1] != "--") {
		flagCompletion = true
	}

	tokens := make([]string, 0, len(trimmedArgs))
	for _, arg := range trimmedArgs {
		if strings.TrimSpace(arg) != "" && arg != "__complete" {
			tokens = append(tokens, arg)
		}
	}

	finalCmd, _, err = root.findCommand(tokens, &tokens)
	if err != nil {
		return
	}

	finalCmd.init()

	var definition *InputDefinition
	definition, err = finalCmd.Definition()
	if err != nil {
		return
	}

	i.Bind(definition)
	inspection, _ := i.Inspect(tokens)

	if (finalCmd.hasFlag(i, "help") && (inspection.Flags["help"] != nil || inspection.Flags["h"] != nil)) ||
		(finalCmd.hasFlag(i, "version") && (inspection.Flags["version"] != nil || inspection.Flags["V"] != nil)) {
		completions = []string{}
		directive = ShellCompDirectiveNoFileComp
		return
	}

	directive = ShellCompDirectiveDefault

	// TODO: fix - if --help flag is present, completion will use those results always.
	// TODO: chaining short flags (e.g. `-abc`, where a, b and c are short flags)
	if flagCompletion {
		includeShort := strings.HasPrefix(toComplete, "-") && !strings.HasPrefix(toComplete, "--")
		isShort := includeShort

		for strings.HasPrefix(toComplete, "-") {
			toComplete = strings.TrimPrefix(toComplete, "-")
		}

		hasValue := false
		if strings.HasSuffix(toComplete, "=") {
			hasValue = true
			toComplete = strings.TrimSuffix(toComplete, "=")
		}

		flag, _ := definition.Flag(toComplete)
		if flag != nil {
			if (hasValue || len(toComplete) == 1) && (FlagRequiresValue(flag) || FlagValueIsOptional(flag)) {
				opts, ok := flag.(HasOptions)
				if !ok {
					return
				}

				completions = opts.Opts()
				directive = ShellCompDirectiveDefault
			} else if isShort {
				completions = make([]string, 0)
				for _, f := range definition.flags {
					if inspection.FlagIsGiven(f) {
						continue
					}

					for _, short := range f.GetShortcuts() {
						completions = append(completions, fmt.Sprintf("%s\t%s", short, f.GetDescription()))
					}

					directive = ShellCompDirectiveNoSpace
				}
			}

			return
		}

		directive = ShellCompDirectiveNoFileComp
		completions = make([]string, 0)

		for _, f := range definition.flags {
			if inspection.FlagIsGiven(f) {
				continue
			}

			if toComplete != "" && !strings.HasPrefix(f.GetName(), toComplete) {
				continue
			}

			var suffix string
			if FlagRequiresValue(f) {
				suffix = "="
				directive = ShellCompDirectiveNoSpace
			}

			completions = append(completions, fmt.Sprintf("--%s%s\t%s", f.GetName(), suffix, f.GetDescription()))

			if includeShort {
				for _, short := range f.GetShortcuts() {
					completions = append(completions, fmt.Sprintf("-%s\t%s", short, f.GetDescription()))
				}
			}
		}

		return
	}

	if len(finalCmd.Arguments) > 0 {
		j := 0
		for i := 0; i < len(finalCmd.Arguments); i++ {
			arg := finalCmd.Arguments[i]

			switch arg.(type) {
			case *StringArg:
				if j >= len(inspection.Args) {
					if len(arg.Opts()) > 0 {
						completions = append(completions, arg.Opts()...)
						return
					}

					if arg.IsRequired() {
						return
					}
				} else {
					j++
				}
			case *ArrayArg:
				if opts := arg.Opts(); len(opts) > 0 {
					availableOptions := make([]string, 0, len(opts))
					for _, opt := range opts {
						if !slices.Contains(inspection.Args, opt) {
							availableOptions = append(availableOptions, opt)
						}
					}

					if len(availableOptions) == 0 {
						directive = ShellCompDirectiveNoFileComp
					} else {
						completions = availableOptions
					}
				} else if arg.IsRequired() {
					return
				}
			default:
				return
			}
		}
	}

	if finalCmd.HasSubcommands() {
		completions = make([]string, 0)

		for _, cmd := range finalCmd.Subcommands() {
			if !cmd.Hidden {
				completions = append(completions, fmt.Sprintf("%s\t%s", cmd.Name, cmd.Description))
			}
		}
	} else if len(finalCmd.Arguments) == 0 || len(inspection.Args) >= len(finalCmd.Arguments) {
		directive = ShellCompDirectiveNoFileComp
		return
	}

	return
}

func (c *Command) InitDefaultCompletionCmd(out *os.File) {
	if !c.HasSubcommands() {
		return
	}

	for _, cmd := range c.commands {
		if cmd.Name == "completion" || slices.Contains(cmd.Aliases, "completion") {
			return
		}
	}

	completionCmd := &Command{
		Name:        "completion",
		Description: "Generate the autocompletion script for the specified shell",
		Help: fmt.Sprintf(`Generate the autocompletion script for %[1]s for the specified shell.
See each sub-command's help for details on how to use the generated script.
`, c.Root().Name),
	}

	c.AddCommand(completionCmd)

	shortDesc := "Generate the autocompletion script for %s"

	zsh := &Command{
		Name:        "zsh",
		Description: fmt.Sprintf(shortDesc, "zsh"),
		Help: fmt.Sprintf(`Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(%[1]s completion zsh)

To load completions for every new session, execute once:

#### Linux:

	%[1]s completion zsh > "${fpath[1]}/_%[1]s"

#### macOS:

	%[1]s completion zsh > $(brew --prefix)/share/zsh/site-functions/_%[1]s

You will need to start a new shell for this setup to take effect.
`, c.Root().Name),
		RunE: func(io *IO) error {
			return io.Command.Root().GenZshCompletion(out)
		},
	}

	completionCmd.AddCommand(zsh)
}

func CompError(msg string) {
	msg = fmt.Sprintf("[Error] %s", msg)
	fmt.Fprint(os.Stderr, msg)
}

func CompErrorln(msg string) {
	CompError(fmt.Sprintf("%s\n", msg))
}
