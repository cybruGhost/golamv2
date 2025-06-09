MAKEFILE_VERSION := 1.0
BINARY_NAME=golamv2
PACKAGE_NAME=golamv2
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BUILD_DIR=BuiltBinaries
SRC_DIR=.
CMD_DIR=cmd
INTERNAL_DIR=internal
PKG_DIR=pkg

GOOS?=linux
GOARCH?=amd64
CGO_ENABLED=0

LINUX_GOOS=linux
LINUX_GOARCH=amd64
WINDOWS_GOOS=windows
WINDOWS_GOARCH=amd64
WINDOWS_EXT=.exe

GO_LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GO_GCFLAGS=-gcflags="-m=2"
GO_BUILDMODE=-buildmode=exe

BUILD_TAGS=netgo,osusergo
GO_FLAGS=-trimpath -mod=readonly

export GOOS
export GOARCH
export CGO_ENABLED

.PHONY: all
all: clean deps build

.PHONY: build-all
build-all: build-linux build-windows

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	go mod verify

.PHONY: build
build: build-linux

.PHONY: build-linux
build-linux: $(BUILD_DIR)/linux/$(BINARY_NAME)

$(BUILD_DIR)/linux/$(BINARY_NAME): $(shell find . -name '*.go' -type f)
	@echo "Building optimized Linux AMD64 binary..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=$(LINUX_GOOS) GOARCH=$(LINUX_GOARCH) go build $(GO_FLAGS) $(GO_LDFLAGS) $(GO_GCFLAGS) -tags="$(BUILD_TAGS)" -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(SRC_DIR)
	@echo "Linux binary built successfully: $(BUILD_DIR)/linux/$(BINARY_NAME)"
	@file $(BUILD_DIR)/linux/$(BINARY_NAME)
	@ls -lh $(BUILD_DIR)/linux/$(BINARY_NAME)

.PHONY: build-windows
build-windows: $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT)

$(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT): $(shell find . -name '*.go' -type f)
	@echo "Building optimized Windows AMD64 binary..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=$(WINDOWS_GOOS) GOARCH=$(WINDOWS_GOARCH) go build $(GO_FLAGS) $(GO_LDFLAGS) $(GO_GCFLAGS) -tags="$(BUILD_TAGS)" -o $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT) $(SRC_DIR)
	@echo "Windows binary built successfully: $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT)"
	@file $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT)
	@ls -lh $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT)

$(BUILD_DIR)/$(BINARY_NAME): $(BUILD_DIR)/linux/$(BINARY_NAME)
	@echo "Creating legacy symlink..."
	@ln -sf linux/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: build-ultra
build-ultra: build-ultra-linux build-ultra-windows

.PHONY: build-ultra-linux
build-ultra-linux:
	@echo "Building ultra-optimized Linux binary..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=$(LINUX_GOOS) GOARCH=$(LINUX_GOARCH) go build $(GO_FLAGS) \
		-ldflags="-s -w -linkmode external -extldflags '-static' -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)" \
		-gcflags="-m=2 -l=4" \
		-asmflags="-trimpath=$(PWD)" \
		-tags="$(BUILD_TAGS),static_build" \
		-o $(BUILD_DIR)/linux/$(BINARY_NAME)-ultra $(SRC_DIR)
	@echo "Ultra-optimized Linux binary built: $(BUILD_DIR)/linux/$(BINARY_NAME)-ultra"
	@file $(BUILD_DIR)/linux/$(BINARY_NAME)-ultra
	@ls -lh $(BUILD_DIR)/linux/$(BINARY_NAME)-ultra

.PHONY: build-ultra-windows
build-ultra-windows:
	@echo "Building ultra-optimized Windows binary..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=$(WINDOWS_GOOS) GOARCH=$(WINDOWS_GOARCH) go build $(GO_FLAGS) \
		-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)" \
		-gcflags="-m=2 -l=4" \
		-asmflags="-trimpath=$(PWD)" \
		-tags="$(BUILD_TAGS)" \
		-o $(BUILD_DIR)/windows/$(BINARY_NAME)-ultra$(WINDOWS_EXT) $(SRC_DIR)
	@echo "Ultra-optimized Windows binary built: $(BUILD_DIR)/windows/$(BINARY_NAME)-ultra$(WINDOWS_EXT)"
	@file $(BUILD_DIR)/windows/$(BINARY_NAME)-ultra$(WINDOWS_EXT)
	@ls -lh $(BUILD_DIR)/windows/$(BINARY_NAME)-ultra$(WINDOWS_EXT)

