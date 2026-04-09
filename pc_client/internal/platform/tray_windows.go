//go:build windows

package platform

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func SetupSystemTray(a fyne.App, w fyne.Window, name string, icon fyne.Resource) {
	if desk, ok := a.(desktop.App); ok {
		itemShow := fyne.NewMenuItem("Show", func() {
			w.Show()
		})

		itemQuit := fyne.NewMenuItem("Quit", func() {
			a.Quit()
		})
        itemQuit.IsQuit = true
		menu := fyne.NewMenu(name, 
			itemShow, 
			fyne.NewMenuItemSeparator(), 
			itemQuit,
		)
		
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(icon)
	}
}
