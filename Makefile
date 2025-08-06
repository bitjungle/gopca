# Makefile for gopca - A professional-grade PCA toolkit
# This Makefile provides common development tasks for building, testing, and running the PCA toolkit

# Variables
BINARY_NAME := gopca-cli
DESKTOP_NAME := gopca-desktop
CSV_NAME := gocsv
BUILD_DIR := build
CLI_PATH := cmd/gopca-cli/main.go
DESKTOP_PATH := cmd/gopca-desktop
CSV_PATH := cmd/gocsv
COVERAGE_FILE := coverage.out

# Shortcuts for CLI builds
cli: build
cli-all: build-all

# Shortcuts for desktop/GUI builds  
desktop: pca-build
desktop-dev: pca-dev
pca: pca-build

# Shortcuts for CSV editor builds
csv: csv-build

# Cross-platform build variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Binary extension (for Windows)
ifeq ($(GOOS),windows)
	EXT := .exe
else
	EXT :=
endif

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod
GOGET := $(GOCMD) get

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags="-s -w \
	-X github.com/bitjungle/gopca/internal/version.Version=$(VERSION) \
	-X github.com/bitjungle/gopca/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/bitjungle/gopca/internal/version.BuildDate=$(BUILD_DATE)"

# Desktop build flags (for Wails)
DESKTOP_LDFLAGS := -ldflags "-s -w \
	-X github.com/bitjungle/gopca/internal/version.Version=$(VERSION) \
	-X github.com/bitjungle/gopca/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/bitjungle/gopca/internal/version.BuildDate=$(BUILD_DATE)"

# Check if golangci-lint is installed
GOLINT := $(shell which golangci-lint 2> /dev/null)

# Check if wails is installed - check in PATH and common locations
WAILS := $(shell which wails 2> /dev/null || echo "$${HOME}/go/bin/wails")

# Default target
.DEFAULT_GOAL := all

# Phony targets
.PHONY: all build cli cli-all build-cross build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64 build-all pca-dev pca-build pca-build-all pca-run pca-deps csv-dev csv-build csv-build-all csv-run csv-deps build-everything test test-verbose test-coverage fmt lint run-pca-iris clean clean-cross install deps deps-all install-hooks sign sign-cli sign-desktop sign-csv notarize notarize-cli notarize-desktop notarize-csv sign-and-notarize help

## all: Build all applications for current platform and run tests
all: build pca-build csv-build test

## build: Build the CLI binary
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(EXT) $(CLI_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(EXT)"

## build-cross: Generic cross-platform build (use with GOOS and GOARCH)
build-cross:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$(GOOS)" = "windows" ]; then \
		GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe $(CLI_PATH); \
		echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe"; \
	else \
		GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) $(CLI_PATH); \
		echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)"; \
	fi

## build-darwin-amd64: Build for macOS Intel
build-darwin-amd64:
	@$(MAKE) build-cross GOOS=darwin GOARCH=amd64

## build-darwin-arm64: Build for macOS Apple Silicon
build-darwin-arm64:
	@$(MAKE) build-cross GOOS=darwin GOARCH=arm64

## build-linux-amd64: Build for Linux x64
build-linux-amd64:
	@$(MAKE) build-cross GOOS=linux GOARCH=amd64

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64:
	@$(MAKE) build-cross GOOS=linux GOARCH=arm64

## build-windows-amd64: Build for Windows x64
build-windows-amd64:
	@$(MAKE) build-cross GOOS=windows GOARCH=amd64

## build-all: Build CLI for all supported platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64
	@echo "All CLI platform builds complete!"

## pca-dev: Run PCA Desktop in development mode with hot reload
pca-dev:
	@if [ -x "$(WAILS)" ]; then \
		echo "Starting PCA Desktop in development mode..."; \
		cd $(DESKTOP_PATH) && $(WAILS) dev; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## pca-build: Build PCA Desktop application for production
pca-build:
	@if [ -x "$(WAILS)" ]; then \
		echo "Building PCA Desktop application..."; \
		cd $(DESKTOP_PATH) && $(WAILS) build $(DESKTOP_LDFLAGS); \
		echo "PCA Desktop build complete. Check $(DESKTOP_PATH)/build/bin/"; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## pca-run: Run the built PCA Desktop application
pca-run:
	@if [ -f "$(DESKTOP_PATH)/build/bin/gopca-desktop.app/Contents/MacOS/gopca-desktop" ]; then \
		echo "Running PCA Desktop application..."; \
		open $(DESKTOP_PATH)/build/bin/gopca-desktop.app; \
	elif [ -f "$(DESKTOP_PATH)/build/bin/gopca-desktop" ]; then \
		echo "Running PCA Desktop application..."; \
		$(DESKTOP_PATH)/build/bin/gopca-desktop; \
	else \
		echo "PCA Desktop application not found. Build it first with 'make pca-build'"; \
		exit 1; \
	fi

## pca-deps: Install frontend dependencies for PCA Desktop
pca-deps:
	@echo "Installing PCA Desktop frontend dependencies..."
	@cd $(DESKTOP_PATH)/frontend && npm install
	@echo "PCA Desktop dependencies installed"

## csv-dev: Run CSV editor in development mode with hot reload
csv-dev:
	@if [ -x "$(WAILS)" ]; then \
		echo "Starting CSV editor in development mode..."; \
		cd $(CSV_PATH) && $(WAILS) dev; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## csv-build: Build CSV editor application for production
csv-build:
	@if [ -x "$(WAILS)" ]; then \
		echo "Building CSV editor application..."; \
		cd $(CSV_PATH) && $(WAILS) build $(DESKTOP_LDFLAGS); \
		echo "CSV editor build complete. Check $(CSV_PATH)/build/bin/"; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## csv-run: Run the built CSV editor application
csv-run:
	@if [ -f "$(CSV_PATH)/build/bin/gocsv.app/Contents/MacOS/gocsv" ]; then \
		echo "Running CSV editor application..."; \
		open $(CSV_PATH)/build/bin/gocsv.app; \
	elif [ -f "$(CSV_PATH)/build/bin/gocsv" ]; then \
		echo "Running CSV editor application..."; \
		$(CSV_PATH)/build/bin/gocsv; \
	else \
		echo "CSV editor application not found. Build it first with 'make csv-build'"; \
		exit 1; \
	fi

## sign: Sign all macOS binaries (requires Apple Developer ID)
sign:
	@echo "Signing macOS binaries..."
	@./scripts/sign-macos.sh

## sign-cli: Sign only the CLI binary
sign-cli:
	@echo "Signing CLI binary..."
	@./scripts/sign-macos.sh | grep -A 3 "CLI"

## sign-desktop: Sign only the GoPCA Desktop app
sign-desktop:
	@echo "Signing GoPCA Desktop app..."
	@if [ -f "$(PCA_PATH)/build/bin/gopca-desktop.app" ]; then \
		codesign --force --deep --sign "$${APPLE_DEVELOPER_ID:-Developer ID Application: Rune Mathisen (LV599Q54BU)}" \
			--options runtime --timestamp \
			"$(PCA_PATH)/build/bin/gopca-desktop.app" && \
		codesign --verify --verbose "$(PCA_PATH)/build/bin/gopca-desktop.app"; \
	else \
		echo "GoPCA Desktop app not found. Build it first with 'make pca-build'"; \
		exit 1; \
	fi

## sign-csv: Sign only the GoCSV app
sign-csv:
	@echo "Signing GoCSV app..."
	@if [ -f "$(CSV_PATH)/build/bin/gocsv.app" ]; then \
		codesign --force --deep --sign "$${APPLE_DEVELOPER_ID:-Developer ID Application: Rune Mathisen (LV599Q54BU)}" \
			--options runtime --timestamp \
			"$(CSV_PATH)/build/bin/gocsv.app" && \
		codesign --verify --verbose "$(CSV_PATH)/build/bin/gocsv.app"; \
	else \
		echo "GoCSV app not found. Build it first with 'make csv-build'"; \
		exit 1; \
	fi

## notarize: Notarize all macOS binaries (requires .env with APPLE_APP_SPECIFIC_PASSWORD)
notarize:
	@echo "Notarizing macOS binaries..."
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs) && ./scripts/notarize-macos.sh; \
	else \
		./scripts/notarize-macos.sh; \
	fi

