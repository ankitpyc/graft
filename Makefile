	# Go parameters
    GOCMD=go
    GOBUILD=$(GOCMD) build
    GOCLEAN=$(GOCMD) clean
    GOTEST=$(GOCMD) test
    GOGET=$(GOCMD) get
    MAIN_PACKAGE_PATH := ./cmd
    BINARY_NAME := bin/gocache
    # Version information
    VERSION=1.0.0
    APP_NAME := goread
    CURRENT_DIR := $(shell pwd)/bin
    INSTALL_DIR := /usr/local/bin


    # Build command
    build:
    	@echo $(shell pwd)/bin
    	$(GOBUILD) -o=${BINARY_NAME} ${MAIN_PACKAGE_PATH}

    # Clean command
    clean:
    	$(GOCLEAN)
    	rm -f $(BINARY_NAME)

    # Test command
    test:
    	$(GOTEST) -v ./...

    # Install dependencies
    deps:
    	$(GOGET) ./...

    # Release command
    release:
    	mkdir -p release
    	GOOS=linux GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME)-$(VERSION)-linux-amd64 -v
    	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME)-$(VERSION)-darwin-amd64 -v
    	GOOS=windows GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe -v


    # Help command
    help:
    	@echo "Available targets:"
    	@echo "  build       - Build the Go application"
    	@echo "  clean       - Clean up generated files"
    	@echo "  test        - Run tests"
    	@echo "  deps        - Install dependencies"
    	@echo "  release     - Create release builds for Linux, macOS, and Windows"

    # install command
    install:
    	sudo chmod +x ${INSTALL_DIR}/
    	sudo cp $(CURRENT_DIR)/${APP_NAME} ${INSTALL_DIR}

    # uninstall command
    uninstall:
    	sudo rm ${INSTALL_DIR}/${APP_NAME}
