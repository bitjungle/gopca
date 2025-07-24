# Makefile for gopca - A professional-grade PCA toolkit
# This Makefile provides common development tasks for building, testing, and running the PCA toolkit

# Variables
BINARY_NAME := gopca-cli
DESKTOP_NAME := gopca-desktop
BUILD_DIR := build
CLI_PATH := cmd/gopca-cli/main.go
DESKTOP_PATH := cmd/gopca-desktop
COVERAGE_FILE := coverage.out

# Shortcuts for CLI builds
cli: build
cli-all: build-all

# Shortcuts for desktop/GUI builds  
desktop: gui-build
desktop-dev: gui-dev

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

# Build flags
LDFLAGS := -ldflags="-s -w -X github.com/bitjungle/gopca/internal/cli.Version=$(VERSION)"

# Check if golangci-lint is installed
GOLINT := $(shell which golangci-lint 2> /dev/null)

# Check if wails is installed - check in PATH and common locations
WAILS := $(shell which wails 2> /dev/null || echo "$${HOME}/go/bin/wails")

# Default target
.DEFAULT_GOAL := all

# Phony targets
.PHONY: all build cli cli-all build-cross build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64 build-all gui-dev gui-build gui-run gui-deps test test-verbose test-coverage fmt lint run-pca-iris clean clean-cross install deps help

## all: Build the binary and run tests
all: build test

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

## build-all: Build for all supported platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64
	@echo "All platform builds complete!"

## gui-dev: Run GUI in development mode with hot reload
gui-dev:
	@if [ -x "$(WAILS)" ]; then \
		echo "Starting GUI in development mode..."; \
		cd $(DESKTOP_PATH) && $(WAILS) dev; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## gui-build: Build GUI application for production
gui-build:
	@if [ -x "$(WAILS)" ]; then \
		echo "Building GUI application..."; \
		cd $(DESKTOP_PATH) && $(WAILS) build; \
		echo "GUI build complete. Check $(DESKTOP_PATH)/build/bin/"; \
	else \
		echo "Wails not found. Install it with:"; \
		echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		exit 1; \
	fi

## gui-run: Run the built GUI application
gui-run:
	@if [ -f "$(DESKTOP_PATH)/build/bin/gopca-desktop.app/Contents/MacOS/gopca-desktop" ]; then \
		echo "Running GUI application..."; \
		open $(DESKTOP_PATH)/build/bin/gopca-desktop.app; \
	elif [ -f "$(DESKTOP_PATH)/build/bin/gopca-desktop" ]; then \
		echo "Running GUI application..."; \
		$(DESKTOP_PATH)/build/bin/gopca-desktop; \
	else \
		echo "GUI application not found. Build it first with 'make gui-build'"; \
		exit 1; \
	fi

## gui-deps: Install frontend dependencies for GUI
gui-deps:
	@echo "Installing GUI frontend dependencies..."
	@cd $(DESKTOP_PATH)/frontend && npm install
	@echo "GUI dependencies installed"

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
	$(BUILD_DIR)/$(BINARY_NAME) analyze -f csv --output-all --include-metrics data/iris_data.csv


## clean: Remove build artifacts and generated files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DESKTOP_PATH)/build/bin
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
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

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
	@echo "Example workflows:"
	@echo "  make                  # Build CLI and test (default)"
	@echo "  make build            # Build the CLI binary for current platform"
	@echo "  make build-all        # Build for all supported platforms"
	@echo "  make build-linux-amd64 # Build for Linux x64"
	@echo "  make build-darwin-arm64 # Build for macOS Apple Silicon"
	@echo "  make build-windows-amd64 # Build for Windows x64"
	@echo ""
	@echo "Cross-compilation examples:"
	@echo "  GOOS=linux GOARCH=amd64 make build    # Build for Linux x64"
	@echo "  GOOS=darwin GOARCH=arm64 make build   # Build for macOS ARM64"
	@echo "  make build-cross GOOS=windows GOARCH=amd64 # Generic cross-build"
	@echo ""
	@echo "GUI development:"
	@echo "  make gui-deps     # Install GUI dependencies (first time)"
	@echo "  make gui-dev      # Run GUI in development mode"
	@echo "  make gui-build    # Build GUI for production"
	@echo "  make gui-run      # Run the built GUI application"
	@echo ""
	@echo "CI targets:"
	@echo "  make ci-setup     # Show CI environment info"
	@echo "  make ci-test      # Run tests for CI"
	@echo "  make ci-lint      # Run linter for CI"
	@echo "  make ci-build-cli # Build CLI for CI"
	@echo "  make ci-build-desktop # Build desktop for CI"
	@echo ""
	@echo "Other targets:"
	@echo "  make test         # Run tests with coverage"
	@echo "  make run-pca-iris # Run PCA on iris dataset"
	@echo "  make clean        # Clean all artifacts"
	@echo "  make clean-cross  # Clean cross-compiled binaries only"