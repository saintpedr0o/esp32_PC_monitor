//go:build linux

package platform

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

func SetupSystemTray(a fyne.App, w fyne.Window, name string) {
	if desk, ok := a.(desktop.App); ok {
		itemShow := fyne.NewMenuItem("Show", func() {
			w.Show()
		})

		menu := fyne.NewMenu(name, itemShow)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(theme.SettingsIcon())
	}
}
