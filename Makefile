BINARY_NAME=protoc-gen-connect-errors
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build test lint clean install proto-gen

all: test build

## Build the protoc plugin binary
build:
	@echo "Building $(BINARY_NAME)..."
	cd cmd/protoc-gen-connect-errors && go build $(LDFLAGS) -o ../../bin/$(BINARY_NAME) .

## Run all tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -cover -count=1 ./...

## Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/

## Install the plugin binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

## Generate code from proto files (requires buf)
proto-gen:
	@echo "Generating proto code..."
	buf generate proto/

## Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .

## Show help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""
	@grep -E '^[a-zA-Z_-]+:' Makefile | sed 's/:.*//' | sort
