GO := $(shell which go)

.PHONY:
	run

all: fmt vet build

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

run: RUN_ARGS=--help
run: fmt vet
	$(GO) run ./cmd/tea $(RUN_ARGS)

build: BUILD_OUTPUT=./bin/tea
build:
	CGO_ENABLED=0 $(GO) build -mod=readonly -a -o $(BUILD_OUTPUT) ./cmd/tea
