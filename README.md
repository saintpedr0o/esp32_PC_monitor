# esp32_PC_monitor

**esp32_PC_monitor** is a complete solution consisting of an ESP32-based device and a PC client to monitor your computer's real-time statistics via Wi-Fi (UDP).

## 🔌 Hardware (ESP32 Firmware)

The `esp32_PC_monitor.ino` file contains the firmware for the ESP32 controller. It acts as a Wi-Fi Access Point and listens for incoming UDP packets to drive the display (TFT 1.77") showing your PC stats.

### 📚 Required Libraries

To compile the firmware, you need to install the following libraries via the **Arduino Library Manager**:

* **WiFi**: Built-in library for ESP32.
* **WiFiUDP**: Built-in library for handling UDP data stream.
* **Adafruit ST7735**: Hardware-specific library for the 1.77" TFT display.

### Wiring Diagram

Below is the connection scheme for the ESP32 and the TFT display.
![Wiring Diagram](./docs/schema.png)

## 🖥 PC Client (pc_client)

**ESP32 WiFi Monitor** is a lightweight cross-platform application designed to gather PC stats and send them via UDP. It supports **Smart Adapter Binding**, allowing it to send data through a specific Wi-Fi adapter (e.g., a dedicated USB dongle) without affecting your main internet connection.

### ⚙️ Key Features

* **Smart Adapter Binding**: Automatically finds and binds to a specific network interface by its **MAC Address**.
* **Zero-Interference**: Keeps your main internet connection clean by using a secondary Wi-Fi adapter for ESP32 communication.
* **System Tray Integration**: Runs quietly in the background with a minimalist interface.
* **Autostart**: Built-in support to launch automatically on system startup.

### 🚀 Configuration & Installation

Before building or running, configure your target adapter in `main.go`:

```go
const (
    // MAC of your dedicated USB Wi-Fi adapter
	targetAdapterMac = "88:88:88:88:88:88"
)

```

#### Windows

1. Go to the [Releases](https://github.com/saintpedr0o/esp32_PC_monitor/releases) page.
2. Download `ESP32Monitor.exe`.
3. Run the executable.

#### Linux (Debian/Ubuntu)
1. Go to the [Releases](https://github.com/saintpedr0o/esp32_PC_monitor/releases) page.
2. Download the `.deb` package.
3. Install via terminal:

```bash
sudo dpkg -i esp32-monitor-amd64.deb

```

### 🛠 Development & Build Instructions
If you wish to build the application from source, ensure you have Go 1.20+ installed.

#### Build Requirements
To compile for all platforms, you will need the following tools:

nfpm: Used for creating Linux .deb packages.

```bash
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

rsrc: Required for embedding the icon into the Windows binary.

```bash
go install github.com/akavel/rsrc@latest
```

MinGW-w64: Necessary for CGO cross-compilation under Windows.

```bash
sudo apt install mingw-w64
```
#### Compiling
The project includes automation scripts to simplify the build process:

For Linux:

```bash
chmod +x build_linux.sh
./build_linux.sh
```
(Generates a .deb package in the build/ directory)

For Windows:

```bash
chmod +x build_windows.sh
./build_windows.sh
```
(Generates a .exe file in the build/ directory)

### 📂 Project Structure

* **pc_client/main.go**: Entry point. Handles `gopsutil` data collection and UDP broadcasting via a specific MAC-bound interface.
* **pc_client/internal/platform/**: Platform-specific Go implementations (system tray logic, autostart).
* **pc_client/build_*.sh**: Automation scripts for cross-compilation and creating installers (Linux `.deb` and Windows `.exe`).
* **pc_client/build/**: Output directory containing compiled binaries and distribution packages.
* **pc_client/nfpm.yaml**: Configuration file, used to generate the Debian package.
* **pc_client/esp32monitor.desktop**: Linux desktop entry file to integrate the app into application menus.
* **pc_client/go.mod & go.sum**: Go module files defining project dependencies.
* **pc_client/icon.ico & icon.png**: Application icons. The `.ico` file is multi-layer (16x16 to 256x256) for Windows compatibility, while `.png` is used for Linux.

## 📄 License

This project is licensed under the MIT License.
