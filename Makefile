BINARY_NAME := nginx-waf-api
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: all build test clean fmt lint

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/nginx-waf-api/

test:
	go test -v -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html

install: build
	install -D -m 755 $(BINARY_NAME) /usr/bin/$(BINARY_NAME)
