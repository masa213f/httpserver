.PHONY: all
all: build

.PHONY: fmt
fmt:
	goimports -w $$(find . -type f -name '*.go' -print)

.PHONY: lint
lint:
	test -z "$$(goimports -l $$(find . -type f -name '*.go' -print) | tee /dev/stderr)"
	staticcheck ./...
	go vet ./...
	goreleaser check

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

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