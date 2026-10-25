# Astra: quality gate. `make check` must pass before every commit (docs/go-rules.md).
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

GO           ?= go
TOOLS        := $(GO) tool -modfile=tools/go.mod
# Every module must pass on its own; a local go.work never changes what we check (docs/monorepo.md).
export GOWORK := off
# Pinned in docs/adr/0002-go-toolchain.md. No trailing comment: make would keep the spaces.
GOLANGCI_VER := 2.15.0
# Found on PATH, else in Go's bin directory where the pinned binary is installed.
GOLANGCI     ?= $(shell command -v golangci-lint 2>/dev/null || echo "$$($(GO) env GOPATH)/bin/golangci-lint")
# Only our code: ./... would also pick up Go files inside web/node_modules.
# Listed only when the folder exists, since golangci-lint fails on a missing one.
PKGS         := $(addsuffix /...,$(wildcard ./cmd ./internal))

.PHONY: setup check fmt tidy lint-version lint deadcode test vuln budgets fix integration check-desktop check-tools

# Run once after clone: enables the commit-msg and pre-push hooks (docs/git-rules.md §12.9).
setup:
	git config core.hooksPath .githooks

check: fmt tidy lint-version lint deadcode test vuln budgets
	@echo "make check: OK"

# Tracked and new files; gofmt is never called without arguments, since it would read stdin.
fmt:
	@files="$$(git ls-files --cached --others --exclude-standard -- '*.go')"; \
	if [ -z "$$files" ]; then exit 0; fi; \
	out="$$(gofmt -l $$files)"; \
	if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

tidy:
	$(GO) mod tidy -diff

lint-version:
	@$(GOLANGCI) --version | grep -qF "version $(GOLANGCI_VER) " || \
	  { echo "golangci-lint $(GOLANGCI_VER) required (see docs/adr/0002-go-toolchain.md)"; exit 1; }

lint:
	$(GOLANGCI) run $(PKGS)

deadcode:
	@out="$$($(TOOLS) deadcode -test $(PKGS))"; \
	if [ -n "$$out" ]; then echo "dead code:"; echo "$$out"; exit 1; fi

test:
	$(GO) test -race -count=1 -coverprofile=coverage.out $(PKGS)

vuln:
	$(TOOLS) govulncheck $(PKGS)

budgets:
	./scripts/check-budgets.sh

fix:
	$(GO) fix $(PKGS)

integration:
	$(GO) test -race -count=1 -tags=integration $(PKGS)

# The desktop module arrives with spike S10; until then there is nothing to check.
check-desktop:
	@if [ ! -f desktop/go.mod ]; then echo "check-desktop: no desktop/go.mod yet, skipped"; exit 0; fi; \
	cd desktop && $(GO) mod tidy -diff && $(GO) vet ./... && $(GO) test ./... && $(GOLANGCI) run ./...

check-tools:
	cd tools && $(GO) mod tidy -diff
