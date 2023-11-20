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
	@echo

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath

.PHONY: install
install:
	CGO_ENABLED=0 go install -ldflags="-s -w" -trimpath