GO ?= go
BINARY ?= aioc
OUT ?= bin/$(BINARY)

.PHONY: test build check proto-check run agents doctor

test:
	$(GO) test ./...

build:
	mkdir -p bin
	$(GO) build -o $(OUT) ./cmd/aioc

check: test build

proto-check:
	proto exec go -- make check GO=go

run:
	$(GO) run ./cmd/aioc run $(ARGS)

agents:
	$(GO) run ./cmd/aioc agents

doctor:
	$(GO) run ./cmd/aioc doctor
