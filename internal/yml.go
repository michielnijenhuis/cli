package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/yaml/go-yaml"
)

type yamlFlag struct {
	Aliases     []string `yaml:"aliases"`
	Description string   `yaml:"description"`
	Default     any      `yaml:"default"`
	Negatable   bool     `yaml:"negatable"`
	Options     []string `yaml:"options"`
}

type yamlFlagMap map[string]yamlFlag

type yamlArg struct {
	Description string   `yaml:"description"`
	Value       string   `yaml:"value"`
	Required    bool     `yaml:"required"`
	Options     []string `yaml:"options"`
}

type yamlArgMap map[string]yamlArg

type yamlCommandMap map[string]yamlCommand

type yamlCommand struct {
	Aliases            []string       `yaml:"aliases"`
	Description        string         `yaml:"description"`
	Help               string         `yaml:"help"`
	Flags              yamlFlagMap    `yaml:"flags"`
	Args               yamlArgMap     `yaml:"args"`
	Commands           yamlCommandMap `yaml:"commands"`
	NoNativeFlags      bool           `yaml:"noNativeFlags"`
	WithoutNativeFlags []string       `yaml:"withoutNativeFlags"`
}

type yamlTheme struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Accent    string `yaml:"accent"`
}

type yamlApplication struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	LongVersion string    `yaml:"longVersion"`
	Intro       string    `yaml:"intro"`
	Theme       yamlTheme `yaml:"theme"`
	yamlCommand `yaml:",inline"`
}

type Payload struct {
	Args  map[string]any
	Flags map[string]any
}

func ParseApp(config string) (*yamlApplication, error) {
	fileInfo, err := os.Stat(config)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path [%s] is a directory, not a file path", config)
	}

	file, err := os.Open(config)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var yamlApp yamlApplication
	err = yaml.NewDecoder(file).Decode(&yamlApp)
	if err != nil {
		return nil, err
	}

	return &yamlApp, nil
}

func Parse(fileName string) (*Command, error) {
	yamlApp, err := ParseApp(fileName)
	if err != nil {
		return nil, err
	}

	root, err := toCommand(yamlApp.Name, yamlApp.yamlCommand, nil)
	if err != nil {
		return nil, err
	}

	if yamlApp.Version != "" {
		root.Version = yamlApp.Version
	}

	if yamlApp.LongVersion != "" {
		root.LongVersion = yamlApp.LongVersion
	}

	if yamlApp.Intro != "" {
		root.Intro = yamlApp.Intro
	}

	root.AutoExit = true

	SetBaseTheme(yamlApp.Theme.Primary, yamlApp.Theme.Secondary, yamlApp.Theme.Accent)

	return root, err
}

