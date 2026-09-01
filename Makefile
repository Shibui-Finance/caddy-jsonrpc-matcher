.DEFAULT_GOAL := test

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
