.PHONY: all build format fmt lint security test-coverage test-coverage-report test-race fuzz fuzz-long release check_clean clean help

# Build variables
VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse --short HEAD)
BUILDTIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
MOD_PATH   := $(shell go list -m)
APP_NAME   := gitree
GOCOVERDIR := ./coverage

# Build targets
all: lint test build

build:
	CGO_ENABLED=0 \
	go build \
		-ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.buildTime=$(BUILDTIME)'" \
		-o bin/$(APP_NAME) \
		./cmd/gitree

format:
	gofumpt -l -w ./cmd ./internal
	golangci-lint run --enable-only=nlreturn,godot,intrange --fix

fmt: format

lint: fmt
	go vet ./...
	staticcheck ./...
	golangci-lint run --show-stats

security:
	gosec ./...
	govulncheck

test:
	go test ./...

test-all: test test-coverage test-race test-fuzz

test-coverage:
	rm -fr "${GOCOVERDIR}" && mkdir -p "${GOCOVERDIR}"
	go test -coverprofile="${GOCOVERDIR}/cover.out" ./...
	go tool cover -func="${GOCOVERDIR}/cover.out"

test-coverage-report: test-coverage
	go tool cover -html="${GOCOVERDIR}/cover.out"
	go tool cover -html="${GOCOVERDIR}/cover.out" -o "${GOCOVERDIR}/coverage.html"

test-race:
	go test -race ./...

test-fuzz:
	go test -fuzz=FuzzGitStatusFormat -fuzztime=15s ./internal/models/
	go test -fuzz=FuzzGitStatusValidate -fuzztime=15s ./internal/models/

check_clean:
	@if [ -n "$(shell git status --porcelain)" ]; then \
		echo "Error: Dirty working tree. Commit or stash changes before proceeding."; \
		exit 1; \
	fi

release-test: lint test security
	goreleaser check
	goreleaser release --snapshot --clean

release: check_clean lint test security
	goreleaser release --clean

clean:
	rm -rf bin/
	rm -rf ${GOCOVERDIR}
	go clean -cache -testcache -modcache
