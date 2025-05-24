.PHONY: test build build-all build-ubuntu20 build-ubuntu22 build-ubuntu24 build-macos build-macos-arm64 build-macos-amd64 build-local clean

# Test the application
test:
	go test ./internal/bkpfile -v

# Build for local development (current platform)
build-local:
	go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=$(shell go env GOOS)-$(shell go env GOARCH)'" -o bin/bkpfile ./cmd/bkpfile

# Build for all platforms
build-all: build-local build-macos build-ubuntu

# Build for Ubuntu
build-ubuntu: build-ubuntu20 build-ubuntu22 build-ubuntu24

# Build for Ubuntu 20.04
build-ubuntu20:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=linux-amd64-u20'" -o bin/bkpfile-ubuntu20.04 ./cmd/bkpfile

# Build for Ubuntu 22.04
build-ubuntu22:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=linux-amd64-u22'" -o bin/bkpfile-ubuntu22.04 ./cmd/bkpfile

# Build for Ubuntu 24.04
build-ubuntu24:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=linux-amd64-u24'" -o bin/bkpfile-ubuntu24.04 ./cmd/bkpfile

# Build for macOS (both ARM64 and AMD64)
build-macos: build-macos-arm64 build-macos-amd64

# Build for macOS ARM64 (Apple Silicon)
build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=darwin-arm64'" -o bin/bkpfile-macos-arm64 ./cmd/bkpfile

# Build for macOS AMD64 (Intel)
build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.compileDate=$(shell date -u +%Y-%m-%d\ %H:%M:%S\ UTC)' -X 'main.platform=darwin-amd64'" -o bin/bkpfile-macos-amd64 ./cmd/bkpfile

# Clean build artifacts
clean:
	rm -rf bin/

# Default build target (preserves original functionality)
build: build-local 