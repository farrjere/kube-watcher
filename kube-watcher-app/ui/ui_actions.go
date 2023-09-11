package ui

import "github.com/ncruces/zenity"

func (u *UI) ChooseDir(title string) string {
	if title == "" {
		title = "Choose a directory"
	}

	file, err := zenity.SelectFileSave(zenity.Directory(), zenity.ConfirmOverwrite(), zenity.Title(title), zenity.Modal())

	if err != nil && err != zenity.ErrCanceled {
		zenity.Error("Error while saving directory", zenity.ErrorIcon)
	}

	return file
}

func (u *UI) OpenFile(title string) string {
	if title == "" {
		title = "Choose a file"
	}

	file, err := zenity.SelectFile(zenity.Title(title), zenity.Modal(), zenity.FileFilters{{Name: "CSV", Patterns: []string{"*.csv"}}})

	if err != nil && err != zenity.ErrCanceled {
		zenity.Error("Error while opening file", zenity.ErrorIcon)
	}

	return file
}
