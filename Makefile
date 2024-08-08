# Define the Go binary name
BINARY_NAME=myapp

# Default target
all: build

.PHONY: lint test vendor clean

export GO111MODULE=on

default: lint test

lint:
	golangci-lint run

# Build the Go application
build:
	go build -o $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run tests and generate coverage report
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

yaegi_test:
	yaegi test -v .

vendor:
	go mod vendor

# Clean up build artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html

# Run the application
run: build
	./$(BINARY_NAME)

# Format Go source code
fmt:
	go fmt ./...

# Lint Go source code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Help target to show available commands
help:
	@echo "Available targets:"
	@echo "  all         - Build the application"
	@echo "  build       - Build the Go application"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests and generate coverage report"
	@echo "  clean       - Clean up build artifacts"
	@echo "  run         - Build and run the application"
	@echo "  fmt         - Format Go source code"
	@echo "  lint        - Lint Go source code"
	@echo "  help        - Show this help message"
