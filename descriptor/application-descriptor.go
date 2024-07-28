package descriptor

import (
	"fmt"
	"sort"

	"github.com/michielnijenhuis/cli/command"
)

type NamespaceCommands struct {
	id       string
	commands []string
}

type ApplicationDescription struct {
	application DescribeableApplication
	namespace   string
	showHidden  bool
	namespaces  map[string]*NamespaceCommands
	commands    map[string]*command.Command
	aliases     map[string]*command.Command
}

func NewApplicationDescription(application DescribeableApplication, namespace string, showHidden bool) *ApplicationDescription {
	return &ApplicationDescription{
		application: application,
		namespace:   namespace,
		showHidden:  showHidden,
		namespaces:  nil,
		commands:    nil,
		aliases:     nil,
	}
}

func (d *ApplicationDescription) Namespaces() map[string]*NamespaceCommands {
	if d.namespaces == nil {
		d.inspectApplication()
	}

	return d.namespaces
}

func (d *ApplicationDescription) Commands() map[string]*command.Command {
	if d.commands == nil {
		d.inspectApplication()
	}

	if d.commands == nil {
		return make(map[string]*command.Command)
	}

	return d.commands
}

func (d *ApplicationDescription) Command(name string) (*command.Command, error) {
	hasCommand := d.commands != nil && d.commands[name] != nil
	hasAlias := d.aliases != nil && d.aliases[name] != nil

	if !hasCommand && !hasAlias {
		return nil, fmt.Errorf("command \"%s\" does not exist", name)
	}

	if hasCommand {
		return d.commands[name], nil
	}

	return d.aliases[name], nil
}

func (d *ApplicationDescription) inspectApplication() {
	d.commands = make(map[string]*command.Command)
	d.namespaces = make(map[string]*NamespaceCommands)

	var all map[string]*command.Command
	if d.namespace != "" {
		all = d.application.All(d.application.FindNamespaces(d.namespace))
	} else {
		all = d.application.All("")
	}

	for namespace, commands := range d.sortCommands(all) {
		names := make([]string, 0)

		for name, cmd := range commands {
			if cmd.GetName() == "" || (!d.showHidden && cmd.IsHidden()) {
				continue
			}

			if cmd.GetName() == name {
				d.commands[name] = cmd
			} else {
				if d.aliases == nil {
					d.aliases = make(map[string]*command.Command)
				}

				d.aliases[name] = cmd
			}

			names = append(names, name)
		}

		d.namespaces[namespace] = &NamespaceCommands{
			id:       namespace,
			commands: names,
		}
	}
}

func (d *ApplicationDescription) sortCommands(commands map[string]*command.Command) map[string](map[string]*command.Command) {
	namespacedCommands := make(map[string]map[string]*command.Command)
	globalCommands := make(map[string]*command.Command)
	sortedCommands := make(map[string]map[string]*command.Command)
	globalNamespace := "_global"

	for name, cmd := range commands {
		key := d.application.ExtractNamespace(name, -1)
		if key == "" || key == globalNamespace {
			globalCommands[name] = cmd
		} else {
			if namespacedCommands[key] == nil {
				namespacedCommands[key] = make(map[string]*command.Command)
			}

			namespacedCommands[key][name] = cmd
		}
	}

	if len(globalCommands) > 0 {
		globalCommands = ksort(globalCommands)
		sortedCommands[globalNamespace] = globalCommands
	}

	if len(namespacedCommands) > 0 {
		namespacedCommands = ksort(namespacedCommands)
		for key, commandsSet := range namespacedCommands {
			namespacedCommands[key] = ksort(commandsSet)
			sortedCommands[key] = commandsSet
		}
	}

	return sortedCommands
}

func ksort[T any](obj map[string]T) map[string]T {
	keys := make([]string, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedObj := make(map[string]T)

	for k := range sortedObj {
		sortedObj[k] = obj[k]
	}

	return obj
}
