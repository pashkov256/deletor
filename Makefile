VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

.PHONY: build
build:
	go build -ldflags "-X github.com/pashkov256/deletor/internal/version.Version=$(VERSION) \
		-X github.com/pashkov256/deletor/internal/version.BuildTime=$(BUILD_TIME) \
		-X github.com/pashkov256/deletor/internal/version.GitCommit=$(GIT_COMMIT)" \
		-o deletor ./cmd/deletor

.PHONY: run
run: build
	./deletor 