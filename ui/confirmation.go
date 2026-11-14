package ui

import "github.com/charmbracelet/huh"

func NewConfirmation(title, description string, selection *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(selection),
		),
	).WithTheme(huh.ThemeBase())
}
