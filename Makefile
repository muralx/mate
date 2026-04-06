.PHONY: fmt lint vet test cover build check

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

test:
	go test -race ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

build:
	go build ./...

check: fmt vet test build
