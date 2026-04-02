package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/tarm/serial"
)

type uiData struct {
	status string
	stats  string
	log    string
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("ESP32 PC Monitor")

	statusLabel := widget.NewLabel("Status: Search...")
	statsLabel := widget.NewLabel("Stats: 0%")
	logView := widget.NewMultiLineEntry()
	logView.Disable()

	uiChan := make(chan uiData, 10)

	content := container.NewVBox(
		statusLabel,
		widget.NewSeparator(),
		statsLabel,
		widget.NewLabel("Log:"),
		container.NewStack(logView),
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(500, 450))

	go runMonitoringLoop(uiChan)

	go func() {
		for data := range uiChan {
			d := data 
			
			fyne.Do(func() {
				if d.status != "" {
					statusLabel.SetText(d.status)
				}
				if d.stats != "" {
					statsLabel.SetText(d.stats)
				}
				if d.log != "" {
					logView.SetText(d.log + "\n" + logView.Text)
				}
			})
		}
	}()

	window.ShowAndRun()
}

func runMonitoringLoop(uiChan chan uiData) {
	for {
		port, activePort := findDevice()
		if port == nil {
			uiChan <- uiData{status: "Status: Search device..."}
			time.Sleep(1 * time.Second)
			continue
		}

		uiChan <- uiData{status: "Connected: " + activePort, log: "[OK] Connected"}

		for {
			data, statsStr, err := collectStats(activePort)
			if err != nil {
				uiChan <- uiData{log: "[!] Error: " + err.Error()}
				break
			}

			_, err = port.Write([]byte(data))
			if err != nil {
				break
			}
			
			uiChan <- uiData{stats: statsStr, log: "Send: " + strings.TrimSpace(data)}
			time.Sleep(1 * time.Second)
		}
		port.Close()
		uiChan <- uiData{status: "Status: Reconnecting..."}
		time.Sleep(2 * time.Second)
	}
}

func collectStats(activePort string) (string, string, error) {
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
		float64(m.Used)/1e9, readSpd/1024, writeSpd/1024, netDown, uptime,
	)

	statsStr := fmt.Sprintf("CPU: %.1f%% | RAM: %.1f%% | Port: %s", cPctAll[0], m.UsedPercent, activePort)

	return data, statsStr, nil
}

func findDevice() (*serial.Port, string) {
	var ports []string
	if runtime.GOOS == "windows" {
		for i := 1; i <= 20; i++ { ports = append(ports, fmt.Sprintf("COM%d", i)) }
	} else {
		for i := 0; i <= 9; i++ { ports = append(ports, fmt.Sprintf("/dev/rfcomm%d", i)) }
	}

	for _, pName := range ports {
		config := &serial.Config{Name: pName, Baud: 115200, ReadTimeout: time.Second * 1}
		p, err := serial.OpenPort(config)
		if err == nil {
			p.Write([]byte("IDENTIFY\n"))
			buf := make([]byte, 64)
			n, _ := p.Read(buf)
			if n > 0 && strings.Contains(string(buf[:n]), "ESP_MONITOR_READY") {
				return p, pName
			}
			p.Close()
		}
	}
	return nil, ""
}
