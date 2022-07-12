GO := $(shell which go)

APP_VERSION ?= "v0.0.0"

.PHONY:
	run  \
	test \
	build

all: fmt vet build

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

run: RUN_ARGS=--help
run: fmt vet
	$(GO) run ./cmd/tea $(RUN_ARGS)

test: generate
	$(GO) test ./... -cover

build: BUILD_OUTPUT=./bin/tea
build: generate
	CGO_ENABLED=0 $(GO) build -ldflags "-X main.version=$(APP_VERSION)" -mod=readonly -o $(BUILD_OUTPUT) ./cmd/tea

generate:
	$(GO) generate ./...
