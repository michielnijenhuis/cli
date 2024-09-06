package cli

import (
	"slices"
	"strings"
)

type Flag interface {
	GetName() string
	GetShortcuts() []string
	GetShortcutString() string
	GetDescription() string
	HasValue() bool
}

type StringFlag struct {
	Name        string
	Shortcuts   []string
	Description string
	Value       string
	Validator   func(string) error
}

type BoolFlag struct {
	Name        string
	Shortcuts   []string
	Description string
	Value       bool
	Negatable   bool
	Validator   func(bool) error
}

type ArrayFlag struct {
	Name        string
	Shortcuts   []string
	Description string
	Value       []string
	Validator   func([]string) error
}

type OptionalStringFlag struct {
	Name        string
	Shortcuts   []string
	Description string
	Boolean     bool
	Value       string
	Validator   func(bool, string) error
}

type OptionalArrayFlag struct {
	Name        string
	Shortcuts   []string
	Description string
	Boolean     bool
	Value       []string
	Validator   func(bool, []string) error
}

func SetFlagValue(f Flag, str string, boolean bool) {
	switch flag := f.(type) {
	case *StringFlag:
		flag.Value = str
	case *BoolFlag:
		flag.Value = boolean
	case *ArrayFlag:
		if flag.Value == nil {
			flag.Value = make([]string, 0)
		}
		flag.Value = append(flag.Value, str)
	case *OptionalStringFlag:
		flag.Value = str
		flag.Boolean = boolean
	case *OptionalArrayFlag:
		if flag.Value == nil {
			flag.Value = make([]string, 0)
		}
		flag.Value = append(flag.Value, str)
		flag.Boolean = boolean
	default:
		panic("invalid flag type")
	}
}

func GetFlagStringValue(f any) string {
	switch flag := f.(type) {
	case *StringFlag:
		return flag.Value
	case *OptionalStringFlag:
		return flag.Value
	default:
		return ""
	}
}

func GetFlagBoolValue(f any) bool {
	switch flag := f.(type) {
	case *BoolFlag:
		return flag.Value
	case *OptionalStringFlag:
		return flag.Boolean || flag.Value != ""
	case *OptionalArrayFlag:
		return flag.Boolean || len(flag.Value) > 0
	default:
		return false
	}
}

func GetFlagArrayValue(f any) []string {
	switch flag := f.(type) {
	case *ArrayFlag:
		arr := flag.Value
		if arr == nil {
			return []string{}
		}
		return arr
	case *OptionalArrayFlag:
		arr := flag.Value
		if arr == nil {
			return []string{}
		}
		return arr
	default:
		return []string{}
	}
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) GetShortcuts() []string {
	return f.Shortcuts
}

func joinShortcuts(shortcuts []string) string {
	return strings.Join(shortcuts, "|")
}

func (f *StringFlag) GetShortcutString() string {
	return joinShortcuts(f.Shortcuts)
}

func (f *StringFlag) GetDescription() string {
	return f.Description
}

func (f *StringFlag) HasValue() bool {
	return f.Value != ""
}

func (f *BoolFlag) GetName() string {
	return f.Name
}

func (f *BoolFlag) GetShortcuts() []string {
	return f.Shortcuts
}

func (f *BoolFlag) GetShortcutString() string {
	return joinShortcuts(f.Shortcuts)
}

func (f *BoolFlag) GetDescription() string {
	return f.Description
}

func (f *BoolFlag) HasValue() bool {
	return f.Value
}

func (f *ArrayFlag) GetName() string {
	return f.Name
}

func (f *ArrayFlag) GetShortcuts() []string {
	return f.Shortcuts
}

func (f *ArrayFlag) GetShortcutString() string {
	return joinShortcuts(f.Shortcuts)
}

func (f *ArrayFlag) GetDescription() string {
	return f.Description
}

func (f *ArrayFlag) HasValue() bool {
	return len(f.Value) > 0
}

func (f *OptionalStringFlag) GetName() string {
	return f.Name
}

func (f *OptionalStringFlag) GetShortcuts() []string {
	return f.Shortcuts
}

func (f *OptionalStringFlag) GetShortcutString() string {
	return joinShortcuts(f.Shortcuts)
}

func (f *OptionalStringFlag) HasValue() bool {
	return f.Boolean || f.Value != ""
}

func (f *OptionalStringFlag) GetDescription() string {
	return f.Description
}

func (f *OptionalArrayFlag) GetName() string {
	return f.Name
}

