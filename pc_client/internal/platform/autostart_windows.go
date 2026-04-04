//go:build windows

package platform

import (
	"os"
	"golang.org/x/sys/windows/registry"
)

func ToggleAutostart(enabled bool, name string) {
	execPath, _ := os.Executable()
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	if enabled {
		_ = k.SetStringValue(name, execPath)
	} else {
		_ = k.DeleteValue(name)
	}
}

func CheckAutostartStatus(name string) bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	
	_, _, err = k.GetStringValue(name)
	return err == nil
}
