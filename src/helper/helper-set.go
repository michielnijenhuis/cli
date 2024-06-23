package helper

import (
	"fmt"

	err "github.com/michielnijenhuis/cli/error"
)

type HelperSet struct {
	helpers map[string]HelperInterface
}

func NewHelperSet(helpers []HelperInterface) *HelperSet {
	hs := &HelperSet{
		helpers: make(map[string]HelperInterface),
	}

	for _, h := range helpers {
		hs.Set(h, "")
	}

	return hs
}

func (hs *HelperSet) Set(helper HelperInterface, alias string) {
	hs.helpers[helper.GetName()] = helper

	if alias != "" {
		hs.helpers[alias] = helper
	}

	helper.SetHelperSet(hs)
}

func (hs *HelperSet) Has(name string) bool {
	return hs.helpers[name] != nil
}

func (hs *HelperSet) Get(name string) (HelperInterface, error) {
	if !hs.Has(name) {
		return nil, err.NewInvalidArgumentError(fmt.Sprintf("The helper \"%s\" is not defined.", name))
	}

	return hs.helpers[name], nil
}

func (hs *HelperSet) Iterate() []HelperInterface {
	sets := make([]HelperInterface, len(hs.helpers))
	for _, h := range hs.helpers {
		sets = append(sets, h)
	}
	return sets
}