func (f *OptionalArrayFlag) GetShortcuts() []string {
	return f.Shortcuts
}

func (f *OptionalArrayFlag) GetShortcutString() string {
	return joinShortcuts(f.Shortcuts)
}

func (f *OptionalArrayFlag) GetDescription() string {
	return f.Description
}

func (f *OptionalArrayFlag) HasValue() bool {
	return f.Boolean || len(f.Value) > 0
}

const (
	flagTypeString = iota
	flagTypeBool
	flagTypeArray
	flagTypeOptionalString
	flagTypeOptionalArray
)

func FlagType(f Flag) uint {
	switch f.(type) {
	case *StringFlag:
		return flagTypeString
	case *BoolFlag:
		return flagTypeBool
	case *ArrayFlag:
		return flagTypeArray
	case *OptionalStringFlag:
		return flagTypeOptionalString
	case *OptionalArrayFlag:
		return flagTypeOptionalArray
	default:
		panic("unknown flag type")
	}
}

func FlagEquals(f1 Flag, f2 Flag) bool {
	if f1.GetName() != f2.GetName() {
		return false
	}

	if !slices.Equal(f1.GetShortcuts(), f2.GetShortcuts()) {
		return false
	}

	if f1.GetDescription() != f2.GetDescription() {
		return false
	}

	f1Type := FlagType(f1)
	f2Type := FlagType(f2)

	if f1Type != f2Type {
		return false
	}

	switch f1Type {
	case flagTypeString:
		return GetFlagStringValue(f1) == GetFlagStringValue(f2)
	case flagTypeBool:
		return GetFlagBoolValue(f1) == GetFlagBoolValue(f2)
	case flagTypeArray:
		return slices.Equal(GetFlagArrayValue(f1), GetFlagArrayValue(f2))
	case flagTypeOptionalString:
		return GetFlagBoolValue(f1) == GetFlagBoolValue(f2) && GetFlagStringValue(f1) == GetFlagStringValue(f2)
	case flagTypeOptionalArray:
		return GetFlagBoolValue(f1) == GetFlagBoolValue(f2) && slices.Equal(GetFlagArrayValue(f1), GetFlagArrayValue(f2))
	default:
		panic("unknown flag type")
	}
}

func FlagIsNegatable(f Flag) bool {
	b, ok := f.(*BoolFlag)
	if !ok {
		return false
	}
	return b.Negatable
}

func FlagIsArray(f Flag) bool {
	t := FlagType(f)
	return t == flagTypeArray || t == flagTypeOptionalArray
}

func FlagAcceptsValue(f Flag) bool {
	t := FlagType(f)
	return t == flagTypeString || t == flagTypeArray || t == flagTypeOptionalString || t == flagTypeOptionalArray
}

func FlagRequiresValue(f Flag) bool {
	t := FlagType(f)
	return t == flagTypeString || t == flagTypeArray
}

func FlagValueIsOptional(f Flag) bool {
	t := FlagType(f)
	return t == flagTypeOptionalString || t == flagTypeOptionalArray
}

func FlagHasDefaultValue(f Flag) bool {
	switch t := f.(type) {
	case *StringFlag:
		return t.Value != ""
	case *BoolFlag:
		return t.Value
	case *ArrayFlag:
		return len(t.Value) > 0
	case *OptionalStringFlag:
		return t.Boolean || t.Value != ""
	case *OptionalArrayFlag:
		return t.Boolean || len(t.Value) > 0
	default:
		return false
	}
}

func ValidateFlag(f Flag) error {
	switch t := f.(type) {
	case *StringFlag:
		if t.Validator != nil {
			return t.Validator(t.Value)
		}
		return nil
	case *BoolFlag:
		if t.Validator != nil {
			return t.Validator(t.Value)
		}
		return nil
	case *ArrayFlag:
		if t.Validator != nil {
			return t.Validator(t.Value)
		}
		return nil
	case *OptionalStringFlag:
		if t.Validator != nil {
			return t.Validator(t.Boolean, t.Value)
		}
		return nil
	case *OptionalArrayFlag:
		if t.Validator != nil {
			return t.Validator(t.Boolean, t.Value)
		}
		return nil
	default:
		panic("invalid flag type")
	}
}

func ArgSetValue(arg Arg, token string) {
	switch a := arg.(type) {
	case *StringArg:
		a.Value = token
	case *ArrayArg:
		if a.Value == nil {
			a.Value = make([]string, 0)
		}
		a.Value = append(a.Value, token)
	default:
		panic("invalid argument type")
	}
}
