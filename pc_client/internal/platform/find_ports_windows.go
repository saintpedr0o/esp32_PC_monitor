//go:build windows

package platform

import "fmt"

func GetSerialPorts() []string {
	var ports []string
	for i := 1; i <= 32; i++ {
		ports = append(ports, fmt.Sprintf("COM%d", i))
	}
	return ports
}
