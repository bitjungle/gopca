# Makefile for complab - A professional-grade PCA toolkit
# This Makefile provides common development tasks for building, testing, and running the PCA toolkit

# Variables
BINARY_NAME := complab-cli
DESKTOP_NAME := complab-desktop
BUILD_DIR := build
CLI_PATH := cmd/complab-cli/main.go
DESKTOP_PATH := cmd/complab-desktop
COVERAGE_FILE := coverage.out

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOMOD := $(GOCMD) mod
GOGET := $(GOCMD) get

# Build flags
LDFLAGS := -ldflags="-s -w"

# Check if golangci-lint is installed
GOLINT := $(shell which golangci-lint 2> /dev/null)

# Check if wails is installed - check in PATH and common locations
WAILS := $(shell which wails 2> /dev/null || echo "$${HOME}/go/bin/wails")

# Default target
.DEFAULT_GOAL := all

# Phony targets
.PHONY: all build gui-dev gui-build gui-run gui-deps test test-verbose test-coverage fmt lint run-pca-iris clean install deps help

## all: Build the binary and run tests
all: build test

## build: Build the CLI binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CLI_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

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
	@if [ -f "$(DESKTOP_PATH)/build/bin/complab-desktop.app/Contents/MacOS/complab-desktop" ]; then \
		echo "Running GUI application..."; \
		open $(DESKTOP_PATH)/build/bin/complab-desktop.app; \
	elif [ -f "$(DESKTOP_PATH)/build/bin/complab-desktop" ]; then \
		echo "Running GUI application..."; \
		$(DESKTOP_PATH)/build/bin/complab-desktop; \
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

## help: Display this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ": "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sed 's/^## //'
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Example workflows:"
	@echo "  make              # Build CLI and test (default)"
	@echo "  make build        # Build the CLI binary"
	@echo "  make gui-deps     # Install GUI dependencies (first time)"
	@echo "  make gui-dev      # Run GUI in development mode"
	@echo "  make gui-build    # Build GUI for production"
	@echo "  make gui-run      # Run the built GUI application"
	@echo "  make test         # Run tests with coverage"
	@echo "  make run-pca-iris # Run PCA on iris dataset"
	@echo "  make clean        # Clean all artifacts"