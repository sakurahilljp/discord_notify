.PHONY: all build test clean install fmt lint help

BINARY_NAME=discord_notify
CMD_PATH=./cmd/discord_notify

all: build

build:
	go build -o $(BINARY_NAME) $(CMD_PATH)

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

install:
	go install $(CMD_PATH)

clean:
	rm -f $(BINARY_NAME)

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  test     - Run unit tests"
	@echo "  fmt      - Format source code"
	@echo "  lint     - Run go vet"
	@echo "  install  - Install binary to GOPATH/bin"
	@echo "  clean    - Remove built binary"
