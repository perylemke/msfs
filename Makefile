BINARY_NAME := msfs
GO := go
PKG := ./...
FS_PKG := ./internal/fs/...
COVERAGE_MIN := 90

.PHONY: all help build run repl script test test-fs coverage coverage-fs lint fmt vet tidy clean docker-build

all: tidy fmt vet test build

## help: List targets
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build msfs into bin/
build:
	@mkdir -p bin
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/msfs

## run: Build and run the REPL
run: build
	./bin/$(BINARY_NAME)

## repl: Run REPL without installing (go run)
repl:
	$(GO) run ./cmd/msfs

## script: Run PDF happy-path script (SCRIPT=path)
script:
	$(GO) run ./cmd/msfs run $(or $(SCRIPT),docs/scripts/pdf-happy-path.txt)

## test: Run all tests (race + shuffle)
test:
	$(GO) test -race -shuffle=on $(PKG)

## test-fs: Run filesystem package tests with coverage gate
test-fs:
	$(GO) test -race -shuffle=on -coverprofile=fs.out $(FS_PKG)
	@pct=$$($(GO) tool cover -func=fs.out | awk '/^total:/ {gsub(/%/,"",$$3); print $$3}'); \
	echo "internal/fs coverage: $${pct}%"; \
	awk -v p="$$pct" -v min="$(COVERAGE_MIN)" 'BEGIN {exit (p+0 >= min+0) ? 0 : 1}'

## coverage: Run all tests and show coverage (coverage.out + coverage.html)
coverage:
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out $(PKG)
	@echo ""
	@echo "Coverage by function:"
	@$(GO) tool cover -func=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "Line-level report: coverage.html (open in a browser)"

## coverage-fs: Coverage report for internal/fs only
coverage-fs:
	$(GO) test -coverprofile=fs.out $(FS_PKG)
	@$(GO) tool cover -func=fs.out

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: go fmt
fmt:
	$(GO) fmt $(PKG)

## vet: go vet
vet:
	$(GO) vet $(PKG)

## tidy: go mod tidy and verify
tidy:
	$(GO) mod tidy
	$(GO) mod verify

## clean: Remove build artifacts and coverage files
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html fs.out all.out cmd.out

## docker-build: Build Docker image (tag msfs:local)
docker-build:
	docker build -t msfs:local .
