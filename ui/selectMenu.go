package ui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

type SelectMenuStringPairing struct {
	Option string
	Value  string
}

func NewSelectMenu(title string, options []SelectMenuStringPairing, selection *string) (*huh.Form, error) {
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

func newSelectMenuOptions(inputs []SelectMenuStringPairing) ([]huh.Option[string], error) {
	if len(inputs) == 0 {
		return nil, errors.New("constructing options, no entries exist")
	}

	var options []huh.Option[string]
	for _, input := range inputs {
		newOption := huh.NewOption(input.Option, input.Value)
		options = append(options, newOption)
	}

	return options, nil
}
