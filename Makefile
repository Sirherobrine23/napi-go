GO ?= go
NODE ?= node
BUILD_DIR ?= build
NODE_INCLUDE ?=
GO_TEST_FLAGS ?= -count=1

# NODE_INCLUDE is the parent directory of node/node_api.h, e.g. /usr/include
# or the include directory inside an official Node headers archive.
ifneq ($(strip $(NODE_INCLUDE)),)
override CGO_CPPFLAGS += -I "$(NODE_INCLUDE)"
export CGO_CPPFLAGS
endif

# The Go integration runner uses these executables for its child processes too.
export NAPI_GO_GO := $(GO)
export NAPI_GO_NODE := $(NODE)

EXAMPLES := class threadsafe eventemitter
EXAMPLE_ADDONS := $(addprefix $(BUILD_DIR)/,$(addsuffix .node,$(EXAMPLES)))
JS_TESTS := internal/integration/integration.cjs internal/integration/eventemitter.cjs

.DEFAULT_GOAL := build
.PHONY: all build examples test test-integration test-eventemitter test-js
.PHONY: test-cgo-check test-race check fmt vet clean help FORCE check-node build-dir

all: build

build:
	$(GO) build ./...

examples: $(EXAMPLE_ADDONS)

$(BUILD_DIR)/%.node: FORCE | build-dir
	$(GO) build -buildmode=c-shared -o "$@" "./internal/examples/$*"

$(BUILD_DIR)/integration.node: FORCE | build-dir
	$(GO) build -buildvcs=false -buildmode=c-shared -o "$@" ./internal/integration/addon

build-dir:
	mkdir -p "$(BUILD_DIR)"

check-node:
	@command -v "$(NODE)" >/dev/null 2>&1 || { \
		printf '%s\n' 'Node.js is required; set NODE to its executable path.' >&2; exit 1; \
	}

test: check-node
	$(GO) test -v $(GO_TEST_FLAGS) ./...

test-integration: check-node
	$(GO) test -v $(GO_TEST_FLAGS) -v ./internal/integration

# Run only the EventEmitter suite against a freshly built addon.
test-eventemitter: check-node $(BUILD_DIR)/integration.node
	NAPI_GO_ADDON="$(abspath $(BUILD_DIR)/integration.node)" \
		"$(NODE)" --expose-gc --test internal/integration/eventemitter.cjs

test-js: check-node $(BUILD_DIR)/integration.node
	NAPI_GO_ADDON="$(abspath $(BUILD_DIR)/integration.node)" \
		"$(NODE)" --expose-gc --test $(JS_TESTS)

test-cgo-check: check-node
	GOEXPERIMENT=cgocheck2 $(GO) test $(GO_TEST_FLAGS) -v ./internal/integration

# Instrument the addon itself, not just the outer Go test process.
test-race: check-node
	NAPI_GO_TEST_RACE=1 $(GO) test $(GO_TEST_FLAGS) -v ./internal/integration

check: build test test-cgo-check

fmt:
	$(GO) fmt ./...

# Kept separate from check: three existing raw-finalizer warnings remain.
vet:
	$(GO) vet ./...

# Remove only this Makefile's known outputs, even with a custom BUILD_DIR.
clean:
	rm -f $(foreach name,$(EXAMPLES) integration,"$(BUILD_DIR)/$(name).node" "$(BUILD_DIR)/$(name).h")
	@if [ -d "$(BUILD_DIR)" ]; then rmdir "$(BUILD_DIR)" 2>/dev/null || :; fi

help:
	@printf '%s\n' \
		'build               Compile all Go packages (default)' \
		'examples            Build class, threadsafe and eventemitter addons' \
		'test                Run all Go tests and both Node integration suites' \
		'test-integration    Run both Node suites through the Go test runner' \
		'test-eventemitter   Run only the EventEmitter suite' \
		'test-js             Build an addon and run both Node suites directly' \
		'test-cgo-check      Run integration with GOEXPERIMENT=cgocheck2' \
		'test-race           Build the addon with -race and run integration' \
		'check               Build, test and strict cgo validation' \
		'fmt / vet / clean   Format, inspect or remove generated outputs' \
		'Variables: GO, NODE, NODE_INCLUDE, BUILD_DIR, GO_TEST_FLAGS'

FORCE:
