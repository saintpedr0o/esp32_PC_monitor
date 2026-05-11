#!/bin/bash

mkdir -p build

go build -ldflags="-s -w" -o build/esp32-monitor-amd64-bin main.go || { echo "Go build failed"; rm -rf build/; exit 1; }

nfpm pkg --packager deb --target build/esp32-monitor-amd64.deb || { echo "nfpm packaging failed"; rm -rf build/; exit 1; }

echo "Linux build done: build/esp32-monitor-amd64.deb"
