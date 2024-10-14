package cli

import "errors"

const (
	FormText = iota
	FormTextArea
	FormArray
	FormSelect
	FormMultiselect
	FormConfirm
	FormWait
	FormPause
)

type FormStep struct {
	Type        uint
	Label       string
	Placeholder string
	String      *string
	Required    bool
	Bool        *bool
	Array       *[]string
	Values      []string
	Labels      []string
	WaitFunc    func()

	prev *FormStep
	next *FormStep
}

type form struct {
	items   *FormStep
	current *FormStep
	i       *Input
	o       *Output
}

func CreateForm(i *Input, o *Output, first *FormStep, steps ...*FormStep) form {
	if len(steps) > 0 {
		first.next = steps[0]
	}

	for i := 0; i < len(steps); i++ {
		cur := steps[i]
		if i+1 < len(steps) {
			cur.next = steps[i+1]
		}
		if i > 0 {
			cur.prev = steps[i-1]
		} else {
			cur.prev = first
		}
	}

	return form{
		items:   first,
		current: first,
		i:       i,
		o:       o,
	}
}

func (f form) Submit() error {
	for cur := f.current; cur != nil; cur = cur.next {
		switch cur.Type {
		case FormText:
			if cur.String == nil {
				return errors.New("form type FormText expects a string pointer")
			}
			prompt := NewTextPrompt(f.i, f.o, cur.Label, *cur.String)
			prompt.Placeholder = cur.Placeholder
			prompt.Required = cur.Required
			answer, err := prompt.Render()
			if err != nil {
				return err
			}
			*cur.String = answer
		case FormTextArea:
			return errors.New("form type not yet implemented")
		case FormArray:
			if cur.Array == nil {
				return errors.New("form type FormText expects a string pointer")
			}
			prompt := NewArrayPrompt(f.i, f.o, cur.Label, *cur.Array)
			prompt.Required = cur.Required
			answer, err := prompt.Render()
			if err != nil {
				return err
			}
			*cur.Array = answer
		case FormSelect:
			if cur.String == nil {
				return errors.New("form type FormSelect expects a string pointer")
			}
			if cur.Values == nil || len(cur.Values) < 2 {
				return errors.New("form type FormSelect expects at least two choices")
			}
			prompt := NewSelectPrompt(f.i, f.o, cur.Label, cur.Values, cur.Labels, *cur.String)
			prompt.Required = cur.Required
			answer, err := prompt.Render()
			if err != nil {
				return err
			}
			*cur.String = answer
		case FormMultiselect:
			return errors.New("form type not yet implemented")
		case FormConfirm:
			if cur.Bool == nil {
				return errors.New("form type FormText expects a bool pointer")
			}
			prompt := NewConfirmPrompt(f.i, f.o, cur.Label, *cur.Bool)
			prompt.Required = cur.Required
			answer, err := prompt.Render()
			if err != nil {
				return err
			}
			*cur.Bool = answer
		case FormWait:
			if cur.WaitFunc == nil {
				return errors.New("form type FormWait expects a WaitFunc")
			}

			spinner := NewSpinner(f.i, f.o, cur.Label, nil, "")
			spinner.Spin(cur.WaitFunc)
		case FormPause:
			prompt := NewPausePrompt(f.i, f.o, cur.Label)
			shouldContinue, err := prompt.Render()
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
		default:
			return errors.New("unsupported form type")
		}
	}

	return nil
}
