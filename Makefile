
##  LazyPoot — Makefile
##  Supports: ARM64 Termux, Linux x86-64, cross compile

BINARY   := lazypoot
MODULE   := lazypoot
PKG      := ./...
GOFLAGS  := -trimpath

VERSION  := 1.0.0
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
GIT_DIRTY  := $(shell git status --porcelain 2>/dev/null | head -1 | wc -l | sed 's/0/clean/; s/1/dirty/')
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := $(shell go version | awk '{print $$3}')

LDFLAGS  := -s -w \
	-X lazypoot/screens.Version=$(VERSION) \
	-X lazypoot/screens.GitCommit=$(GIT_COMMIT) \
	-X lazypoot/screens.BuildDate=$(BUILD_DATE) \
	-X lazypoot/screens.BuildGoVer=$(GO_VERSION) \
	-X lazypoot/screens.GitDirty=$(GIT_DIRTY)

.PHONY: all build run tidy clean install arm64

## Default target,tidy deps then build
all: tidy build

## Download dependencies (run once with internet access)
tidy:
	go mod tidy

## Build for the current platform
build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

## Build specifically for ARM64 (Termux / Android)
arm64:
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY)-arm64 .

## Run directly
run: build
	./$(BINARY)

## Install binary into PATH (Termux: $PREFIX/bin)
install: build
	install -Dm755 $(BINARY) $${PREFIX:-/usr/local}/bin/$(BINARY)

## Remove build artefacts
clean:
	rm -f $(BINARY) $(BINARY)-arm64

## Show module graph
graph:
	go mod graph

## Vet all packages
vet:
	go vet $(PKG)
