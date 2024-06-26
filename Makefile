# Project-specific settings
APP_NAME := kafka-producer-consumer-tester

BUILD_DIR := ./build
MAIN_DIR := ./cmd/$(APP_NAME)
PKG := github.com/raw-leak/$(APP_NAME)

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOSRC := $(GOBASE)/src

# Go files
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## build: Build the binary file for server
build:
	@echo "  >  Building binary..."
	@GOBIN=$(GOBIN) go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_DIR)

## run: Run the application in development mode
run: build
	@echo "  >  Running application..."
	@$(BUILD_DIR)/$(APP_NAME)

## fmt: Format the Go source code
fmt:
	@echo "  >  Formatting code..."
	@gofmt -w ${GOFMT_FILES}

## lint: Run linting checks
lint: fmt
	@echo "Running go vet..."
	@go vet ./...

	@echo "Running staticcheck..."
	@staticcheck ./...

## clean: Clean build files. Runs `go clean` internally
clean:
	@echo "  >  Cleaning build cache"
	@go clean

## rock: Runs produces-consumer test
rock:
	@echo "  >  Fasten your seatbelts, let's take of 🚀"
	./scripts/rock.sh

## rock-local: Runs produces-consumer test in local
rock-local:
	@echo "  >  Fasten your seatbelts, let's take of 🚀 in local"
	./scripts/rock-local.sh

## help: Show this help message
help: Makefile
	@echo
	@echo " Choose a command run in "$(APP_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: build fmt clean run run-local help
