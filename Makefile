.DEFAULT_GOAL := test

# runtime's software factory runs `make init` in every job's clone and in its image build.
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
