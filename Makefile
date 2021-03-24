BINDIR      := $(CURDIR)/bin
BINNAME     ?= cdev

GOPATH        = $(shell go env GOPATH)
GOIMPORTS     = $(GOPATH)/bin/goimports
ARCH          = $(shell uname -p)

SRC        := $(shell find . -type f -name '*.go' -print)

VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`
CONFIG_PKG="github.com/shalb/cluster.dev/pkg/config"

# Required for globs to work correctly
SHELL      = /usr/bin/env bash

.PHONY: all
all: build

.PHONY: build
build:
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" -o $(BINDIR)/$(BINNAME) ./cmd/$(BINNAME)

.PHONY: install
install:
	GO111MODULE=on CGO_ENABLED=0 go install -ldflags "-w -s -X ${CONFIG_PKG}.Version=${VERSION} -X ${CONFIG_PKG}.BuildTimestamp=${BUILD}" ./cmd/$(BINNAME)

.PHONY: clean
clean:
	rm $(BINDIR)/$(BINNAME)