func toCommand(name string, c yamlCommand, parent *Command) (*Command, error) {
	cmd := &Command{
		Name:               name,
		Aliases:            c.Aliases,
		Description:        c.Description,
		Help:               c.Help,
		NoNativeFlags:      c.NoNativeFlags,
		WithoutNativeFlags: c.WithoutNativeFlags,
	}

	if len(c.Flags) > 0 {
		for name, flag := range c.Flags {
			var f Flag
			switch value := flag.Default.(type) {
			case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				f = &StringFlag{
					Name:        name,
					Shortcuts:   flag.Aliases,
					Description: flag.Description,
					Value:       fmt.Sprintf("%v", value),
					Options:     flag.Options,
				}

				err := cmd.AddFlag(f)
				if err != nil {
					return nil, err
				}
			case bool, nil:
				boolean := false
				if value == true {
					boolean = true
				}

				f = &BoolFlag{
					Name:        name,
					Shortcuts:   flag.Aliases,
					Description: flag.Description,
					Negatable:   flag.Negatable,
					Value:       boolean,
				}

				err := cmd.AddFlag(f)

				if err != nil {
					return nil, err
				}
			}
		}
	}

	if len(c.Args) > 0 {
		for name, arg := range c.Args {
			err := cmd.AddArg(&StringArg{
				Name:        name,
				Description: arg.Description,
				Required:    arg.Required,
				Value:       arg.Value,
				Options:     arg.Options,
			})

			if err != nil {
				return nil, err
			}
		}
	}

	if len(c.Commands) > 0 {
		for name, yamlCmd := range c.Commands {
			subCmd, err := toCommand(name, yamlCmd, cmd)
			if err != nil {
				return nil, err
			}

			err = cmd.AddCommand(subCmd)
			if err != nil {
				return nil, err
			}
		}
	}

	cmd.SetParent(parent)

	cmd.RunE = func(args []string, ctx *CommandCtx) error {
		// step 1: get the fn name to run
		var fnName string

		path := make([]string, 0)
		cur := cmd
		depth := 0

		for cur != nil {
			if cur.parent != nil {
				path = append(path, cur.Name)
				cur = cur.parent
				depth++
			} else {
				break
			}
		}

		if len(path) == 0 && !ctx.Command.HasSubcommands() {
			fnName = "main"
		} else if len(path) > 0 {
			slices.Reverse(path)
			fnName = strings.Join(path, "_")
		}

		// step 2: parse input
		definition, err := cmd.Definition()
		if err != nil {
			ctx.IO.Critical(err)
			os.Exit(1)
		}

		forwardedArgs := make([]string, 0, len(args))
		for _, arg := range args {
			cur := ctx.Command
			isCommandArg := false
			for cur != nil && !isCommandArg {
				if cur.Name == arg {
					isCommandArg = true
				} else {
					isCommandArg = slices.Contains(cur.Aliases, arg)
				}

				cur = cur.parent
			}

			if !isCommandArg {
				forwardedArgs = append(forwardedArgs, arg)
			}
		}

		input := NewInput(forwardedArgs...)
		err = input.Bind(definition)
		if err != nil {
			return err
		}

		// step 3: serialize parsed input
		flags := definition.GetFlags()
		arguments := definition.GetArguments()

		p := Payload{
			Args:  make(map[string]any),
			Flags: make(map[string]any),
		}

		for _, arg := range arguments {
			p.Args[arg.GetName()] = GetArgStringValue(arg)
		}

		for _, flag := range flags {
			switch flag.(type) {
			case *StringFlag:
				p.Flags[flag.GetName()] = GetFlagStringValue(flag)
			case *BoolFlag:
				p.Flags[flag.GetName()] = GetFlagBoolValue(flag)
			}
		}

		serializedPayload, err := json.Marshal(p)
		if err != nil {
			return err
		}

		payloadString := base64.StdEncoding.EncodeToString(serializedPayload)

		// step 4: get cmd source
		config := os.Getenv("CLI_CMD_CONFIG")
		if config == "" {
			return errors.New("CLI_CMD_CONFIG is not set")
		}

		script := os.Getenv("CLI_CMD_SCRIPT")
		if script == "" {
			return errors.New("CLI_CMD_SCRIPT is not set")
		}

		// step 5: set context
		fg, err := GetTheme("primary")
		if err != nil {
			return err
		} else {
			os.Setenv("CLI_THEME_PRIMARY", fg.Foreground)
		}

		bg, err := GetTheme("secondary")
		if err != nil {
			return err
		} else {
			os.Setenv("CLI_THEME_SECONDARY", bg.Foreground)
		}

		accent, err := GetTheme("accent")
		if err != nil {
			return err
		} else {
			os.Setenv("CLI_THEME_ACCENT", accent.Foreground)
		}

		// step 5: run the fn
		if fnName != "" {
			os.Setenv("CLI_PAYLOAD", payloadString)

			err = ctx.IO.Exec(
				"zsh",
				"-c",
				fmt.Sprintf("sh %s %s", script, fnName),
			)

			if err != nil {
				isExitError := strings.Contains(err.Error(), "exit status")
				if isExitError {
					os.Exit(1)
				}
			}

			return err
		} else {
			ctx.Command.printHelp(ctx.IO.Output, nil, "")

			return nil
		}
	}

	return cmd, nil
}
