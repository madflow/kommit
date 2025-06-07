.PHONY: build test

# Build the binary
build:
	@echo "Building kommit..."
	@go build 

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
