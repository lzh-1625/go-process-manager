GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0
BINARY_NAME ?= gpm
VERSION ?= $(shell date +%Y%m%d%H%M)

SLIM ?= false # slim cannot support bleve
EPOLL ?= false

CMD_DIR := ./cmd/go_process_manager

LDFLAGS := -s -w -X main.Version=$(VERSION)

BUILD_FLAGS := -trimpath

BUILD_TAGS :=

ifeq ($(SLIM),true)
	BUILD_TAGS += slim
endif

ifeq ($(EPOLL),true)
ifeq ($(GOOS),linux)
	BUILD_TAGS += epoll
endif
endif

ifneq ($(strip $(BUILD_TAGS)),)
	TAGS := -tags="$(BUILD_TAGS)"
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
