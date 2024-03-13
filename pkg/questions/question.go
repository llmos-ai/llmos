package questions

import (
	"github.com/erikgeiser/promptkit/textinput"
)

func Prompt(prompt, initialValue, placeHolder string, empty, hidden bool) (string, error) {
	input := textinput.New(prompt)
	input.InitialValue = initialValue
	input.Placeholder = placeHolder
	if empty {
		input.Validate = func(s string) error { return nil }
	}
	input.Hidden = hidden

	return input.RunPrompt()
}
