package selectmenu

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func NewSelectMenu(title string, options []huh.Option[string], selection *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(
					options...,
				).Value(selection))).WithTheme(huh.ThemeBase())
}

func ConstructOptions(input map[string]string) ([]huh.Option[string], error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("constructing options, no entries exist")
	}

	var options []huh.Option[string]
	for key, value := range input {
		newOption := huh.NewOption(key, value)
		options = append(options, newOption)
	}

	return options, nil
}
