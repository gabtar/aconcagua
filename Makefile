BINARY_NAME := aconcagua
MAIN_PACKAGE := ./cmd/$(BINARY_NAME)
OUTPUT_DIR := bin
GOAMD64=v1
MICROARCHS = v1 v2 v3 v4
PLATFORMS = linux darwin windows

# Architecture labels for main instructions set for each GOAMD64 version
LABEL_v1 = base
LABEL_v2 = popcnt
LABEL_v3 = avx2
LABEL_v4 = avx512

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

## test-cover: run all tests and display coverage
.PHONY: test-cover
test-cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
##			usage: make build GOAMD64=(v1|v2|v3|v4)
##			GOAMD64=v1 (default): The baseline. Exclusively generates instructions that all 64-bit x86 processors can execute.
##			GOAMD64=v2: all v1 instructions, plus CMPXCHG16B, LAHF, SAHF, POPCNT, SSE3, SSE4.1, SSE4.2, SSSE3.
##			GOAMD64=v3: all v2 instructions, plus AVX, AVX2, BMI1, BMI2, F16C, FMA, LZCNT, MOVBE, OSXSAVE.
##			GOAMD64=v4: all v3 instructions, plus AVX512F, AVX512BW, AVX512CD, AVX512DQ, AVX512VL.
.PHONY: build
build:
	@echo "--> Building $(BINARY_NAME) with GOAMD64=$(GOAMD64)..."
	GOARCH=amd64 GOOS=darwin  GOAMD64=$(GOAMD64) go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-x86_64  $(MAIN_PACKAGE)
	GOARCH=amd64 GOOS=linux   GOAMD64=$(GOAMD64) go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-x86_64   $(MAIN_PACKAGE)
	GOARCH=amd64 GOOS=windows GOAMD64=$(GOAMD64) go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-x86_64.exe $(MAIN_PACKAGE)

## build-all: build binaries for all os/architectures
.PHONY: build-all
build-all: $(foreach p,$(PLATFORMS),$(foreach m,$(MICROARCHS),build-$(p)-$(m)))

build-%:
	$(eval OS := $(word 1,$(subst -, ,$*)))
	$(eval M_ARCH := $(word 2,$(subst -, ,$*)))
	$(eval LABEL := $(LABEL_$(M_ARCH)))
	@echo "--> Building $(BINARY_NAME) for $(OS) [$(M_ARCH) -> $(LABEL)]..."
	GOARCH=amd64 GOOS=$(OS) GOAMD64=$(M_ARCH) go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-$(OS)-x86_64-$(LABEL)$(if $(filter windows,$(OS)),.exe,) $(MAIN_PACKAGE)

## run: run the  application
.PHONY: run
run:
	@echo "--> Running $(BINARY_NAME)..."
	go run $(MAIN_PACKAGE)
