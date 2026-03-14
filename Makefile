
CLI_NAME=boxy
CLI_PATH=./cmd/cli

# Resolve GOBIN/GOPATH from `go env` so install target doesn't fall back to '/boxy'
GOBIN ?= $(shell go env GOBIN 2>/dev/null)
GOPATH := $(shell go env GOPATH 2>/dev/null)
ifeq ($(strip $(GOBIN)),)
GOBIN := $(GOPATH)/bin
endif

.PHONY: all build install clean uninstall

all: build

build:
	go build -o $(CLI_NAME) $(CLI_PATH)

install: build
	install -m 0755 $(CLI_NAME) $(GOBIN)/$(CLI_NAME)

clean:
	rm -f $(CLI_NAME)

uninstall:
	rm -f $(GOBIN)/$(CLI_NAME)