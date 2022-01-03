export PATH := $(PATH):$(shell go env GOPATH)/bin

# TODO: Check if used latest version
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0

# TODO: Check if used latest version
mock-install:
	go install github.com/golang/mock/mockgen@v1.6.0

generate:
	go generate ./...

lint:
	golangci-lint run

# TODO: [?] Test examples
test:
	go test -coverprofile cover.out \
	$(shell go list ./... | grep -v /examples/ | grep -v /test | grep -v /internal/ | grep -v /mock)

cover: test
	go tool cover -func cover.out

race:
	go test -race ./...

pre-commit: test lint

# TODO: Remove generator and fully replace it with generator-v2
# Usage: make generator RUN="types types-tests methods"
generator: ./internal/generator-v2
	go run ./internal/generator-v2 $$RUN

.PHONY: lint-install mock-install generate lint test cover race pre-commit generator
