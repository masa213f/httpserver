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
	goreleaser check

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath

.PHONY: install
install:
	CGO_ENABLED=0 go install -ldflags="-s -w" -trimpath

.PHONY: clean
clean:
	rm -rf dist testhttpserver

.PHONY: release-build
release-build:
	goreleaser release --snapshot --clean

.PHONY: setup
setup:
	go install github.com/goreleaser/goreleaser@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/goimports@latest