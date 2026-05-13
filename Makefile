GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0
BINARY_NAME ?= gpm
VERSION ?= $(shell date +%Y%m%d%H%M)

SLIM ?= false # slim cannot support bleve

CMD_DIR := ./cmd/go_process_manager

LDFLAGS := -s -w -X main.Version=$(VERSION)

BUILD_FLAGS := -trimpath

ifeq ($(SLIM),true)
	TAGS := -tags="slim"
endif

.PHONY: build clean

build:
	@echo "Building $(BINARY_NAME)..."
	@echo "Version=$(VERSION)"

	CGO_ENABLED=$(CGO_ENABLED) \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(TAGS) \
	-ldflags="$(LDFLAGS)" \
	$(BUILD_FLAGS) \
	-o $(BINARY_NAME) \
	$(CMD_DIR)

clean:
	rm -f $(BINARY_NAME)