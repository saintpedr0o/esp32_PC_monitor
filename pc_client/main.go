package main

import (
	_ "embed"
	"fmt"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	netutil "github.com/shirou/gopsutil/v3/net"

	"pc_monitor/internal/platform"
)

//go:embed icon.png
var iconBytes []byte

type uiData struct {
	status string
	stats  string
	log    string
}

const (
	appName          = "ESP32Monitor"
	targetAdapterMac = "88:88:88:88:88:88"
	targetAddr       = "192.168.4.1:1234"
)

func main() {
	iconRes := fyne.NewStaticResource("icon.png", iconBytes)
	myApp := app.NewWithID("ESP32 Monitor")
    myApp.SetIcon(iconRes)
	myApp.Settings().SetTheme(theme.DarkTheme())
	window := myApp.NewWindow("ESP32 Monitor")

	statusLabel := widget.NewLabel("Status: 🔍 Searching...")
	statusLabel.Alignment = fyne.TextAlignCenter

	statsLabel := widget.NewLabel("CPU: ----% | RAM: ----% | Adapter: ----")
	statsLabel.Alignment = fyne.TextAlignCenter

	logLabel := widget.NewLabel("Last Sent Data:")
	logView := widget.NewMultiLineEntry()
	logView.Disable()

	autostartCheck := widget.NewCheck("Autostart", func(c bool) { platform.ToggleAutostart(c, appName) })
	autostartCheck.Checked = platform.CheckAutostartStatus(appName)

	platform.SetupSystemTray(myApp, window, appName, iconRes)

	uiChan := make(chan uiData, 5)

	window.SetContent(container.NewVBox(
		statusLabel,
		widget.NewSeparator(),
		statsLabel,
		widget.NewSeparator(),
		logLabel,
		container.NewGridWithRows(1, logView),
		autostartCheck,
	))
	window.Resize(fyne.NewSize(520, 200))
	window.SetCloseIntercept(window.Hide)

	go func() {
		for d := range uiChan {
			fyne.Do(func() {
				if d.status != "" { statusLabel.SetText(d.status) }
				if d.stats != "" { statsLabel.SetText(d.stats) }
				if d.log != "" { logView.SetText(d.log) }
			})
		}
	}()

	go func() {
		var conn *net.UDPConn
		var err error

		for {
			localIP, adapterName := getAdapterInfoByMAC(targetAdapterMac)
			if localIP == "" {
				uiChan <- uiData{status: "Error: Adapter [" + targetAdapterMac + "] not found"}
				time.Sleep(3 * time.Second)
				continue
			}

			lAddr, _ := net.ResolveUDPAddr("udp", localIP+":0")
			rAddr, _ := net.ResolveUDPAddr("udp", targetAddr)

			conn, err = net.DialUDP("udp", lAddr, rAddr)
			if err != nil {
				uiChan <- uiData{status: "Dial Error: " + err.Error()}
				time.Sleep(2 * time.Second)
				continue
			}

			uiChan <- uiData{status: "Status: ✅ Connected " + localIP}

			for {
				statsRaw, cpuP, ramP := collectAllStats()
				
				uiChan <- uiData{
					stats: fmt.Sprintf("CPU: %.1f%% | RAM: %.1f%% | Adapter: %s", cpuP, ramP, adapterName),
					log:   statsRaw,
				}

				_, err := conn.Write([]byte(statsRaw))
				if err != nil {
					uiChan <- uiData{status: "Send Error: Connection Lost"}
					conn.Close()
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	window.ShowAndRun()
}

func getAdapterInfoByMAC(targetMac string) (string, string) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if strings.ToLower(iface.HardwareAddr.String()) == strings.ToLower(targetMac) {
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), iface.Name
					}
				}
			}
		}
	}
	return "", ""
}

func collectAllStats() (string, float64, float64) {
	d1, _ := disk.IOCounters()
	n1, _ := netutil.IOCounters(false)
	time.Sleep(1 * time.Second)

	cPctAll, _ := cpu.Percent(0, false)
	m, _ := mem.VirtualMemory()
	cInfo, _ := cpu.Info()
	t, _ := host.SensorsTemperatures()
	uptime, _ := host.Uptime()
	d2, _ := disk.IOCounters()
	n2, _ := netutil.IOCounters(false)

	var cpuTemp float64
	if len(t) > 0 { cpuTemp = t[0].Temperature }

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

	raw := fmt.Sprintf("CP:%.1f|CT:%.1f|CF:%d|RM:%.1f|RG:%.1f|DR:%d|DW:%d|ND:%d|UP:%d\n",
		cPctAll[0], cpuTemp, int(cInfo[0].Mhz), m.UsedPercent,
		float64(m.Used)/1e9, readSpd/1024, writeSpd/1024, netDown, uptime,
	)

	return raw, cPctAll[0], m.UsedPercent
}
