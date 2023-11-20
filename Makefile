.PHONY: all
all: help

## General

.PHONY: help
help:
	@echo "Choose one of the following target"
	@echo
	@echo "  fmt            Run go fmt against code."
	@echo "  vet            Run go vet against code."
	@echo "  build          Build all binaries."
	@echo "  install        Install all binaries."
	@echo "  setup          Install tools for development."
	@echo

.PHONY: fmt
fmt:
	goimports -w $$(find . -type f -name '*.go' -print)

.PHONY: lint
lint:
	test -z "$$(goimports -l $$(find . -type f -name '*.go' -print) | tee /dev/stderr)"
	staticcheck ./...
	go vet ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath

.PHONY: install
install:
	CGO_ENABLED=0 go install -ldflags="-s -w" -trimpath

.PHONY: setup
setup:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/goimports@latest