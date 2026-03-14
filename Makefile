
TARGET ?= cli

# Names/paths for each target
CLI_NAME := boxy
CLI_PATH := ./cmd/cli
DAEMON_NAME := boxyd
DAEMON_PATH := ./cmd/daemon

# Resolve GOBIN/GOPATH from `go env` so install target doesn't fall back to '/boxy'
GOBIN ?= $(shell go env GOBIN 2>/dev/null)
GOPATH := $(shell go env GOPATH 2>/dev/null)
ifeq ($(strip $(GOBIN)),)
GOBIN := $(GOPATH)/bin
endif

.PHONY: all build install clean uninstall build-cli build-daemon install-cli install-daemon

all: build

# Choose which package and binary name based on TARGET
ifeq ($(TARGET),daemon)
    BUILD_NAME := $(DAEMON_NAME)
    BUILD_PATH := $(DAEMON_PATH)
else
    BUILD_NAME := $(CLI_NAME)
    BUILD_PATH := $(CLI_PATH)
endif

build:
	go build -o $(BUILD_NAME) $(BUILD_PATH)

install: build
	install -m 0755 $(BUILD_NAME) $(GOBIN)/$(BUILD_NAME)

clean:
	rm -f $(CLI_NAME) $(DAEMON_NAME)

uninstall:
	rm -f $(GOBIN)/$(CLI_NAME) $(GOBIN)/$(DAEMON_NAME)

# Convenience explicit targets
build-cli:
	$(MAKE) build TARGET=cli

build-daemon:
	$(MAKE) build TARGET=daemon

install-cli:
	$(MAKE) install TARGET=cli

install-daemon:
	$(MAKE) install TARGET=daemon