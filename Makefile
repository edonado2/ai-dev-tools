.PHONY: build test lint fmt vet clean install

# Build the binary
build:
	go build -o ai-dev ./cmd/ai-dev

# Run tests
test:
	go test ./... -v -race

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f ai-dev
	rm -f *.test
	rm -f *.out

# Install the binary
install: build
	go install ./cmd/ai-dev

# Run all checks
check: fmt vet lint test

# Generate mocks (if needed)
mocks:
	go generate ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy

# Run integration tests (requires docker)
integration-test:
	go test ./... -tags=integration -v

# Build for release (static linking)
release-build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ai-dev ./cmd/ai-dev

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install the binary"
	@echo "  check         - Run all checks (fmt, vet, lint, test)"
	@echo "  deps          - Download dependencies"
	@echo "  update-deps   - Update dependencies"
	@echo "  release-build - Build for release (static)"
	@echo "  help          - Show this help"