BINARY := twitter-dore

.PHONY: build test lint fmt ci

build:
	mkdir -p bin
	go build -o bin/$(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

ci: lint test build
