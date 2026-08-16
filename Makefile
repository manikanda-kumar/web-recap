.PHONY: build build-all dmg install-safari-helper test clean install help

# Variables
VERSION ?= 0.1.0
BINARY_NAME := web-recap
DIST_DIR := dist
GO := go
GOFLAGS := -ldflags="-s -w"
# Signing identity for WebRecap.app. Use a stable certificate (e.g. Apple
# Development) so macOS privacy grants (Full Disk Access) survive rebuilds.
# Ad-hoc signing ("-") breaks TCC grants on every rebuild (cdhash changes).
SIGN_IDENTITY ?= Apple Development: manikandakumar@gmail.com (YYR6L2AUP4)

# Platform targets
LINUX_AMD64 := $(DIST_DIR)/$(BINARY_NAME)-linux-amd64
DARWIN_AMD64 := $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64
DARWIN_ARM64 := $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64
WINDOWS_AMD64 := $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe

help:
	@echo "web-recap build targets:"
	@echo "  make build          - Build for current platform"
	@echo "  make build-all      - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make dmg            - Build WebRecap.app and package it into dist/WebRecap.dmg"
	@echo "  make install-safari-helper - Install /opt/homebrew/bin/web-recap-safari helper"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install binary to GOBIN"
	@echo "  make help           - Show this help message"

build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/web-recap

build-all: $(LINUX_AMD64) $(DARWIN_AMD64) $(DARWIN_ARM64) $(WINDOWS_AMD64)
	@echo "✓ Built all platforms"

dmg:
	@echo "Building WebRecap.app wrapper..."
	@APP_DIR="$$HOME/Applications/WebRecap.app"; \
	mkdir -p "$$APP_DIR/Contents/MacOS" "$$APP_DIR/Contents/Resources" "$(DIST_DIR)/dmg-stage"; \
	printf '%s\n' '<?xml version="1.0" encoding="UTF-8"?>' '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' '<plist version="1.0">' '<dict>' '  <key>CFBundleName</key>' '  <string>WebRecap</string>' '  <key>CFBundleDisplayName</key>' '  <string>WebRecap</string>' '  <key>CFBundleIdentifier</key>' '  <string>com.manik.webrecap.wrapper</string>' '  <key>CFBundleVersion</key>' '  <string>1.0</string>' '  <key>CFBundleShortVersionString</key>' '  <string>1.0</string>' '  <key>CFBundlePackageType</key>' '  <string>APPL</string>' '  <key>CFBundleExecutable</key>' '  <string>WebRecap</string>' '  <key>LSMinimumSystemVersion</key>' '  <string>11.0</string>' '  <key>LSUIElement</key>' '  <true/>' '</dict>' '</plist>' > "$$APP_DIR/Contents/Info.plist"; \
	printf '%s\n' '#!/bin/sh' 'set -eu' '' 'APP_DIR=$$(CDPATH= cd -- "$$(dirname -- "$$0")" && pwd)' 'BINARY="$$APP_DIR/web-recap-bin"' '' 'if [ ! -x "$$BINARY" ]; then' '  osascript -e '\''display alert "WebRecap wrapper error" message "Bundled web-recap binary is missing or not executable." as critical'\''' '  exit 1' 'fi' '' 'if [ "$$#" -eq 0 ]; then' '  osascript <<'\''APPLESCRIPT'\''' 'display dialog "WebRecap.app is a privacy wrapper for the CLI.\n\nAfter granting this app Full Disk Access, run:\n~/Applications/WebRecap.app/Contents/MacOS/WebRecap bookmarks --browser safari\n\nor\n~/Applications/WebRecap.app/Contents/MacOS/WebRecap --browser safari --date $$(date +%F)" buttons {"OK"} default button "OK"' 'APPLESCRIPT' '  exit 0' 'fi' '' 'exec "$$BINARY" "$$@"' > "$$APP_DIR/Contents/MacOS/WebRecap"; \
	chmod +x "$$APP_DIR/Contents/MacOS/WebRecap"; \
	$(GO) build $(GOFLAGS) -o "$$APP_DIR/Contents/MacOS/web-recap-bin" ./cmd/web-recap; \
	chmod +x "$$APP_DIR/Contents/MacOS/web-recap-bin"; \
	if security find-identity -v -p codesigning | grep -qF "$(SIGN_IDENTITY)"; then \
	  codesign --force --deep --sign "$(SIGN_IDENTITY)" "$$APP_DIR"; \
	else \
	  echo "WARNING: signing identity '$(SIGN_IDENTITY)' not found; falling back to ad-hoc (TCC grants will break on rebuild)"; \
	  codesign --force --deep --sign - "$$APP_DIR"; \
	fi; \
	rm -rf "$(DIST_DIR)/dmg-stage"; \
	mkdir -p "$(DIST_DIR)/dmg-stage"; \
	cp -R "$$APP_DIR" "$(DIST_DIR)/dmg-stage/"; \
	ln -sfn /Applications "$(DIST_DIR)/dmg-stage/Applications"; \
	rm -f "$(DIST_DIR)/WebRecap.dmg"; \
	hdiutil create -volname "WebRecap" -srcfolder "$(DIST_DIR)/dmg-stage" -ov -format UDZO "$(DIST_DIR)/WebRecap.dmg" >/dev/null; \
	hdiutil verify "$(DIST_DIR)/WebRecap.dmg" >/dev/null; \
	echo "✓ Created $(DIST_DIR)/WebRecap.dmg"

install-safari-helper:
	@echo "Installing Safari helper to /opt/homebrew/bin/web-recap-safari..."
	@mkdir -p /opt/homebrew/bin
	@cp scripts/web-recap-safari.sh /opt/homebrew/bin/web-recap-safari
	@chmod +x /opt/homebrew/bin/web-recap-safari
	@echo "✓ Installed /opt/homebrew/bin/web-recap-safari"

$(LINUX_AMD64):
	@mkdir -p $(DIST_DIR)
	@echo "Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $@ ./cmd/web-recap

$(DARWIN_AMD64):
	@mkdir -p $(DIST_DIR)
	@echo "Building macOS Intel..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $@ ./cmd/web-recap

$(DARWIN_ARM64):
	@mkdir -p $(DIST_DIR)
	@echo "Building macOS ARM64..."
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $@ ./cmd/web-recap

$(WINDOWS_AMD64):
	@mkdir -p $(DIST_DIR)
	@echo "Building Windows AMD64..."
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $@ ./cmd/web-recap

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

test-coverage:
	$(GO) test -cover ./...

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -rf $(DIST_DIR)
	rm -f $(BINARY_NAME)
	rm -rf $$HOME/Applications/WebRecap.app
	rm -f /opt/homebrew/bin/web-recap-safari

install: build
	@echo "Installing web-recap..."
	$(GO) install ./cmd/web-recap
	@if [ -w /opt/homebrew/bin ]; then \
		cp $(BINARY_NAME) /opt/homebrew/bin/$(BINARY_NAME); \
		echo "✓ Installed /opt/homebrew/bin/$(BINARY_NAME)"; \
	fi

deps:
	$(GO) mod download
	$(GO) mod verify

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run ./...

vet:
	$(GO) vet ./...

.PHONY: build build-all dmg install-safari-helper test test-verbose test-coverage clean install deps fmt lint vet help