## notarize-cli: Notarize only the CLI binary
notarize-cli:
	@echo "Notarizing CLI binary..."
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "CLI binary not found. Build it first with 'make build'"; \
		exit 1; \
	fi
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
	fi; \
	APPLE_ID="$${APPLE_ID:-mathrune@icloud.com}" \
	APPLE_TEAM_ID="$${APPLE_TEAM_ID:-LV599Q54BU}" \
	./scripts/notarize-macos.sh cli-only

## notarize-desktop: Notarize only the GoPCA Desktop app
notarize-desktop:
	@echo "Notarizing GoPCA Desktop app..."
	@if [ ! -d "$(DESKTOP_PATH)/build/bin/gopca-desktop.app" ]; then \
		echo "GoPCA Desktop app not found. Build it first with 'make pca-build'"; \
		exit 1; \
	fi
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
	fi; \
	APPLE_ID="$${APPLE_ID:-mathrune@icloud.com}" \
	APPLE_TEAM_ID="$${APPLE_TEAM_ID:-LV599Q54BU}" \
	./scripts/notarize-macos.sh desktop-only

## notarize-csv: Notarize only the GoCSV app
notarize-csv:
	@echo "Notarizing GoCSV app..."
	@if [ ! -d "$(CSV_PATH)/build/bin/gocsv.app" ]; then \
		echo "GoCSV app not found. Build it first with 'make csv-build'"; \
		exit 1; \
	fi
	@if [ -f .env ]; then \
		export $$(grep -v '^#' .env | xargs); \
	fi; \
	APPLE_ID="$${APPLE_ID:-mathrune@icloud.com}" \
	APPLE_TEAM_ID="$${APPLE_TEAM_ID:-LV599Q54BU}" \
	./scripts/notarize-macos.sh csv-only

## sign-and-notarize: Sign and notarize all macOS binaries
sign-and-notarize: sign notarize
	@echo "All binaries signed and notarized!"

## csv-deps: Install frontend dependencies for CSV editor
csv-deps:
	@echo "Installing CSV editor frontend dependencies..."
	@cd $(CSV_PATH)/frontend && npm install
	@echo "CSV editor dependencies installed"

## test: Run all tests with coverage
test:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

## test-verbose: Run tests with detailed output
test-verbose:
	@echo "Running tests with verbose output..."
	$(GOTEST) -v -cover ./...

## test-coverage: Run tests and generate detailed coverage report
test-coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format all Go code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) ./...
	@echo "Formatting complete"

## lint: Run golangci-lint (if available)
lint:
ifdef GOLINT
	@echo "Running golangci-lint..."
	golangci-lint run
else
	@echo "golangci-lint not found. Install it with:"
	@echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@echo "Skipping lint step..."
endif

## run-pca-iris: Execute PCA analysis on iris dataset
run-pca-iris: build
	@echo "Running PCA analysis on iris dataset..."
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f json --output-all --include-metrics internal/datasets/iris.csv


## clean: Remove build artifacts and generated files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DESKTOP_PATH)/build/bin
	@rm -rf $(CSV_PATH)/build/bin
	@rm -f $(COVERAGE_FILE) coverage.html
	@echo "Clean complete"

## clean-cross: Remove cross-compiled binaries
clean-cross:
	@echo "Cleaning cross-compiled binaries..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)-*
	@echo "Cross-compiled binaries removed"

## install: Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	$(GOCMD) install $(CLI_PATH)
	@echo "Installation complete"

## deps: Download and tidy module dependencies
deps:
	@echo "Downloading Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Go dependencies updated"

## deps-all: Install all dependencies (Go + npm for all apps)
deps-all: deps
	@echo "Installing all frontend dependencies..."
	@npm install
	@echo "All dependencies installed"

## install-hooks: Install git pre-commit hooks
install-hooks:
	@echo "Installing git hooks..."
	@./scripts/install-hooks.sh

## ci-test: Run tests for CI (excluding desktop)
ci-test:
	@echo "Running CI tests (excluding desktop)..."
	@./scripts/ci/test-core.sh

## ci-lint: Run linter for CI (excluding desktop)
ci-lint:
	@echo "Running CI linter (excluding desktop)..."
ifdef GOLINT
	golangci-lint run --timeout=5m ./internal/... ./pkg/... ./cmd/gopca-cli/...
else
	@echo "golangci-lint not found, skipping..."
endif

## ci-build-cli: Build CLI for all platforms in CI
ci-build-cli: build-all

