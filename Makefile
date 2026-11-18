.DEFAULT_GOAL := test

# Setup step for automated environments; the module needs nothing beyond the Go toolchain.
.PHONY: init
init:
	@echo "Nothing to install."

.PHONY: build
build:
	go build ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: check
check: build vet test

.PHONY: clean
clean:
	go clean -testcache
