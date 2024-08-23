# Variables
BINARY_NAME := terraform-provider-daw
VERSION := 0.1.0
BUILD_DIR := ./bin
SOURCE_DIR := .

# Architectures
ARCHITECTURES := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Build targets
.PHONY: all build clean test

all: build

build: $(ARCHITECTURES)
	@echo "Build completed for all architectures."

$(ARCHITECTURES):
	@echo "Building $(BINARY_NAME) version $(VERSION) for $@..."
	@GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
	go build -o $(BUILD_DIR)/$(BINARY_NAME)_v$(VERSION)_$(subst /,_,$@) $(SOURCE_DIR)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	@go test -v $(SOURCE_DIR)/...