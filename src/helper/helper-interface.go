package helper

type HelperInterface interface {
	SetHelperSet(helperSet *HelperSet)
	GetHelperSet() HelperSet
	GetName() string
}
