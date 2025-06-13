.PHONY: build clean test install help

# Build the monday CLI
build:
	go build -o monday .

# Clean build artifacts
clean:
	rm -f monday

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the binary to GOPATH/bin
install: build
	cp monday $(GOPATH)/bin/monday

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the monday CLI binary"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  deps     - Install dependencies"
	@echo "  install  - Install binary to GOPATH/bin"
	@echo "  help     - Show this help message"

# Default target
all: deps build
