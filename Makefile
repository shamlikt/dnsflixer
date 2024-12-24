
# Variables
APP_NAME = dnsflixer
CONFIG_FILE = config.toml
BUILD_DIR = build

# Go variables
GO = go
GOFMT = gofmt
SRC = $(wildcard *.go) dns/*.go httpserver/*.go

# Targets
.PHONY: all build run clean fmt install

all: build

build:
	@echo "Building $(APP_NAME)..."
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./

run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME) -config=$(CONFIG_FILE)

fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w $(SRC)

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

install:
	@echo "Installing dependencies..."
	$(GO) mod tidy

# Default target
.DEFAULT_GOAL := build
