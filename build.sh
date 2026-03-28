#!/usr/bin/env bash

set -euo pipefail

# Home Assistant OS on Raspberry Pi 5 (ARM64)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o zone2 zone2.go

# Local macOS builds
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o zone2-macos-arm64 zone2.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o zone2-macos-amd64 zone2.go
