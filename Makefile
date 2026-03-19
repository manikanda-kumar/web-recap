.PHONY: build build-all dmg install-safari-helper test clean install help

# Variables
VERSION ?= 0.1.0
BINARY_NAME := web-recap
DIST_DIR := dist
GO := go
GOFLAGS := -ldflags="-s -w"

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
	codesign --force --deep --sign - "$$APP_DIR"; \
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
	@printf '%s\n' '#!/bin/sh' 'set -eu' '' 'APP_PATH=""' 'if [ -d "/Applications/WebRecap.app" ]; then' '  APP_PATH="/Applications/WebRecap.app"' 'elif [ -d "$$HOME/Applications/WebRecap.app" ]; then' '  APP_PATH="$$HOME/Applications/WebRecap.app"' 'else' '  echo "WebRecap.app not found in /Applications or ~/Applications" >&2' '  exit 1' 'fi' '' 'if [ "$$#" -eq 0 ]; then' '  cat >&2 <<'\''EOF'\''' 'Usage:' '  web-recap-safari bookmarks --browser safari' '  web-recap-safari --browser safari --date "$$(date +%F)"' '  web-recap-safari bookmarks --browser safari -o output.json' '' 'Notes:' '  - Launches the FDA-enabled WebRecap.app via macOS '\''open'\''' '  - If you do not pass -o/--output, output is captured to a temp file and printed' '  - Intended for Safari commands where direct CLI execution is blocked by macOS privacy' 'EOF' '  exit 1' 'fi' '' 'has_output=0' 'output_path=""' 'next_is_output_path=0' 'for arg in "$$@"; do' '  if [ "$$next_is_output_path" -eq 1 ]; then' '    has_output=1' '    output_path="$$arg"' '    next_is_output_path=0' '    continue' '  fi' '' '  case "$$arg" in' '    -o|--output)' '      next_is_output_path=1' '      ;;' '    --output=*)' '      has_output=1' '      output_path=$${arg#*=}' '      ;;' '  esac' 'done' '' 'if [ "$$next_is_output_path" -eq 1 ]; then' '  echo "Missing value for -o/--output" >&2' '  exit 1' 'fi' '' 'cleanup_tmp=0' 'if [ "$$has_output" -eq 0 ]; then' '  output_path=$$(mktemp /tmp/web-recap-safari.XXXXXX.json)' '  cleanup_tmp=1' 'fi' '' 'cleanup() {' '  if [ "$$cleanup_tmp" -eq 1 ]; then' '    rm -f "$$output_path"' '  fi' '}' 'trap cleanup EXIT INT TERM' '' 'rm -f "$$output_path"' '' 'if [ "$$has_output" -eq 1 ]; then' '  open -n -a "$$APP_PATH" --args "$$@" >/dev/null 2>/dev/null' 'else' '  open -n -a "$$APP_PATH" --args "$$@" -o "$$output_path" >/dev/null 2>/dev/null' 'fi' '' 'found=0' 'prev_size=-1' 'stable_count=0' 'attempt=0' 'while [ "$$attempt" -lt 120 ]; do' '  if [ -f "$$output_path" ]; then' '    found=1' '    size=$$(wc -c < "$$output_path" | tr -d '\'' '\'')' '    if [ "$$size" = "$$prev_size" ] && [ "$$size" -gt 0 ] 2>/dev/null; then' '      stable_count=$$((stable_count + 1))' '      if [ "$$stable_count" -ge 2 ]; then' '        break' '      fi' '    else' '      stable_count=0' '      prev_size=$$size' '    fi' '  fi' '  attempt=$$((attempt + 1))' '  sleep 0.25' 'done' '' 'if [ "$$found" -ne 1 ] || [ ! -f "$$output_path" ]; then' '  echo "Timed out waiting for WebRecap.app output: $$output_path" >&2' '  exit 1' 'fi' '' 'if [ "$$cleanup_tmp" -eq 1 ]; then' '  cat "$$output_path"' 'else' '  echo "Wrote output to $$output_path"' 'fi' > /opt/homebrew/bin/web-recap-safari
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
