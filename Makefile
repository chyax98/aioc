BINARY ?= aioc
OUT ?= bin/$(BINARY)

.PHONY: test build check run agents doctor

test:
	go test ./...

build:
	mkdir -p bin
	go build -o $(OUT) ./cmd/aioc

check: test build

run:
	go run ./cmd/aioc run $(ARGS)

agents:
	go run ./cmd/aioc agents

doctor:
	go run ./cmd/aioc doctor
