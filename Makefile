# APP_NAME=ultahost-agent
APP_BASE_NAME := ultahost-agent-binary
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

APP_NAME := $(APP_BASE_NAME)-$(OS)-$(ARCH)

BUILD_DIR=dist
ENTRY_POINT=cmd/agent/main.go

build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_BASE_NAME)-darwin-arm64 $(ENTRY_POINT)
	strip $(BUILD_DIR)/$(APP_BASE_NAME)-darwin-arm64 || true
	upx --best --lzma $(BUILD_DIR)/$(APP_BASE_NAME)-darwin-arm64 || true

size:
	du -h $(BUILD_DIR)/$(APP_NAME)

clean:
	rm -f $(BUILD_DIR)/$(APP_NAME)
