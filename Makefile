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

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: build
build:
	CGO_ENABLED=0 go build -o httpserver -trimpath -ldflags "-s -w" .

.PHONY: clean
clean:
	rm -rf httpserver

.PHONY: setup
setup:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
