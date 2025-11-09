BINARY_NAME := aconcagua
MAIN_PACKAGE := ./cmd/$(BINARY_NAME)
OUTPUT_DIR := bin

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## format: format code and clean modfile
.PHONY: format
format:
	go fmt ./...

## checkhealth: run quality control checks
.PHONY: checkhealth
checkhealth:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
.PHONY: build
build:
	@echo "--> Building $(BINARY_NAME)..."
	GOARCH=amd64 GOOS=darwin go build -o $(OUTPUT_DIR)/${BINARY_NAME}-darwin $(MAIN_PACKAGE)
	GOARCH=amd64 GOOS=linux go build -o $(OUTPUT_DIR)/${BINARY_NAME}-linux $(MAIN_PACKAGE)
	GOARCH=amd64 GOOS=windows go build -o $(OUTPUT_DIR)/${BINARY_NAME}-windows $(MAIN_PACKAGE)

## run: run the  application
.PHONY: run
run:
	@echo "--> Running $(BINARY_NAME)..."
	go run $(MAIN_PACKAGE)
