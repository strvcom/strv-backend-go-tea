GO := $(shell which go)

APP_VERSION ?= "v0.0.0"

.PHONY: all
all: fmt vet build

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: run
run: RUN_ARGS=--help
run: fmt vet
	$(GO) run ./cmd/tea $(RUN_ARGS)

.PHONY: test
test: generate
	$(GO) test ./... -cover

.PHONY: build
build: BUILD_OUTPUT=./bin/tea
build: generate
	CGO_ENABLED=0 $(GO) build -ldflags "-X main.version=$(APP_VERSION)" -mod=readonly -o $(BUILD_OUTPUT) ./cmd/tea

.PHONY: generate
generate:
	$(GO) generate ./...
