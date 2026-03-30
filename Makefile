.PHONY: all build dist linux-arm64 macos-arm64 macos-amd64 clean clean-dist

DIST_DIR := dist

all: build

build: linux-arm64 macos-arm64 macos-amd64

dist: clean-dist build
	mkdir -p $(DIST_DIR)
	cp zone2 $(DIST_DIR)/zone2-linux-arm64
	cp zone2-macos-arm64 $(DIST_DIR)/zone2-macos-arm64
	cp zone2-macos-amd64 $(DIST_DIR)/zone2-macos-amd64

linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o zone2 zone2.go

macos-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o zone2-macos-arm64 zone2.go

macos-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o zone2-macos-amd64 zone2.go

clean: clean-dist
	rm -f zone2 zone2-macos-arm64 zone2-macos-amd64

clean-dist:
	rm -rf $(DIST_DIR)
