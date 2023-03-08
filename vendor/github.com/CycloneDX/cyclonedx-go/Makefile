build:
	go build -v
.PHONY: build

test:
	go test -v -cover
.PHONY: test

clean:
	go clean
.PHONY: clean

generate:
	go generate
.PHONY: generate

lint:
	golangci-lint run
.PHONY: lint

all: clean build test
.PHONY: all
