package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/tarm/serial"
)

func getPossiblePorts() []string {
	var ports []string
	switch runtime.GOOS {
	case "windows":
		for i := 1; i <= 32; i++ {
			ports = append(ports, fmt.Sprintf("COM%d", i))
		}
	case "darwin":
		files, _ := os.ReadDir("/dev")
		for _, f := range files {
			name := f.Name()
			if strings.HasPrefix(name, "tty.") && (strings.Contains(name, "Bluetooth") || strings.Contains(name, "ESP")) {
				ports = append(ports, "/dev/"+name)
			}
		}
	case "linux":
		for i := 0; i <= 9; i++ {
			ports = append(ports, fmt.Sprintf("/dev/rfcomm%d", i))
		}
	}
	return ports
}

func main() {
	// auto reconnect
	for {
		port, activePort := findDevice()
		if port == nil {
			continue
		}

		fmt.Printf("\n[OK] Connected to: %s\n", activePort)
		
		for {
			data, err := collectStats(activePort)
			if err != nil {
				fmt.Println("\n[!] Error collecting stats:", err)
				break 
			}

			_, err = port.Write([]byte(data))
			if err != nil {
				fmt.Printf("\n[!] Port %s failed. Reconnecting...\n", activePort)
				break
			}
			time.Sleep(1 * time.Second)
		}
		port.Close()
		time.Sleep(2 * time.Second)
	}
}

func findDevice() (*serial.Port, string) {
	fmt.Println("\n=== Searching for ESP32 Monitor ===")
	dots := ""
	for {
		dots += "."
		if len(dots) > 3 {
			dots = "."
		}
		fmt.Printf("\rSearching for device%-3s", dots)

		for _, pName := range getPossiblePorts() {
			config := &serial.Config{Name: pName, Baud: 115200, ReadTimeout: time.Second * 2}
			p, err := serial.OpenPort(config)
			if err == nil {
				p.Write([]byte("IDENTIFY\n"))
				
				buf := make([]byte, 64)
				n, _ := p.Read(buf)
				
				if n > 0 && strings.Contains(string(buf[:n]), "ESP_MONITOR_READY") {
					fmt.Printf("\n[OK] Device found on %s\n", pName)
					return p, pName
				}
				p.Close()
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func collectStats(activePort string) (string, error) {
	cPctAll, _ := cpu.Percent(0, false)
	cInfo, _ := cpu.Info()
	m, _ := mem.VirtualMemory()
	t, _ := host.SensorsTemperatures()
	uptime, _ := host.Uptime()
	
	var cpuTemp float64
	if len(t) > 0 { cpuTemp = t[0].Temperature }

	d1, _ := disk.IOCounters()
	n1, _ := net.IOCounters(false)
	time.Sleep(500 * time.Millisecond)
	d2, _ := disk.IOCounters()
	n2, _ := net.IOCounters(false)

	var readSpd, writeSpd uint64
	for _, s := range d2 {
		if d1[s.Name].ReadBytes > 0 {
			readSpd += s.ReadBytes - d1[s.Name].ReadBytes
			writeSpd += s.WriteBytes - d1[s.Name].WriteBytes
		}
	}

	var netDown uint64
	if len(n2) > 0 && len(n1) > 0 {
		netDown = (n2[0].BytesRecv - n1[0].BytesRecv) / 1024
	}

	data := fmt.Sprintf("CP:%.1f|CT:%.1f|CF:%d|RM:%.1f|RG:%.1f|DR:%d|DW:%d|ND:%d|UP:%d\n",
		cPctAll[0], cpuTemp, int(cInfo[0].Mhz), m.UsedPercent,
		float64(m.Used)/1024/1024/1024, readSpd/1024, writeSpd/1024,
		netDown, uptime,
	)

	fmt.Printf("\rSending: CPU %.1f%% | RAM %.1f%% | Port: %s    ", cPctAll[0], m.UsedPercent, activePort)
	return data, nil
}
