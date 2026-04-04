//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

func ToggleAutostart(enabled bool, name string) {
	execPath, _ := os.Executable()
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "autostart")
	file := filepath.Join(dir, name+".desktop")

	if enabled {
		_ = os.MkdirAll(dir, 0755)
		data := fmt.Sprintf("[Desktop Entry]\nName=%s\nExec=%s\n", name, execPath)
		_ = os.WriteFile(file, []byte(data), 0644)
	} else {
		_ = os.Remove(file)
	}
}

func CheckAutostartStatus(name string) bool {
    home, _ := os.UserHomeDir()
    file := filepath.Join(home, ".config", "autostart", name+".desktop")
    _, err := os.Stat(file)
    return err == nil
}
