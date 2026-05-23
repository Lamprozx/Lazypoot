# LazyPoot Makefile
# Supports: Linux x86-64, ARM64 Termux, cross compile


BINARY      ?= lazypoot
MODULE      ?= lazypoot
PKG         ?= ./...
GO          ?= go
GOFLAGS     ?= -trimpath

VERSION     ?= 1.0.0
GIT_COMMIT  := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
GIT_DIRTY   := $(shell git diff --quiet HEAD 2>/dev/null && echo clean || echo dirty)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION  := $(shell $(GO) version | awk '{print $$3}')

LDFLAGS     := -s -w \
	-X $(MODULE)/screens.Version=$(VERSION) \
	-X $(MODULE)/screens.GitCommit=$(GIT_COMMIT) \
	-X $(MODULE)/screens.BuildDate=$(BUILD_DATE) \
	-X $(MODULE)/screens.BuildGoVer=$(GO_VERSION) \
	-X $(MODULE)/screens.GitDirty=$(GIT_DIRTY)

.PHONY: all build run install clean fmt graph vet arm64 lint tidy test help

.DEFAULT_GOAL := build

all: build vet

tidy:
	$(GO) mod tidy

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

arm64:
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)-arm64 .

run: build
	./$(BINARY)

install: build
	install -Dm755 $(BINARY) $${PREFIX:-/usr/local}/bin/$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY)-arm64

fmt:
	$(GO) fmt $(PKG)

graph:
	$(GO) mod graph

vet:
	$(GO) vet $(PKG)

lint:
	$(GO) vet $(PKG)

test:
	$(GO) test $(PKG)

help:
	@echo "Targets:"
	@echo "  all       Build and vet (default)"
	@echo "  build     Build for current platform"
	@echo "  arm64     Cross compile for ARM64 (Termux)"
	@echo "  run       Build and execute"
	@echo "  install   Build and install to PREFIX/bin"
	@echo "  clean     Remove build artefacts"
	@echo "  fmt       Run go fmt"
	@echo "  vet       Run go vet"
	@echo "  lint      Alias for vet"
	@echo "  test      Run tests"
	@echo "  tidy      Tidy go.mod and go.sum"
	@echo "  graph     Show module dependency graph"
