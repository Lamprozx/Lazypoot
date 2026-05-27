# LazyPoot Makefile

BINARY      ?= lazypoot
MODULE      ?= lazypoot
PKG         ?= ./...
GO          ?= go

VERSION     ?= 2.1.0
GOFLAGS     ?= -trimpath -buildvcs=false
TAGS        ?= osusergo netgo

GIT_COMMIT  := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
GIT_DIRTY   := $(shell git diff --quiet HEAD 2>/dev/null && echo clean || echo dirty)
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION  := $(shell $(GO) version | awk '{print $$3}')

LDFLAGS := -s -w \
	-X $(MODULE)/screens.Version=$(VERSION) \
	-X $(MODULE)/screens.GitCommit=$(GIT_COMMIT) \
	-X $(MODULE)/screens.BuildDate=$(BUILD_DATE) \
	-X $(MODULE)/screens.BuildGoVer=$(GO_VERSION) \
	-X $(MODULE)/screens.GitDirty=$(GIT_DIRTY)

.PHONY: all build run install clean fmt vet lint test tidy \
	termux termux-arm64 termux-armhf \
	linux-amd64 linux-arm64 linux-armhf help

.DEFAULT_GOAL := build

all: build vet

build:
	CGO_ENABLED=0 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY) .

run: build
	./$(BINARY)

install: build
	install -Dm755 $(BINARY) $${PREFIX:-/usr/local}/bin/$(BINARY)

clean:
	rm -f \
	$(BINARY) \
	$(BINARY)-termux \
	$(BINARY)-arm64 \
	$(BINARY)-armhf \
	$(BINARY)-linux-amd64 \
	$(BINARY)-linux-arm64 \
	$(BINARY)-linux-armhf

fmt:
	$(GO) fmt $(PKG)

vet:
	CGO_ENABLED=0 $(GO) vet $(PKG)

lint: vet

test:
	CGO_ENABLED=0 $(GO) test $(PKG)

tidy:
	$(GO) mod tidy

# Native Termux build
termux:
	CGO_ENABLED=0 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY)-termux .

# Android / Termux ARM64
termux-arm64:
	CGO_ENABLED=0 GOOS=android GOARCH=arm64 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY)-arm64 .

# Android / Termux ARMHF
# Build ini paling aman langsung di Termux
termux-armhf:
	CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 CC=clang $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS) -linkmode external" \
	-o $(BINARY)-armhf .

# Linux AMD64
linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY)-linux-amd64 .

# Linux ARM64
linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY)-linux-arm64 .

# Linux ARMHF
linux-armhf:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build \
	$(GOFLAGS) \
	-tags "$(TAGS)" \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY)-linux-armhf .

help:
	@echo "Targets:"
	@echo "  build          Build native"
	@echo "  run            Build and run"
	@echo "  install        Install binary"
	@echo "  clean          Remove artifacts"
	@echo "  fmt            Run go fmt"
	@echo "  vet            Run go vet"
	@echo "  lint           Alias for vet"
	@echo "  test           Run tests"
	@echo "  tidy           Run go mod tidy"
	@echo ""
	@echo "Android / Termux:"
	@echo "  termux         Native Termux build"
	@echo "  termux-arm64   Android ARM64"
	@echo "  termux-armhf   Android ARMv7"
	@echo ""
	@echo "Linux:"
	@echo "  linux-amd64    Linux x86_64"
	@echo "  linux-arm64    Linux ARM64"
	@echo "  linux-armhf    Linux ARMv7"
