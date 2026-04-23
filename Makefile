BINARY := site-audit

.PHONY: build run clean tidy fmt vet check help

## build: compile to ./site-audit
build:
	go build -o $(BINARY) .

## run: run directly with go run (no compile step)
run:
	go run .

## clean: remove compiled binary
clean:
	rm -f $(BINARY)

## tidy: sync go.mod and go.sum
tidy:
	go mod tidy

## fmt: format all Go source files
fmt:
	gofmt -w .

## vet: run static analysis
vet:
	go vet ./...

## check: fmt + vet (run before committing)
check: fmt vet

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## /  /'
