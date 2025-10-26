## Makefile for house_app_server
# Targets:
# - go-run: run the server
# - go-test: run unit tests
# - go-test-html: run tests and generate coverage.html
# - docs: run swag init to generate swagger docs
# - build: build the application binary
# - init: go mod tidy, install swag, download modules
# - go-clean: fmt, vet, mod tidy, and regen docs

.PHONY: go-run go-test go-test-html docs init go-clean build

GO := go
SWAG := swag
BINARY_NAME := house_app

go-run:
	$(GO) run .

go-test:
	$(GO) test -cover ./...

go-test-html:
	$(GO) test ./... -coverprofile=coverage.out -v
	$(GO) tool cover -html=coverage.out -o coverage.html

docs:
	$(SWAG) init -g main.go -o docs

build:
	$(GO) build -o $(BINARY_NAME) main.go

init:
	$(GO) mod tidy
	# Install swag CLI for generating docs
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) mod download

go-clean:
	$(GO) fmt ./...
	$(GO) vet ./...
	$(GO) mod tidy
	$(SWAG) init -g main.go -o docs
