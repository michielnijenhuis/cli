package cli

import (
	"fmt"
	"sort"
)

type NamespaceCommands struct {
	id       string
	commands []string
}

type ApplicationDescription struct {
	Application *Application
	Namespace   string
	ShowHidden  bool
	namespaces  map[string]*NamespaceCommands
	commands    map[string]*Command
	aliases     map[string]*Command
}

func (d *ApplicationDescription) Namespaces() map[string]*NamespaceCommands {
	if d.namespaces == nil {
		d.inspectApplication()
	}

	return d.namespaces
}

func (d *ApplicationDescription) Commands() map[string]*Command {
	if d.commands == nil {
		d.inspectApplication()
	}

	if d.commands == nil {
		return make(map[string]*Command)
	}

	return d.commands
}

func (d *ApplicationDescription) Command(name string) (*Command, error) {
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
	d.commands = make(map[string]*Command)
	d.namespaces = make(map[string]*NamespaceCommands)

	var all map[string]*Command
	if d.Namespace != "" {
		ns, _ := d.Application.FindNamespace(d.Namespace)
		all = d.Application.All(ns)
	} else {
		all = d.Application.All("")
	}

	for namespace, commands := range d.sortCommands(all) {
		names := make([]string, 0)

		for name, cmd := range commands {
			if cmd.Name == "" || (!d.ShowHidden && cmd.Hidden) {
				continue
			}

			if cmd.Name == name {
				d.commands[name] = cmd
			} else {
				if d.aliases == nil {
					d.aliases = make(map[string]*Command)
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

func (d *ApplicationDescription) sortCommands(commands map[string]*Command) map[string](map[string]*Command) {
	namespacedCommands := make(map[string]map[string]*Command)
	globalCommands := make(map[string]*Command)
	sortedCommands := make(map[string]map[string]*Command)
	globalNamespace := "_global"

	for name, cmd := range commands {
		key := d.Application.ExtractNamespace(name, -1)
		if key == "" || key == globalNamespace {
			globalCommands[name] = cmd
		} else {
			if namespacedCommands[key] == nil {
				namespacedCommands[key] = make(map[string]*Command)
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
	keys := make([]string, 0, len(obj))
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
