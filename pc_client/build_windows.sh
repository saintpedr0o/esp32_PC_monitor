#!/bin/bash

mkdir -p build

rsrc -arch amd64 -ico icon.ico -o icon_windows.syso || { echo "rsrc failed"; rm -rf build/; exit 1; }

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
go build -ldflags="-s -w -H=windowsgui" -o build/ESP32BTMonitor.exe .

rm -f icon_windows.syso

echo "Windows build done: build/ESP32BTMonitor.exe"
