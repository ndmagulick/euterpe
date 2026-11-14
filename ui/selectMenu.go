package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func NewSelectMenu(title string, options map[string]string, selection *string) (*huh.Form, error) {
	menuOptions, err := newSelectMenuOptions(options)
	if err != nil {
		return nil, fmt.Errorf("building select menu: %w", err)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(
					menuOptions...,
				).Value(selection))).WithTheme(huh.ThemeBase()), nil
}

func newSelectMenuOptions(input map[string]string) ([]huh.Option[string], error) {
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
