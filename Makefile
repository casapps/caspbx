PROJECTNAME := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)(\.git)?$$|\1|' || basename "$$(pwd)")
PROJECTORG := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)/[^/]+(\.git)?$$|\1|' || basename "$$(dirname "$$(pwd)")")

VERSION ?= $(shell cat release.txt 2>/dev/null || echo "0.1.0")
BUILD_DATE := $(shell date +"%B %-d, %Y at %H:%M:%S")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
OFFICIALSITE ?= $(shell [ -f site.txt ] && cat site.txt || echo "")
MAINTAINER_NAME ?= casappps
MAINTAINER_EMAIL ?= docker-admin@casjaysdev.pro
PLATFORM_REPO_URL ?= $(shell git remote get-url origin 2>/dev/null | sed -E 's#^git@([^:]+):#https://\1/#; s#\.git$$##' || echo "")
LICENSE_NAME ?= MIT

LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.CommitID=$(COMMIT_ID)' \
	-X 'main.BuildDate=$(BUILD_DATE)' \
	-X 'main.OfficialSite=$(OFFICIALSITE)'

BINDIR := binaries
RELDIR := releases

GODIR := $(HOME)/.local/share/go
GOCACHE := $(HOME)/.local/share/go/build

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64 freebsd/arm64

REGISTRY ?= ghcr.io/$(PROJECTORG)/$(PROJECTNAME)
GO_DOCKER := docker run --rm \
	-v $(PWD):/build \
	-v $(GOCACHE):/root/.cache/go-build \
	-v $(GODIR):/go \
	-w /build \
	-e CGO_ENABLED=0 \
	golang:alpine

.PHONY: build local release docker test dev clean

build: clean
	@mkdir -p $(BINDIR) $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod tidy
	@$(GO_DOCKER) go mod download
	@$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME) ./src"
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		OUTPUT=$(BINDIR)/$(PROJECTNAME)-$$OS-$$ARCH; \
		[ "$$OS" = "windows" ] && OUTPUT=$$OUTPUT.exe; \
		$(GO_DOCKER) sh -c "GOOS=$$OS GOARCH=$$ARCH go build -ldflags \"$(LDFLAGS)\" -o $$OUTPUT ./src" || exit 1; \
	done

local: clean
	@mkdir -p $(BINDIR) $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod tidy
	@$(GO_DOCKER) go mod download
	@$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME) ./src"

release: build
	@mkdir -p $(RELDIR)
	@echo "$(VERSION)" > $(RELDIR)/version.txt
	@for f in $(BINDIR)/$(PROJECTNAME)-*; do \
		[ -f "$$f" ] || continue; \
		strip "$$f" 2>/dev/null || true; \
		cp "$$f" $(RELDIR)/; \
	done
	@tar --exclude='.git' --exclude='.github' --exclude='.gitea' \
		--exclude='binaries' --exclude='releases' --exclude='*.tar.gz' \
		-czf $(RELDIR)/$(PROJECTNAME)-$(VERSION)-source.tar.gz .

docker:
	@docker buildx version > /dev/null 2>&1 || (echo "docker buildx required" && exit 1)
	@docker buildx create --name $(PROJECTNAME)-builder --use 2>/dev/null || docker buildx use $(PROJECTNAME)-builder
	@docker buildx build \
		-f docker/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION="$(VERSION)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg COMMIT_ID="$(COMMIT_ID)" \
		--build-arg OFFICIALSITE="$(OFFICIALSITE)" \
		--build-arg MAINTAINER_NAME="$(MAINTAINER_NAME)" \
		--build-arg MAINTAINER_EMAIL="$(MAINTAINER_EMAIL)" \
		--build-arg PLATFORM_REPO_URL="$(PLATFORM_REPO_URL)" \
		--build-arg LICENSE="$(LICENSE_NAME)" \
		-t $(REGISTRY):$(VERSION) \
		-t $(REGISTRY):latest \
		--push \
		.

test:
	@mkdir -p $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod download
	@$(GO_DOCKER) go test -v -cover -coverprofile=coverage.out ./...
	@COVERAGE=$$($(GO_DOCKER) go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < 100" | bc -l) -eq 1 ]; then \
		echo "ERROR: Coverage is $$COVERAGE%, must be 100%"; \
		exit 1; \
	fi

dev:
	@mkdir -p $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod tidy
	@mkdir -p "$${TMPDIR:-/tmp}/$(PROJECTORG)" && \
		BUILD_DIR=$$(mktemp -d "$${TMPDIR:-/tmp}/$(PROJECTORG)/$(PROJECTNAME)-XXXXXX") && \
		$(GO_DOCKER) go build -o $$BUILD_DIR/$(PROJECTNAME) ./src && \
		echo "Built: $$BUILD_DIR/$(PROJECTNAME)"

clean:
	@rm -rf $(BINDIR) $(RELDIR) coverage.out
