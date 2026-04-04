//go:build linux

package platform

import "fmt"

func GetSerialPorts() []string {
	var ports []string
	for i := 0; i <= 9; i++ {
		ports = append(ports, fmt.Sprintf("/dev/rfcomm%d", i))
	}
	return ports
}
