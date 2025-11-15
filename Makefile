.PHONY: all build format fmt lint vuln test release check_clean clean help cov-integration cov-unit

# Build variables
VERSION    := $(shell git describe --tags --always --dirty)
COMMIT     := $(shell git rev-parse --short HEAD)
BUILDTIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
MOD_PATH   := $(shell go list -m)
APP_NAME   := gitree
GOCOVERDIR := ./covdatafiles

# Build targets
all: lint test build

build:
	CGO_ENABLED=0 \
	go build \
		-ldflags "-s -w" \
		-o bin/$(APP_NAME) \
		cmd/gitree/main.go

format:
	gofumpt -l -w .
	golangci-lint run --enable-only=nlreturn,godot,intrange --fix

fmt: format

lint: fmt
	go vet ./...
	staticcheck ./...
	golangci-lint run --show-stats

vuln:
	gosec ./...
	govulncheck

cov-integration:
	rm -fr "${GOCOVERDIR}" && mkdir -p "${GOCOVERDIR}"
	go build \
		-ldflags "-s -w" \
		-o bin/$(APP_NAME) \
		-cover \
		cmd/gitree/main.go
	go tool covdata percent -i=covdatafiles

cov-unit:
	rm -fr "${GOCOVERDIR}" && mkdir -p "${GOCOVERDIR}"
	go test -coverprofile="${GOCOVERDIR}/cover.out" ./...
	go tool cover -func="${GOCOVERDIR}/cover.out"
	go tool cover -html="${GOCOVERDIR}/cover.out"
	go tool cover -html="${GOCOVERDIR}/cover.out" -o "${GOCOVERDIR}/coverage.html"

test:
	go test ./...

check_clean:
	@if [ -n "$(shell git status --porcelain)" ]; then \
		echo "Error: Dirty working tree. Commit or stash changes before proceeding."; \
		exit 1; \
	fi

release-test: lint test vuln
	goreleaser check
	goreleaser release --snapshot --clean

release: check_clean lint test vuln
	goreleaser release --clean

clean:
	rm -rf bin/
	rm -rf ${GOCOVERDIR}
	go clean -cache -testcache -modcache
