BINARY_NAME := weasel
INSTALL_PATH := /usr/local/bin
SCRIPTS_DIR := ./scripts

.PHONY: all build clean install uninstall format

all: build

build:
	@echo "Building..."
	@go build -o $(BINARY_NAME) ./cmd/weasel/main.go

run:
	@go run ./cmd/weasel/main.go

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

install: build
	@chmod +x $(SCRIPTS_DIR)/install.sh
	@$(SCRIPTS_DIR)/install.sh

uninstall:
	@chmod +x $(SCRIPTS_DIR)/uninstall.sh
	@$(SCRIPTS_DIR)/uninstall.sh

format:
	@go fmt ./...