.PHONY: build-debug
build-debug: build-debug-linux build-debug-windows

.PHONY: build-debug-linux
build-debug-linux:
	@echo "Building Linux debug binary..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=$(LINUX_GOOS) GOARCH=$(LINUX_GOARCH) CGO_ENABLED=1 go build $(GO_FLAGS) \
		-ldflags="-X main.Version=$(VERSION)-debug -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)" \
		-gcflags="all=-N -l" \
		-race \
		-o $(BUILD_DIR)/linux/$(BINARY_NAME)-debug $(SRC_DIR)
	@echo "Linux debug binary built: $(BUILD_DIR)/linux/$(BINARY_NAME)-debug"

.PHONY: build-debug-windows
build-debug-windows:
	@echo "Building Windows debug binary..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=$(WINDOWS_GOOS) GOARCH=$(WINDOWS_GOARCH) CGO_ENABLED=1 go build $(GO_FLAGS) \
		-ldflags="-X main.Version=$(VERSION)-debug -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)" \
		-gcflags="all=-N -l" \
		-race \
		-o $(BUILD_DIR)/windows/$(BINARY_NAME)-debug$(WINDOWS_EXT) $(SRC_DIR)
	@echo "Windows debug binary built: $(BUILD_DIR)/windows/$(BINARY_NAME)-debug$(WINDOWS_EXT)"

.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race ./...

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean -cache
	go clean -testcache

.PHONY: install
install: build-linux
	@echo "Installing Linux binary..."
	cp $(BUILD_DIR)/linux/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

.PHONY: security
security:
	@echo "Running security check..."
	@which gosec > /dev/null || (echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...

.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@which govulncheck > /dev/null || (echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest" && exit 1)
	govulncheck ./...

.PHONY: build-info
build-info:
	@echo "=== Build Information ==="
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit Hash: $(COMMIT_HASH)"
	@echo "Linux Target: $(LINUX_GOOS)/$(LINUX_GOARCH)"
	@echo "Windows Target: $(WINDOWS_GOOS)/$(WINDOWS_GOARCH)"
	@echo "CGO Enabled: $(CGO_ENABLED)"
	@echo "Build Tags: $(BUILD_TAGS)"

.PHONY: profile-cpu
profile-cpu: build-linux
	@echo "Running CPU profiling..."
	go tool pprof -http=:8080 $(BUILD_DIR)/linux/$(BINARY_NAME)

.PHONY: profile-mem
profile-mem: build-linux
	@echo "Running memory profiling..."
	go tool pprof -http=:8081 -alloc_space $(BUILD_DIR)/linux/$(BINARY_NAME)

.PHONY: dev
dev: clean deps fmt vet test build-all

.PHONY: prod
prod: clean deps test lint security build-info build-all

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

.PHONY: release
release: prod
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/release
	
	cp $(BUILD_DIR)/linux/$(BINARY_NAME) $(BUILD_DIR)/release/
	tar -czf $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR)/release $(BINARY_NAME)
	@echo "Linux release package created: $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz"
	
	cp $(BUILD_DIR)/windows/$(BINARY_NAME)$(WINDOWS_EXT) $(BUILD_DIR)/release/
	zip -j $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/release/$(BINARY_NAME)$(WINDOWS_EXT)
	@echo "Windows release package created: $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip"
	
	rm -f $(BUILD_DIR)/release/$(BINARY_NAME) $(BUILD_DIR)/release/$(BINARY_NAME)$(WINDOWS_EXT)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Clean, install deps, and build (default)"
	@echo "  build         - Build optimized Linux AMD64 binary (default)"
	@echo "  build-all     - Build for both Linux and Windows"
	@echo "  build-linux   - Build optimized Linux AMD64 binary"
	@echo "  build-windows - Build optimized Windows AMD64 binary"
	@echo "  build-ultra   - Build with maximum optimizations (both platforms)"
	@echo "  build-debug   - Build with debug information (both platforms)"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Install and verify dependencies"
	@echo "  test          - Run tests"
	@echo "  benchmark     - Run benchmarks"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Lint code (requires golangci-lint)"
	@echo "  security      - Run security check (requires gosec)"
	@echo "  vuln-check    - Check for vulnerabilities (requires govulncheck)"
	@echo "  install       - Install Linux binary to GOPATH/bin"
	@echo "  dev           - Full development workflow (both platforms)"
	@echo "  prod          - Production build workflow (both platforms)"
	@echo "  release       - Create release packages for both platforms"
	@echo "  build-info    - Show build information"
	@echo "  help          - Show this help message"