## pca-build-all: Build PCA Desktop for all platforms
pca-build-all:
	@if [ -x "$(WAILS)" ]; then \
		echo "Building PCA Desktop for all platforms..."; \
		cd $(DESKTOP_PATH) && $(WAILS) build -platform darwin/amd64,darwin/arm64,windows/amd64,linux/amd64 $(DESKTOP_LDFLAGS); \
		echo "PCA Desktop builds complete for all platforms"; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## csv-build-all: Build CSV editor for all platforms
csv-build-all:
	@if [ -x "$(WAILS)" ]; then \
		echo "Building CSV editor for all platforms..."; \
		cd $(CSV_PATH) && $(WAILS) build -platform darwin/amd64,darwin/arm64,windows/amd64,linux/amd64 $(DESKTOP_LDFLAGS); \
		echo "CSV editor builds complete for all platforms"; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## build-everything: Build all applications for all platforms
build-everything: build-all pca-build-all csv-build-all
	@echo "All applications built for all platforms!"

## ci-build-desktop: Build desktop app in CI
ci-build-desktop:
	@echo "Building desktop app for CI..."
	@PLATFORM=$(GOOS) ./scripts/ci/build-desktop.sh

## ci-setup: Setup CI environment
ci-setup:
	@./scripts/ci/setup-environment.sh

## ci-install-deps: Install platform-specific dependencies
ci-install-deps:
ifeq ($(shell uname -s),Linux)
	@./scripts/ci/install-linux-deps.sh
else
	@echo "No special dependencies needed for $(shell uname -s)"
endif

## help: Display this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ": "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sed 's/^## //'
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Quick start:"
	@echo "  make deps-all         # Install all dependencies (first time)"
	@echo "  make                  # Build CLI and test (default)"
	@echo "  make pca-dev          # Run PCA Desktop in dev mode"
	@echo "  make csv-dev          # Run CSV editor in dev mode"
	@echo ""
	@echo "Building CLI:"
	@echo "  make build            # Build CLI for current platform"
	@echo "  make build-all        # Build CLI for all platforms"
	@echo "  make build-linux-amd64   # Build for Linux x64"
	@echo "  make build-darwin-arm64  # Build for macOS Apple Silicon"
	@echo ""
	@echo "Code Signing (macOS):"
	@echo "  make sign             # Sign all macOS binaries"
	@echo "  make sign-cli         # Sign CLI binary only"
	@echo "  make sign-desktop     # Sign GoPCA Desktop app only"
	@echo "  make sign-csv         # Sign GoCSV app only"
	@echo ""
	@echo "Notarization (macOS):"
	@echo "  make notarize         # Notarize all signed binaries"
	@echo "  make notarize-cli     # Notarize CLI only"
	@echo "  make notarize-desktop # Notarize GoPCA Desktop only"
	@echo "  make notarize-csv     # Notarize GoCSV only"
	@echo "  make sign-and-notarize # Sign and notarize everything"
	@echo "  make build-windows-amd64 # Build for Windows x64"
	@echo ""
	@echo "PCA Desktop application:"
	@echo "  make pca-deps         # Install dependencies"
	@echo "  make pca-dev          # Run in development mode"
	@echo "  make pca-build        # Build for current platform"
	@echo "  make pca-build-all    # Build for all platforms"
	@echo "  make pca-run          # Run the built application"
	@echo ""
	@echo "CSV editor application:"
	@echo "  make csv-deps         # Install dependencies"
	@echo "  make csv-dev          # Run in development mode"
	@echo "  make csv-build        # Build for current platform"
	@echo "  make csv-build-all    # Build for all platforms"
	@echo "  make csv-run          # Run the built application"
	@echo ""
	@echo "Build everything:"
	@echo "  make build-everything # Build all apps for all platforms"
	@echo ""
	@echo "Testing & quality:"
	@echo "  make test             # Run tests with coverage"
	@echo "  make test-verbose     # Run tests with detailed output"
	@echo "  make fmt              # Format all Go code"
	@echo "  make lint             # Run linter (if installed)"
	@echo "  make run-pca-iris     # Run PCA on example dataset"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean            # Clean all build artifacts"
	@echo "  make clean-cross      # Clean cross-compiled binaries"
	@echo "  make deps             # Update Go dependencies"
	@echo "  make deps-all         # Install all dependencies"
	@echo "  make install-hooks    # Install git pre-commit hooks"