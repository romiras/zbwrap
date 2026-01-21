BINARY_NAME=zbwrap
BUILD_DIR=.
CMD_DIR=./cmd/zbwrap

# Build the binary
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Run all tests (unit + integration)
test:
	go test -v ./internal/... ./tests/...

# Run only unit tests (faster, no external deps mocked)
test-unit:
	go test -v ./internal/...

# Run only E2E tests
test-e2e:
	go test -v ./tests/...

# Clean build artifacts
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	go clean

# Format code
fmt:
	go fmt ./...

# Update dependencies
deps:
	go mod tidy

.PHONY: build test test-unit test-e2e clean fmt deps
