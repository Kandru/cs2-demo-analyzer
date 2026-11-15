# CS2 Battlelog — all Go work runs through Docker so no local toolchain is needed.
SHELL := /bin/sh
.DEFAULT_GOAL := help

BINARY          := cs2-battlelog
MODULE          := github.com/Kandru/cs2-demo-analyzer
VERSION         := $(shell cat VERSION 2>/dev/null || echo dev)
COMMIT          := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION      := 1.27
DOCKER_GO_IMAGE ?= golang:$(GO_VERSION)
GO_MOD_CACHE    ?= cs2-battlelog-go-mod
GO_BUILD_CACHE  ?= cs2-battlelog-go-build
DIST            := dist
GOFLAGS         := -trimpath -buildvcs=false
LDFLAGS         := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(BUILD_DATE)

# Run a command inside the Go container with the repository mounted.
# The two named volumes make repeated builds fast without polluting the repo.
DOCKER_RUN = docker run --rm \
	-v "$(CURDIR)":/workspace \
	-v $(GO_MOD_CACHE):/go/pkg/mod \
	-v $(GO_BUILD_CACHE):/root/.cache/go-build \
	-w /workspace/src \
	-e CGO_ENABLED=0 \
	$(DOCKER_GO_IMAGE)

RELEASE_TARGETS := linux/amd64 linux/arm64 linux/386 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

.PHONY: help build build-local test test-cover fmt vet tidy release docker-build run analyze watch version version-bump clean check-docker

help: ## Show this help (default)
	@printf '\nCS2 Battlelog — make targets\n\n'
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf '\nExamples:\n'
	@printf '  make build                         # linux/amd64 binary in dist/\n'
	@printf '  make test                          # go test ./... in Docker\n'
	@printf '  make release                       # all release archives + SHA256SUMS\n'
	@printf '  make run ARGS="watch --input ./demos --output ./reports"\n'
	@printf '  make version-bump                 # bump YY.MM.BuildNr in VERSION\n\n'
	@printf 'Version: %s (%s)\n\n' '$(VERSION)' '$(COMMIT)'

check-docker:
	@command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed"; exit 1; }

build: check-docker ## Build linux/amd64 binary into dist/ (Docker)
	@mkdir -p $(DIST)
	$(DOCKER_RUN) go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o /workspace/$(DIST)/$(BINARY) ./cmd/$(BINARY)
	@printf 'built %s/%s (%s)\n' '$(DIST)' '$(BINARY)' '$(VERSION)'

build-local: ## Build with a local Go toolchain (development only)
	@mkdir -p $(DIST)
	cd src && CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o ../$(DIST)/$(BINARY) ./cmd/$(BINARY)

test: check-docker ## Run all tests in Docker
	$(DOCKER_RUN) go test ./...

test-cover: check-docker ## Run tests with coverage summary
	$(DOCKER_RUN) go test -cover ./...

fmt: check-docker ## Format all Go code
	$(DOCKER_RUN) gofmt -w .

vet: check-docker ## Run go vet in Docker
	$(DOCKER_RUN) go vet ./...

tidy: check-docker ## Run go mod tidy in Docker
	$(DOCKER_RUN) go mod tidy

release: check-docker ## Cross-compile all release archives into dist/ (Docker)
	@mkdir -p $(DIST)
	@for target in $(RELEASE_TARGETS); do \
		os=$${target%/*}; arch=$${target#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		name="$(BINARY)-$(VERSION)-$$os-$$arch"; \
		printf 'building %s...\n' "$$name"; \
		$(DOCKER_RUN) go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o /workspace/$(DIST)/$$name/$(BINARY)$$ext ./cmd/$(BINARY) || exit 1; \
		cp README.md VERSION $(DIST)/$$name/ 2>/dev/null || true; \
		( cd $(DIST) && tar -czf "$$name.tar.gz" "$$name" && rm -rf "$$name" ) || exit 1; \
	done
	@cd $(DIST) && sha256sum *.tar.gz > SHA256SUMS && printf '\nRelease %s artifacts:\n' '$(VERSION)' && ls -lh *.tar.gz SHA256SUMS

docker-build: check-docker ## Build the runtime Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(BINARY):$(VERSION) -t $(BINARY):latest .

run: check-docker ## Run the container with ./demos and ./reports mounted
	@mkdir -p demos reports
	docker run --rm -it \
		-v "$(CURDIR)/demos":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest $(ARGS)

analyze: docker-build ## Analyze every demo in ./stuff into ./reports
	@mkdir -p reports
	docker run --rm \
		-v "$(CURDIR)/stuff":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest analyze --input /demos --output /out

watch: docker-build ## Watch ./stuff and write reports into ./reports
	@mkdir -p reports
	docker run --rm -it \
		-v "$(CURDIR)/stuff":/demos:ro \
		-v "$(CURDIR)/reports":/out \
		$(BINARY):latest watch --input /demos --output /out

version: ## Print the current version
	@cat VERSION

version-bump: ## Increment YY.MM.BuildNr (resets on month change)
	@sh hack/bump-version.sh

clean: ## Remove build output and local caches
	rm -rf $(DIST) .cache
