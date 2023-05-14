BINDIR       := $(CURDIR)/bin
BINNAME      ?= cdev

LINUX_AMD64  := linux-amd64
LINUX_ARM64  := linux-arm64
DARWIN_AMD64 := darwin-amd64
DARWIN_ARM64 := darwin-arm64

CUR_GOOS     := $(shell go env GOOS)
CUR_GOARCH   := $(shell go env GOARCH)

GOPATH       := $(shell go env GOPATH)
GOIMPORTS    := $(GOPATH)/bin/goimports
ARCH         := $(shell uname -p)

SRC        	 := $(shell find . -type f -name '*.go' -print)

VERSION      = `git describe --tag --abbrev=0`
BUILD        = `date +%FT%T%z`
CONFIG_PKG   = "github.com/shalb/cluster.dev/pkg/config"

TLIST        = `cdev project create --list-templates`

# Required for globs to work correctly
SHELL      = /usr/bin/env bash

all: clean build

darwin_amd64:
	go get ./cmd/$(BINNAME) && GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" -o $(BINDIR)/$(DARWIN_AMD64)/$(BINNAME) ./cmd/$(BINNAME)

darwin_arm64:
	go get ./cmd/$(BINNAME) && GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" -o $(BINDIR)/$(DARWIN_ARM64)/$(BINNAME) ./cmd/$(BINNAME)

linux_amd64:
	go get ./cmd/$(BINNAME) && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" -o $(BINDIR)/$(LINUX_AMD64)/$(BINNAME) ./cmd/$(BINNAME)

linux_arm64:
	go get ./cmd/$(BINNAME) && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" -o $(BINDIR)/$(LINUX_ARM64)/$(BINNAME) ./cmd/$(BINNAME)

build: darwin_amd64 darwin_arm64 linux_amd64 linux_arm64
	@echo version: $(VERSION)

examples:
	rm -rf ./examples/*
	for tmpl in $(TLIST); do \
		cd examples ; \
    	mkdir $$tmpl ; \
		cd $$tmpl && cdev project create $$tmpl ; \
		cd ../ ; \
	done

install:
	go get ./cmd/$(BINNAME) && GO111MODULE=on CGO_ENABLED=0 GOOS=$(CUR_GOOS) GOARCH=$(CUR_GOARCH) go install -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" ./cmd/$(BINNAME)

quick-install:
	go fmt ./... && GO111MODULE=on CGO_ENABLED=0 GOOS=$(CUR_GOOS) GOARCH=$(CUR_GOARCH) go install -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" ./cmd/$(BINNAME)

clean:
	rm -rf $(BINDIR)/*

.PHONY: all clean examples
