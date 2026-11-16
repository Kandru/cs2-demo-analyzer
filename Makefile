# CS2 Demo Analyzer — all Go work runs through Docker so no local toolchain is needed.
SHELL := /bin/sh
.DEFAULT_GOAL := help

BINARY          := cs2-demo-analyzer
MODULE          := github.com/Kandru/cs2-demo-analyzer
VERSION         := $(shell cat VERSION 2>/dev/null || echo dev)
COMMIT          := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION      := 1.27.4
# Pinned by digest so CI and release builds use the exact audited toolchain.
DOCKER_GO_IMAGE ?= golang:$(GO_VERSION)@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280
HOST_UID        := $(shell id -u 2>/dev/null)
HOST_GID        := $(shell id -g 2>/dev/null)
DOCKER_USER_ARG := $(if $(HOST_UID),--user "$(HOST_UID):$(HOST_GID)")
CACHE_DIR       ?= $(CURDIR)/.cache
DIST            := dist
GOFLAGS         := -trimpath -buildvcs=false
LDFLAGS         := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(BUILD_DATE)

# Run a command inside the Go container with the repository mounted.
# The container runs as the invoking user so files written to dist/ and the
# repository stay owned by that user. Go caches live in .cache/ on the host.
DOCKER_RUN = docker run --rm \
	$(DOCKER_USER_ARG) \
	-v "$(CURDIR)":/workspace \
	-v "$(CACHE_DIR)":/cache \
	-w /workspace/src \
	-e CGO_ENABLED=0 \
	-e GOTOOLCHAIN=local \
	-e GOOS -e GOARCH \
	-e HOME=/tmp \
	-e GOPATH=/cache/go \
	-e GOMODCACHE=/cache/go-mod \
	-e GOCACHE=/cache/go-build \
	$(DOCKER_GO_IMAGE)

# Extra CLI flags for `make analyze` / `make watch`, e.g. ARGS="--json --index".
ARGS ?=

.PHONY: help build build-local test test-cover fmt vet tidy tidy-check audit release release-local docker-build run analyze watch version version-bump clean check-docker check-install

help: ## Show this help (default)
	@printf '\nCS2 Demo Analyzer — make targets\n\n'
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf '\nExamples:\n'
	@printf '  make build                         # linux/amd64 binary in dist/\n'
	@printf '  make test                          # go test ./... in Docker\n'
	@printf '  make audit                         # module verify + govulncheck\n'
	@printf '  make release                       # all release archives + SHA256SUMS\n'
	@printf '  make release-local                 # same, using the local Go toolchain\n'
	@printf '  make run ARGS="watch --input ./demos --output ./reports"\n'
	@printf '  make analyze ARGS="--json --index"  # analyze ./stuff into ./reports\n'
	@printf '  make watch ARGS="--cheat=false"    # watch ./stuff into ./reports\n'
	@printf '  make version-bump                 # bump YY.MM.BuildNr in VERSION\n\n'
	@printf 'Version: %s (%s)\n\n' '$(VERSION)' '$(COMMIT)'

check-docker:
	@command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed"; exit 1; }
	@mkdir -p "$(CACHE_DIR)/go" "$(CACHE_DIR)/go-mod" "$(CACHE_DIR)/go-build"

check-install: ## Validate install.sh syntax
	@sh -n install.sh

build: check-docker ## Build linux/amd64 binary into dist/ (Docker)
	@mkdir -p $(DIST)
	$(DOCKER_RUN) go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o /workspace/$(DIST)/$(BINARY) ./cmd/$(BINARY)
	@printf 'built %s/%s (%s)\n' '$(DIST)' '$(BINARY)' '$(VERSION)'

build-local: ## Build with a local Go toolchain (development only)
	@mkdir -p $(DIST)
	cd src && CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o ../$(DIST)/$(BINARY) ./cmd/$(BINARY)

test: check-docker check-install ## Run all tests in Docker
	$(DOCKER_RUN) go test ./...

test-cover: check-docker ## Run tests with coverage summary
	$(DOCKER_RUN) go test -cover ./...

fmt: check-docker ## Format all Go code
	$(DOCKER_RUN) gofmt -w .

vet: check-docker ## Run go vet in Docker
	$(DOCKER_RUN) go vet ./...

tidy: check-docker ## Run go mod tidy in Docker
	$(DOCKER_RUN) go mod tidy

tidy-check: check-docker ## Fail if go.mod/go.sum need tidying
	$(DOCKER_RUN) go mod tidy -diff

audit: check-docker ## Verify modules and scan for known vulnerabilities
	$(DOCKER_RUN) go mod verify
	$(DOCKER_RUN) go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...

release: check-docker ## Cross-compile all release archives into dist/ (Docker)
	@rm -rf $(DIST)
	@mkdir -p $(DIST)
	@sh hack/release.sh /workspace/$(DIST) $(DOCKER_RUN) go build $(GOFLAGS) -ldflags '$(LDFLAGS)'

release-local: ## Cross-compile all release archives into dist/ (local Go toolchain)
	@rm -rf $(DIST)
	@mkdir -p $(DIST)
	@sh hack/release.sh "$(CURDIR)/$(DIST)" go build $(GOFLAGS) -ldflags '$(LDFLAGS)'

docker-build: check-docker ## Build the runtime Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY):$(VERSION) -t $(BINARY):latest .

run: check-docker ## Run the container with ./demos and ./reports mounted
	@mkdir -p demos reports
	docker run --rm -it $(DOCKER_USER_ARG) \
		-v "$(CURDIR)/demos":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest $(ARGS)

analyze: docker-build ## Analyze every demo in ./stuff into ./reports (ARGS="--json --index")
	@mkdir -p reports
	docker run --rm $(DOCKER_USER_ARG) \
		-v "$(CURDIR)/stuff":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest analyze --input /demos --output /out $(ARGS)

watch: docker-build ## Watch ./stuff and write reports into ./reports (ARGS="--cheat=false")
	@mkdir -p reports
	docker run --rm -it $(DOCKER_USER_ARG) \
		-v "$(CURDIR)/stuff":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest watch --input /demos --output /out $(ARGS)

version: ## Print the current version
	@cat VERSION

version-bump: ## Increment YY.MM.BuildNr (resets on month change)
	@sh hack/bump-version.sh

clean: ## Remove build output and local caches
	@chmod -R u+w .cache 2>/dev/null || true
	rm -rf $(DIST) .cache
