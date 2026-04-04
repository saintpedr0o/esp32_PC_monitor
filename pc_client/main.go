package main

import (
	"fmt"
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

    "pc_monitor/internal/platform"
)

type uiData struct {
	status string
	stats  string
	log    string
}

const appName = "ESP32BTMonitor"

func main() {
	myApp := app.NewWithID("ESP32 Monitor")
	window := myApp.NewWindow("ESP32 Bluetooth Monitor")

	statusLabel := widget.NewLabel("")
	statsLabel := widget.NewLabel("")
	logView := widget.NewMultiLineEntry()
	logView.Disable()

	autostartCheck := widget.NewCheck("Autostart", func(c bool) { platform.ToggleAutostart(c, appName) })
	autostartCheck.Checked = platform.CheckAutostartStatus(appName)

	platform.SetupSystemTray(myApp, window, appName)

	uiChan := make(chan uiData, 5)

	window.SetContent(container.NewVBox(
		statusLabel,
		widget.NewSeparator(),
		statsLabel,
		container.NewGridWithRows(1, logView),
		autostartCheck,
	))
	window.Resize(fyne.NewSize(600, 200))
	window.SetCloseIntercept(window.Hide)

	go runMonitoringLoop(uiChan)

	go func() {
		for d := range uiChan {
			fyne.Do(func() {
				statusLabel.SetText(d.status)
				statsLabel.SetText(d.stats)
				logView.SetText(d.log)
			})
		}
	}()

	window.ShowAndRun()
}

func runMonitoringLoop(uiChan chan uiData) {
	disconnectedState := uiData{
		status: "Status: Searching...",
		stats:  "CPU: ----% | RAM: ----% | Port: None",
		log:    "Disconnected",
	}

	for {
		port, activePort := findDevice()
		if port == nil {
			uiChan <- disconnectedState
			time.Sleep(2 * time.Second)
			continue
		}

		for {
			data, statsStr, err := collectStats(activePort)
			if err != nil {
				uiChan <- disconnectedState
				break
			}

			if _, err := port.Write([]byte(data)); err != nil {
				break
			}

			uiChan <- uiData{
				status: "Connected: " + activePort,
				stats:  statsStr,
				log:    "Sent: " + strings.TrimSpace(data),
			}
			time.Sleep(time.Second)
		}
		port.Close()
	}
}

func collectStats(activePort string) (string, string, error) {
	cPctAll, _ := cpu.Percent(0, false)
	m, _ := mem.VirtualMemory()
	cInfo, _ := cpu.Info()
	t, _ := host.SensorsTemperatures()
	uptime, _ := host.Uptime()

	var cpuTemp float64
	if len(t) > 0 {
		cpuTemp = t[0].Temperature
	}

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
	ports := platform.GetSerialPorts() 

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
